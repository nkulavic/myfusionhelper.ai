package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// NewDropboxIt creates a new DropboxIt helper instance
func NewDropboxIt() helpers.Helper { return &DropboxIt{} }

func init() {
	helpers.Register("dropbox_it", func() helpers.Helper { return &DropboxIt{} })
}

// DropboxIt uploads contact data to Dropbox as a text file
type DropboxIt struct{}

func (h *DropboxIt) GetName() string     { return "Dropbox It" }
func (h *DropboxIt) GetType() string     { return "dropbox_it" }
func (h *DropboxIt) GetCategory() string { return "integration" }
func (h *DropboxIt) GetDescription() string {
	return "Upload contact data to Dropbox as a text file"
}
func (h *DropboxIt) RequiresCRM() bool       { return true }
func (h *DropboxIt) SupportedCRMs() []string { return nil }

func (h *DropboxIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Dropbox file path (e.g., /contacts/{first_name}_{last_name}.txt)",
			},
			"content_template": map[string]interface{}{
				"type":        "string",
				"description": "File content template supporting {first_name}, {last_name}, {email}, {phone}, {company} placeholders",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to overwrite existing file",
				"default":     false,
			},
			"apply_tag": map[string]interface{}{
				"type":        "string",
				"description": "Tag ID to apply after successful upload",
			},
			"service_connection_ids": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dropbox": map[string]interface{}{
						"type":        "string",
						"description": "Service connection ID for Dropbox",
					},
				},
			},
		},
		"required": []string{"file_path", "content_template"},
	}
}

func (h *DropboxIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["file_path"].(string); !ok || config["file_path"] == "" {
		return fmt.Errorf("file_path is required")
	}
	if _, ok := config["content_template"].(string); !ok || config["content_template"] == "" {
		return fmt.Errorf("content_template is required")
	}
	return nil
}

func (h *DropboxIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	filePath := input.Config["file_path"].(string)
	contentTemplate := input.Config["content_template"].(string)
	overwrite, _ := input.Config["overwrite"].(bool)
	applyTag, _ := input.Config["apply_tag"].(string)

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Get Dropbox service auth
	auth := input.ServiceAuths["dropbox"]
	if auth == nil {
		output.Message = "Dropbox connection required"
		return output, fmt.Errorf("Dropbox connection required")
	}

	// Get contact data
	contact := input.ContactData
	if contact == nil {
		var err error
		contact, err = input.Connector.GetContact(ctx, input.ContactID)
		if err != nil {
			output.Message = fmt.Sprintf("Failed to get contact data: %v", err)
			return output, err
		}
	}

	// Interpolate file path and content with contact data
	interpolatedPath := interpolateContactData(filePath, contact)
	interpolatedContent := interpolateContactData(contentTemplate, contact)

	output.Logs = append(output.Logs, fmt.Sprintf("Uploading to Dropbox: %s", interpolatedPath))

	// Determine write mode
	writeMode := "add"
	if overwrite {
		writeMode = "overwrite"
	}

	// Upload to Dropbox via API
	uploadURL := "https://content.dropboxapi.com/2/files/upload"

	dropboxArgs := map[string]interface{}{
		"path": interpolatedPath,
		"mode": writeMode,
	}
	dropboxArgsJSON, _ := json.Marshal(dropboxArgs)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader([]byte(interpolatedContent)))
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create upload request: %v", err)
		return output, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(dropboxArgsJSON))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Dropbox upload failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read Dropbox response: %v", err)
		return output, err
	}

	if resp.StatusCode != http.StatusOK {
		output.Message = fmt.Sprintf("Dropbox API returned status %d: %s", resp.StatusCode, string(body))
		return output, fmt.Errorf("Dropbox API error: %d", resp.StatusCode)
	}

	var uploadResult map[string]interface{}
	if err := json.Unmarshal(body, &uploadResult); err != nil {
		output.Message = fmt.Sprintf("Failed to parse Dropbox response: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("File uploaded successfully: %s", interpolatedPath))

	// Apply tag if configured
	if applyTag != "" {
		err := input.Connector.ApplyTag(ctx, input.ContactID, applyTag)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply tag '%s': %v", applyTag, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "tag_applied",
				Target: input.ContactID,
				Value:  applyTag,
			})
			output.Logs = append(output.Logs, fmt.Sprintf("Applied tag '%s'", applyTag))
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("Uploaded contact data to Dropbox: %s", interpolatedPath)
	output.ModifiedData = map[string]interface{}{
		"file_path": interpolatedPath,
		"file_size": len(interpolatedContent),
		"overwrite": overwrite,
	}

	return output, nil
}

// interpolateContactData replaces placeholders with contact data
func interpolateContactData(template string, contact *connectors.NormalizedContact) string {
	result := template
	result = strings.ReplaceAll(result, "{first_name}", contact.FirstName)
	result = strings.ReplaceAll(result, "{last_name}", contact.LastName)
	result = strings.ReplaceAll(result, "{email}", contact.Email)
	result = strings.ReplaceAll(result, "{phone}", contact.Phone)
	result = strings.ReplaceAll(result, "{company}", contact.Company)
	return result
}
