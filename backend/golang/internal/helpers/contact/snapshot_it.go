package contact

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

func init() {
	helpers.Register("snapshot_it", func() helpers.Helper { return &SnapshotIt{} })
}

// SnapshotIt captures a complete snapshot of a contact's data (all fields and tags)
// and produces it as output for downstream storage (queue, database, etc.).
// Ported from legacy PHP snapshot_it helper.
type SnapshotIt struct{}

func (h *SnapshotIt) GetName() string     { return "Snapshot It" }
func (h *SnapshotIt) GetType() string     { return "snapshot_it" }
func (h *SnapshotIt) GetCategory() string { return "contact" }
func (h *SnapshotIt) GetDescription() string {
	return "Capture a full snapshot of a contact's data including all fields and tags for archival"
}
func (h *SnapshotIt) RequiresCRM() bool       { return true }
func (h *SnapshotIt) SupportedCRMs() []string { return nil }

func (h *SnapshotIt) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"include_tags": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include tag data in the snapshot",
				"default":     true,
			},
			"include_custom_fields": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include custom fields in the snapshot",
				"default":     true,
			},
		},
		"required": []string{},
	}
}

func (h *SnapshotIt) ValidateConfig(config map[string]interface{}) error {
	// All config fields are optional
	return nil
}

func (h *SnapshotIt) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	includeTags := true
	if it, ok := input.Config["include_tags"].(bool); ok {
		includeTags = it
	}

	includeCustomFields := true
	if icf, ok := input.Config["include_custom_fields"].(bool); ok {
		includeCustomFields = icf
	}

	output := &helpers.HelperOutput{
		Logs: make([]string, 0),
	}

	// Get contact data
	contact, err := input.Connector.GetContact(ctx, input.ContactID)
	if err != nil {
		output.Message = fmt.Sprintf("Failed to get contact: %v", err)
		return output, err
	}

	// Build snapshot data
	now := time.Now()
	snapshotID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s_%s_%s_%d", input.UserID, input.HelperID, input.ContactID, now.UnixNano()))))

	snapshot := map[string]interface{}{
		"id":                  snapshotID,
		"timestamp":          now.Unix(),
		"snapshot_contact_id": fmt.Sprintf("%s_%s_%s", input.HelperID, input.AccountID, input.ContactID),
		"table_name":         "myfusion_helper_contact_snapshots",
		"account_id":         input.AccountID,
		"helper_id":          input.HelperID,
		"contact_id":         contact.ID,
		"first_name":         contact.FirstName,
		"last_name":          contact.LastName,
		"email":              contact.Email,
		"phone":              contact.Phone,
		"company":            contact.Company,
		"job_title":          contact.JobTitle,
		"full_name":          strings.TrimSpace(contact.FirstName + " " + contact.LastName),
		"source_crm":         contact.SourceCRM,
	}

	if contact.CreatedAt != nil {
		snapshot["created_at"] = contact.CreatedAt.Format(time.RFC3339)
	}
	if contact.UpdatedAt != nil {
		snapshot["updated_at"] = contact.UpdatedAt.Format(time.RFC3339)
	}

	// Include custom fields
	if includeCustomFields && contact.CustomFields != nil {
		for key, value := range contact.CustomFields {
			snapshot[key] = value
		}
	}

	// Include tags
	if includeTags && contact.Tags != nil {
		tagData := make([]map[string]string, 0, len(contact.Tags))
		for _, tag := range contact.Tags {
			tagData = append(tagData, map[string]string{
				"id":   tag.ID,
				"name": tag.Name,
			})
		}
		snapshot["tags"] = tagData
		snapshot["tag_count"] = len(contact.Tags)
	}

	output.Success = true
	output.Message = fmt.Sprintf("Snapshot captured for contact %s", input.ContactID)
	output.Actions = []helpers.HelperAction{
		{
			Type:   "snapshot_captured",
			Target: input.ContactID,
			Value:  snapshotID,
		},
	}
	output.ModifiedData = snapshot
	output.Logs = append(output.Logs, fmt.Sprintf("Snapshot %s captured for contact %s at %s", snapshotID, input.ContactID, now.Format(time.RFC3339)))

	return output, nil
}
