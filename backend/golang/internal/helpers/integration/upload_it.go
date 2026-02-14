package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewUploadIt creates a new UploadIt helper instance
func NewUploadIt() helpers.Helper { return &UploadIt{} }

func init() {
	helpers.Register("upload_it", func() helpers.Helper { return &UploadIt{} })
}

// UploadIt uploads files to external services (S3, Dropbox, etc)
type UploadIt struct{}

func (h *UploadIt) GetName() string     { return "Upload It" }
func (h *UploadIt) GetType() string     { return "upload_it" }
func (h *UploadIt) GetCategory() string { return "integration" }
func (h *UploadIt) GetDescription() string {
	return "Upload files to external services or webhooks via HTTP POST"
}
func (h *UploadIt) RequiresCRM() bool       { return false }
func (h *UploadIt) SupportedCRMs() []string { return nil }

func (h *UploadIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"upload_url": map[string]interface{}{
				"type":        "string",
				"description": "Webhook URL to upload the file to",
			},
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Local file path to upload",
			},
			"file_field_name": map[string]interface{}{
				"type":        "string",
				"description": "Form field name for the file upload",
				"default":     "file",
			},
			"additional_fields": map[string]interface{}{
				"type":        "object",
				"description": "Additional form fields to include in the upload",
			},
			"http_method": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"POST", "PUT"},
				"description": "HTTP method for upload",
				"default":     "POST",
			},
			"auth_header": map[string]interface{}{
				"type":        "string",
				"description": "Optional Authorization header value",
			},
		},
		"required": []string{"upload_url", "file_path"},
	}
}

func (h *UploadIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["upload_url"].(string); !ok || config["upload_url"] == "" {
		return fmt.Errorf("upload_url is required")
	}
	if _, ok := config["file_path"].(string); !ok || config["file_path"] == "" {
		return fmt.Errorf("file_path is required")
	}
	return nil
}

func (h *UploadIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	uploadURL := input.Config["upload_url"].(string)
	filePath := input.Config["file_path"].(string)
	fileFieldName, _ := input.Config["file_field_name"].(string)
	if fileFieldName == "" {
		fileFieldName = "file"
	}
	httpMethod, _ := input.Config["http_method"].(string)
	if httpMethod == "" {
		httpMethod = "POST"
	}
	authHeader, _ := input.Config["auth_header"].(string)

	additionalFields := make(map[string]string)
	if fields, ok := input.Config["additional_fields"].(map[string]interface{}); ok {
		for k, v := range fields {
			additionalFields[k] = fmt.Sprintf("%v", v)
		}
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Uploading file: %s to %s", filePath, uploadURL))

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to open file: %v", err)
		return output, err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		output.Message = fmt.Sprintf("Failed to stat file: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("File size: %d bytes", fileInfo.Size()))

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile(fileFieldName, fileInfo.Name())
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create form file: %v", err)
		return output, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to copy file to form: %v", err)
		return output, err
	}

	// Add additional fields
	for key, value := range additionalFields {
		err = writer.WriteField(key, value)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Warning: Failed to add field '%s': %v", key, err))
		}
	}

	err = writer.Close()
	if err != nil {
		output.Message = fmt.Sprintf("Failed to close multipart writer: %v", err)
		return output, err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, httpMethod, uploadURL, body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create upload request: %v", err)
		return output, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Execute upload
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("Upload failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to read upload response: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Upload response status: %d", resp.StatusCode))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		output.Message = fmt.Sprintf("Upload failed with status %d: %s", resp.StatusCode, string(respBody))
		return output, fmt.Errorf("upload failed: %d", resp.StatusCode)
	}

	// Try to parse response as JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err == nil {
		output.Logs = append(output.Logs, "Upload response is valid JSON")
	} else {
		responseData = map[string]interface{}{
			"raw_response": string(respBody),
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("File uploaded successfully: %s", fileInfo.Name())
	output.ModifiedData = map[string]interface{}{
		"upload_url":      uploadURL,
		"file_path":       filePath,
		"file_name":       fileInfo.Name(),
		"file_size":       fileInfo.Size(),
		"http_status":     resp.StatusCode,
		"response":        responseData,
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "file_uploaded",
		Target: uploadURL,
		Value:  fileInfo.Name(),
	})

	return output, nil
}
