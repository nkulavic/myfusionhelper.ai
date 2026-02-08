package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForGroupIt struct {
	fieldValue       interface{}
	getFieldError    error
	tags             []connectors.Tag
	getTagsError     error
	applyTagError    error
	tagApplied       string
}

func (m *mockConnectorForGroupIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		return nil, m.getFieldError
	}
	return m.fieldValue, nil
}

func (m *mockConnectorForGroupIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	if m.getTagsError != nil {
		return nil, m.getTagsError
	}
	return m.tags, nil
}

func (m *mockConnectorForGroupIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	m.tagApplied = tagID
	return nil
}

func (m *mockConnectorForGroupIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForGroupIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForGroupIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestGroupIt_GetMetadata(t *testing.T) {
	helper := &GroupIt{}
	if helper.GetName() != "Group It" {
		t.Errorf("Expected name 'Group It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "group_it" {
		t.Errorf("Expected type 'group_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "tagging" {
		t.Errorf("Expected category 'tagging', got '%s'", helper.GetCategory())
	}
}

func TestGroupIt_ValidateConfig_Success(t *testing.T) {
	helper := &GroupIt{}
	err := helper.ValidateConfig(map[string]interface{}{"field": "state", "tag_prefix": "Location:"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestGroupIt_ValidateConfig_MissingField(t *testing.T) {
	helper := &GroupIt{}
	err := helper.ValidateConfig(map[string]interface{}{"tag_prefix": "Location:"})
	if err == nil {
		t.Error("Expected error for missing field")
	}
}

func TestGroupIt_ValidateConfig_MissingPrefix(t *testing.T) {
	helper := &GroupIt{}
	err := helper.ValidateConfig(map[string]interface{}{"field": "state"})
	if err == nil {
		t.Error("Expected error for missing tag_prefix")
	}
}

func TestGroupIt_Execute_Success(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{
		fieldValue: "California",
		tags: []connectors.Tag{
			{ID: "tag1", Name: "Location:California"},
			{ID: "tag2", Name: "Location:Texas"},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success")
	}
	if mock.tagApplied != "tag1" {
		t.Errorf("Expected tag1 to be applied, got: %s", mock.tagApplied)
	}
}

func TestGroupIt_Execute_EmptyField(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{fieldValue: ""}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success for empty field")
	}
	if mock.tagApplied != "" {
		t.Error("Expected no tag to be applied for empty field")
	}
}

func TestGroupIt_Execute_NilField(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{fieldValue: nil}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success for nil field")
	}
}

func TestGroupIt_Execute_TagNotFound(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{
		fieldValue: "California",
		tags: []connectors.Tag{
			{ID: "tag1", Name: "Location:Texas"},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if output.Success {
		t.Error("Expected failure when tag not found")
	}
	if mock.tagApplied != "" {
		t.Error("Expected no tag to be applied when not found")
	}
}

func TestGroupIt_Execute_GetFieldError(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{getFieldError: fmt.Errorf("field error")}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetContactFieldValue failure")
	}
}

func TestGroupIt_Execute_GetTagsError(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{
		fieldValue:   "California",
		getTagsError: fmt.Errorf("tags error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetTags failure")
	}
}

func TestGroupIt_Execute_ApplyTagError(t *testing.T) {
	helper := &GroupIt{}
	mock := &mockConnectorForGroupIt{
		fieldValue: "California",
		tags: []connectors.Tag{
			{ID: "tag1", Name: "Location:California"},
		},
		applyTagError: fmt.Errorf("tag error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"field": "state", "tag_prefix": "Location:"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for ApplyTag failure")
	}
}
