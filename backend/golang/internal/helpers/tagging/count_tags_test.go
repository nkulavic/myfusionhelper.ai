package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForCountTags struct {
	contact         *connectors.NormalizedContact
	getContactError error
	tags            []connectors.Tag
	getTagsError    error
	setFieldError   error
	fieldsSet       map[string]interface{}
}

func (m *mockConnectorForCountTags) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForCountTags) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	if m.getTagsError != nil {
		return nil, m.getTagsError
	}
	return m.tags, nil
}

func (m *mockConnectorForCountTags) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForCountTags) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountTags) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForCountTags) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestCountTags_GetMetadata(t *testing.T) {
	helper := &CountTags{}
	if helper.GetName() != "Count Tags" {
		t.Errorf("Expected name 'Count Tags', got '%s'", helper.GetName())
	}
	if helper.GetType() != "count_tags" {
		t.Errorf("Expected type 'count_tags', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "tagging" {
		t.Errorf("Expected category 'tagging', got '%s'", helper.GetCategory())
	}
}

func TestCountTags_ValidateConfig_Success(t *testing.T) {
	helper := &CountTags{}
	err := helper.ValidateConfig(map[string]interface{}{"target_field": "tag_count"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCountTags_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &CountTags{}
	err := helper.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing target_field")
	}
}

func TestCountTags_Execute_CountAllTags(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Tag1"},
				{ID: "tag2", Name: "Tag2"},
				{ID: "tag3", Name: "Tag3"},
			},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count"},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success")
	}
	if mock.fieldsSet["tag_count"] != "3" {
		t.Errorf("Expected tag_count to be '3', got: %v", mock.fieldsSet["tag_count"])
	}
}

func TestCountTags_Execute_NoTags(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{
		contact: &connectors.NormalizedContact{ID: "123", Tags: []connectors.TagRef{}},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["tag_count"] != "0" {
		t.Errorf("Expected tag_count to be '0', got: %v", mock.fieldsSet["tag_count"])
	}
}

func TestCountTags_Execute_WithCategory(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "Tag1"},
				{ID: "tag2", Name: "Tag2"},
				{ID: "tag3", Name: "Tag3"},
			},
		},
		tags: []connectors.Tag{
			{ID: "tag1", Name: "Tag1", Category: "marketing"},
			{ID: "tag2", Name: "Tag2", Category: "sales"},
			{ID: "tag3", Name: "Tag3", Category: "marketing"},
		},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count", "category": "marketing"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["tag_count"] != "2" {
		t.Errorf("Expected tag_count to be '2' (marketing tags), got: %v", mock.fieldsSet["tag_count"])
	}
}

func TestCountTags_Execute_GetContactError(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{getContactError: fmt.Errorf("contact error")}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetContact failure")
	}
}

func TestCountTags_Execute_GetTagsError(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{
		contact: &connectors.NormalizedContact{ID: "123", Tags: []connectors.TagRef{{ID: "tag1"}}},
		getTagsError: fmt.Errorf("tags error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count", "category": "marketing"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetTags failure when category specified")
	}
}

func TestCountTags_Execute_SetFieldError(t *testing.T) {
	helper := &CountTags{}
	mock := &mockConnectorForCountTags{
		contact: &connectors.NormalizedContact{ID: "123", Tags: []connectors.TagRef{{ID: "tag1"}}},
		setFieldError: fmt.Errorf("field error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"target_field": "tag_count"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for SetContactFieldValue failure")
	}
}
