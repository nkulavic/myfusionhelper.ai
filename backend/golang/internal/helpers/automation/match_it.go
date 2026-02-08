package automation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("match_it", func() helpers.Helper { return &MatchIt{} })
}

// MatchIt performs pattern matching on contact field values and triggers actions
type MatchIt struct{}

func (h *MatchIt) GetName() string     { return "Match It" }
func (h *MatchIt) GetType() string     { return "match_it" }
func (h *MatchIt) GetCategory() string { return "automation" }
func (h *MatchIt) GetDescription() string {
	return "Match contact field values against patterns (contains, regex, equals) and trigger conditional actions"
}
func (h *MatchIt) RequiresCRM() bool       { return true }
func (h *MatchIt) SupportedCRMs() []string { return nil }

func (h *MatchIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_field": map[string]interface{}{
				"type":        "string",
				"description": "Contact field to match against",
			},
			"match_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"equals", "contains", "starts_with", "ends_with", "regex"},
				"description": "Type of pattern matching",
				"default":     "contains",
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Pattern to match (string or regex depending on match_type)",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether matching should be case-sensitive",
				"default":     false,
			},
			"on_match_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if pattern matches",
			},
			"on_no_match_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply if pattern doesn't match",
			},
			"save_result_to": map[string]interface{}{
				"type":        "string",
				"description": "Optional field to save match result (true/false)",
			},
		},
		"required": []string{"source_field", "pattern"},
	}
}

func (h *MatchIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["source_field"].(string); !ok || config["source_field"] == "" {
		return fmt.Errorf("source_field is required")
	}

	if _, ok := config["pattern"].(string); !ok || config["pattern"] == "" {
		return fmt.Errorf("pattern is required")
	}

	matchType, _ := config["match_type"].(string)
	if matchType == "regex" {
		pattern := config["pattern"].(string)
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	return nil
}

func (h *MatchIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	sourceField := input.Config["source_field"].(string)
	pattern := input.Config["pattern"].(string)
	matchType, _ := input.Config["match_type"].(string)
	if matchType == "" {
		matchType = "contains"
	}
	caseSensitive, _ := input.Config["case_sensitive"].(bool)
	onMatchTag, _ := input.Config["on_match_tag"].(string)
	onNoMatchTag, _ := input.Config["on_no_match_tag"].(string)
	saveResultTo, _ := input.Config["save_result_to"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get source field value
	sourceValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, sourceField)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get source field '%s': %v", sourceField, err)
		return output, err
	}

	sourceText := fmt.Sprintf("%v", sourceValue)
	output.Logs = append(output.Logs, fmt.Sprintf("Source text: %s = '%s'", sourceField, sourceText))

	// Perform matching
	matched := false
	var matchErr error

	switch matchType {
	case "equals":
		if caseSensitive {
			matched = sourceText == pattern
		} else {
			matched = strings.EqualFold(sourceText, pattern)
		}

	case "contains":
		if caseSensitive {
			matched = strings.Contains(sourceText, pattern)
		} else {
			matched = strings.Contains(strings.ToLower(sourceText), strings.ToLower(pattern))
		}

	case "starts_with":
		if caseSensitive {
			matched = strings.HasPrefix(sourceText, pattern)
		} else {
			matched = strings.HasPrefix(strings.ToLower(sourceText), strings.ToLower(pattern))
		}

	case "ends_with":
		if caseSensitive {
			matched = strings.HasSuffix(sourceText, pattern)
		} else {
			matched = strings.HasSuffix(strings.ToLower(sourceText), strings.ToLower(pattern))
		}

	case "regex":
		re, err := regexp.Compile(pattern)
		if err != nil {
			matchErr = fmt.Errorf("failed to compile regex: %w", err)
			matched = false
		} else {
			matched = re.MatchString(sourceText)
		}

	default:
		matchErr = fmt.Errorf("unknown match_type: %s", matchType)
	}

	if matchErr != nil {
		output.Message = fmt.Sprintf("Match error: %v", matchErr)
		return output, matchErr
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Match result: %t (type: %s, pattern: '%s', case_sensitive: %t)", matched, matchType, pattern, caseSensitive))

	// Save result to field if configured
	if saveResultTo != "" {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, saveResultTo, matched)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to save match result to field '%s': %v", saveResultTo, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: saveResultTo,
				Value:  matched,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Saved match result to field '%s'", saveResultTo))
		}
	}

	// Apply tags based on match result
	if matched && onMatchTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, onMatchTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply match tag '%s': %v", onMatchTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  onMatchTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' (matched)", onMatchTag))
		}
	}

	if !matched && onNoMatchTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, onNoMatchTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply no-match tag '%s': %v", onNoMatchTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  onNoMatchTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s' (no match)", onNoMatchTag))
		}
	}

	output.Success = true
	if matched {
		output.Message = fmt.Sprintf("Pattern matched: '%s' %s '%s'", sourceText, matchType, pattern)
	} else {
		output.Message = fmt.Sprintf("Pattern did not match: '%s' %s '%s'", sourceText, matchType, pattern)
	}

	output.ModifiedData = map[string]interface{}{
		"matched":        matched,
		"match_type":     matchType,
		"pattern":        pattern,
		"source_field":   sourceField,
		"source_value":   sourceText,
		"case_sensitive": caseSensitive,
	}

	return output, nil
}
