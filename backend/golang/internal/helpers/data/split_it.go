package data

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewSplitIt creates a new SplitIt helper instance
func NewSplitIt() helpers.Helper { return &SplitIt{} }

func init() {
	helpers.Register("split_it", func() helpers.Helper { return &SplitIt{} })
}

// SplitIt performs A/B or N-way split testing by alternating between options
type SplitIt struct{}

func (h *SplitIt) GetName() string { return "Split It" }
func (h *SplitIt) GetType() string { return "split_it" }
func (h *SplitIt) GetCategory() string { return "data" }
func (h *SplitIt) GetDescription() string {
	return "A/B or N-way split testing - alternate between options on each execution (state_field or DynamoDB counter mode)"
}
func (h *SplitIt) RequiresCRM() bool       { return true }
func (h *SplitIt) SupportedCRMs() []string { return nil }

func (h *SplitIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"split_method": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"state_field", "counter"},
				"description": "Method for tracking splits: 'state_field' (contact-level alternation) or 'counter' (atomic DynamoDB counter)",
				"default":     "state_field",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"tag", "goal"},
				"description": "Whether to apply tags or achieve goals for the split",
			},
			"split_count": map[string]interface{}{
				"type":        "number",
				"description": "Number of split options (2 for A/B, 3 for A/B/C, etc.). Default: 2",
				"default":     2,
			},
			"option_a": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option A",
			},
			"option_b": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option B",
			},
			"option_c": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option C (optional, for 3+ way splits)",
			},
			"option_d": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option D (optional, for 4+ way splits)",
			},
			"option_e": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID or goal name for option E (optional, for 5 way splits)",
			},
			"state_field": map[string]interface{}{
				"type":        "string",
				"description": "Field to store the last choice (A/B/C/D/E) for alternation (required if split_method is 'state_field')",
			},
		},
		"required": []string{"mode", "option_a", "option_b"},
	}
}

func (h *SplitIt) ValidateConfig(config map[string]interface{}) error {
	mode, ok := config["mode"].(string)
	if !ok || mode == "" {
		return fmt.Errorf("mode is required")
	}
	if mode != "tag" && mode != "goal" {
		return fmt.Errorf("mode must be 'tag' or 'goal'")
	}

	splitMethod, _ := config["split_method"].(string)
	if splitMethod == "" {
		splitMethod = "state_field"
	}
	if splitMethod != "state_field" && splitMethod != "counter" {
		return fmt.Errorf("split_method must be 'state_field' or 'counter'")
	}

	// state_field required for state_field method
	if splitMethod == "state_field" {
		if _, ok := config["state_field"].(string); !ok || config["state_field"] == "" {
			return fmt.Errorf("state_field is required when split_method is 'state_field'")
		}
	}

	if _, ok := config["option_a"].(string); !ok || config["option_a"] == "" {
		return fmt.Errorf("option_a is required")
	}
	if _, ok := config["option_b"].(string); !ok || config["option_b"] == "" {
		return fmt.Errorf("option_b is required")
	}

	// Validate split_count and options match
	splitCount := 2
	if sc, ok := config["split_count"].(float64); ok && sc > 0 {
		splitCount = int(sc)
	}

	if splitCount < 2 || splitCount > 5 {
		return fmt.Errorf("split_count must be between 2 and 5")
	}

	// Check that required options exist
	requiredOptions := []string{"option_a", "option_b"}
	if splitCount >= 3 {
		requiredOptions = append(requiredOptions, "option_c")
	}
	if splitCount >= 4 {
		requiredOptions = append(requiredOptions, "option_d")
	}
	if splitCount >= 5 {
		requiredOptions = append(requiredOptions, "option_e")
	}

	for _, opt := range requiredOptions {
		if _, ok := config[opt].(string); !ok || config[opt] == "" {
			return fmt.Errorf("%s is required for split_count=%d", opt, splitCount)
		}
	}

	return nil
}

func (h *SplitIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	mode := input.Config["mode"].(string)
	splitMethod, _ := input.Config["split_method"].(string)
	if splitMethod == "" {
		splitMethod = "state_field"
	}

	splitCount := 2
	if sc, ok := input.Config["split_count"].(float64); ok && sc > 0 {
		splitCount = int(sc)
	}

	// Collect options
	options := []string{
		input.Config["option_a"].(string),
		input.Config["option_b"].(string),
	}
	if splitCount >= 3 {
		options = append(options, input.Config["option_c"].(string))
	}
	if splitCount >= 4 {
		options = append(options, input.Config["option_d"].(string))
	}
	if splitCount >= 5 {
		options = append(options, input.Config["option_e"].(string))
	}

	output := &helpers.HelperOutput{
		Actions:      make([]helpers.HelperAction, 0),
		ModifiedData: make(map[string]interface{}),
		Logs:         make([]string, 0),
	}

	var currentChoice string
	var currentOption string
	var choiceIndex int

	// Determine which option to use based on split method
	switch splitMethod {
	case "state_field":
		choiceIndex, currentChoice, currentOption = h.executeStateFieldMethod(ctx, input, options, splitCount)

	case "counter":
		var err error
		choiceIndex, currentChoice, currentOption, err = h.executeCounterMethod(ctx, input, options, splitCount)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get counter value: %v", err)
			return output, err
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Split method: %s, count: %d, selected: %s (index %d)", splitMethod, splitCount, currentChoice, choiceIndex))

	// Apply the selected option
	switch mode {
	case "tag":
		err := input.Connector.ApplyTag(ctx, input.ContactID, currentOption)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to apply split tag %s: %v", currentOption, err)
			return output, err
		}
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "tag_applied",
			Target: input.ContactID,
			Value:  currentOption,
		})

	case "goal":
		err := input.Connector.AchieveGoal(ctx, input.ContactID, currentOption, "mfh")
		if err != nil {
			output.Message = fmt.Sprintf("Failed to achieve split goal '%s': %v", currentOption, err)
			return output, err
		}
		output.Actions = append(output.Actions, helpers.HelperAction{
			Type:   "goal_achieved",
			Target: input.ContactID,
			Value:  currentOption,
		})
	}

	// Update state field if using state_field method
	if splitMethod == "state_field" {
		stateField := input.Config["state_field"].(string)
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, stateField, currentChoice)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: failed to update state field '%s': %v", stateField, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: stateField,
				Value:  currentChoice,
			})
			output.ModifiedData[stateField] = currentChoice
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Split test: selected option %s (%s mode, %s method)", currentChoice, mode, splitMethod)
	output.Logs = append(output.Logs, fmt.Sprintf("%d-way split on contact %s: chose %s (option: %s, mode: %s)", splitCount, input.ContactID, currentChoice, currentOption, mode))

	return output, nil
}

// executeStateFieldMethod uses contact field to track last choice and alternate
func (h *SplitIt) executeStateFieldMethod(ctx context.Context, input helpers.HelperInput, options []string, splitCount int) (int, string, string) {
	stateField := input.Config["state_field"].(string)
	choiceLabels := []string{"A", "B", "C", "D", "E"}

	// Read current state
	lastChoice := ""
	stateValue, err := input.Connector.GetContactFieldValue(ctx, input.ContactID, stateField)
	if err == nil && stateValue != nil {
		strVal := fmt.Sprintf("%v", stateValue)
		if strVal != "<nil>" {
			lastChoice = strVal
		}
	}

	// Find last choice index
	lastIndex := -1
	for i, label := range choiceLabels[:splitCount] {
		if lastChoice == label {
			lastIndex = i
			break
		}
	}

	// Increment to next option (wrap around)
	currentIndex := (lastIndex + 1) % splitCount
	currentChoice := choiceLabels[currentIndex]
	currentOption := options[currentIndex]

	return currentIndex, currentChoice, currentOption
}

// executeCounterMethod uses DynamoDB atomic counter to determine split
func (h *SplitIt) executeCounterMethod(ctx context.Context, input helpers.HelperInput, options []string, splitCount int) (int, string, string, error) {
	choiceLabels := []string{"A", "B", "C", "D", "E"}

	// Load DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("RATE_LIMITS_TABLE")
	if tableName == "" {
		return 0, "", "", fmt.Errorf("RATE_LIMITS_TABLE environment variable not set")
	}

	// Counter key: split:{helper_id}
	counterKey := fmt.Sprintf("split:%s", input.HelperID)

	// Atomic increment
	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]ddbtypes.AttributeValue{
			"rate_key": &ddbtypes.AttributeValueMemberS{Value: counterKey},
		},
		UpdateExpression: aws.String("ADD #count :inc"),
		ExpressionAttributeNames: map[string]string{
			"#count": "count",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":inc": &ddbtypes.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: ddbtypes.ReturnValueAllNew,
	})
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to increment counter: %w", err)
	}

	// Extract counter value
	counterValue := int64(0)
	if countAttr, ok := result.Attributes["count"].(*ddbtypes.AttributeValueMemberN); ok {
		counterValue, _ = strconv.ParseInt(countAttr.Value, 10, 64)
	}

	// Modulo to determine split choice
	currentIndex := int(counterValue % int64(splitCount))
	currentChoice := choiceLabels[currentIndex]
	currentOption := options[currentIndex]

	return currentIndex, currentChoice, currentOption, nil
}
