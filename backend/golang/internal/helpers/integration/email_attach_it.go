package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewEmailAttachIt creates a new EmailAttachIt helper instance
func NewEmailAttachIt() helpers.Helper { return &EmailAttachIt{} }

func init() {
	helpers.Register("email_attach_it", func() helpers.Helper { return &EmailAttachIt{} })
}

// EmailAttachIt downloads a file from a URL and attaches it to a CRM email
type EmailAttachIt struct{}

func (h *EmailAttachIt) GetName() string     { return "Email Attach It" }
func (h *EmailAttachIt) GetType() string     { return "email_attach_it" }
func (h *EmailAttachIt) GetCategory() string { return "integration" }
func (h *EmailAttachIt) GetDescription() string {
	return "Download a file from URL and prepare it as an email attachment for CRM"
}
func (h *EmailAttachIt) RequiresCRM() bool       { return false }
func (h *EmailAttachIt) SupportedCRMs() []string { return nil }

func (h *EmailAttachIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_url": map[string]interface{}{
				"type":        "string",
				"description": "URL of the file to download and attach",
			},
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Optional custom filename for the attachment",
			},
			"max_size_mb": map[string]interface{}{
				"type":        "number",
				"description": "Maximum file size in MB (default: 10)",
				"default":     10,
			},
		},
		"required": []string{"file_url"},
	}
}

func (h *EmailAttachIt) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["file_url"].(string); !ok || config["file_url"] == "" {
		return fmt.Errorf("file_url is required")
	}
	return nil
}

func (h *EmailAttachIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	fileURL := input.Config["file_url"].(string)
	customFilename, _ := input.Config["filename"].(string)
	maxSizeMB := 10.0
	if ms, ok := input.Config["max_size_mb"].(float64); ok && ms > 0 {
		maxSizeMB = ms
	}

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	output.Logs = append(output.Logs, fmt.Sprintf("Downloading file from: %s", fileURL))

	// Download file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create download request: %v", err)
		return output, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Message = fmt.Sprintf("File download failed: %v", err)
		return output, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		output.Message = fmt.Sprintf("Download failed with status %d", resp.StatusCode)
		return output, fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	// Check file size
	contentLength := resp.ContentLength
	if contentLength > 0 {
		sizeMB := float64(contentLength) / (1024 * 1024)
		if sizeMB > maxSizeMB {
			output.Message = fmt.Sprintf("File too large: %.2f MB (max: %.2f MB)", sizeMB, maxSizeMB)
			return output, fmt.Errorf("file too large")
		}
		output.Logs = append(output.Logs, fmt.Sprintf("File size: %.2f MB", sizeMB))
	}

	// Determine filename
	filename := customFilename
	if filename == "" {
		// Extract from URL
		filename = filepath.Base(fileURL)
		if filename == "." || filename == "/" {
			filename = "attachment"
		}
	}

	// Create temp directory
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, filename)

	// Save file
	file, err := os.Create(tempFile)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to create temp file: %v", err)
		return output, err
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to save file: %v", err)
		return output, err
	}

	output.Logs = append(output.Logs, fmt.Sprintf("File saved to: %s (size: %d bytes)", tempFile, written))

	output.Success = true
	output.Message = fmt.Sprintf("File downloaded and ready for attachment: %s", filename)
	output.ModifiedData = map[string]interface{}{
		"file_url":   fileURL,
		"filename":   filename,
		"temp_path":  tempFile,
		"file_size":  written,
		"size_mb":    float64(written) / (1024 * 1024),
	}

	output.Actions = append(output.Actions, helpers.HelperAction{
		Type:   "file_downloaded",
		Target: filename,
		Value:  tempFile,
	})

	return output, nil
}
