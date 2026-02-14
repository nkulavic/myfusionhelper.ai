package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForCountItTags struct {
	contactList      *connectors.ContactList
	getContactsError error
	applyTagError    error
	setFieldError    error
	fieldsSet        map[string]interface{}
	tagsApplied      []string
}

func (m *mockConnectorForCountItTags) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	if m.getContactsError != nil {
		return nil, m.getContactsError
	}
	return m.contactList, nil
}

func (m *mockConnectorForCountItTags) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		return m.applyTagError
	}
	if m.tagsApplied == nil {
		m.tagsApplied = make([]string, 0)
	}
	m.tagsApplied = append(m.tagsApplied, tagID)
	return nil
}

func (m *mockConnectorForCountItTags) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForCountItTags) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForCountItTags) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForCountItTags) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestCountItTags_GetMetadata(t *testing.T) {
	helper := &CountItTags{}
	if helper.GetName() != "Count It Tags" {
		t.Errorf("Expected name 'Count It Tags', got '%s'", helper.GetName())
	}
	if helper.GetType() != "count_it_tags" {
		t.Errorf("Expected type 'count_it_tags', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "tagging" {
		t.Errorf("Expected category 'tagging', got '%s'", helper.GetCategory())
	}
}

func TestCountItTags_ValidateConfig_Success(t *testing.T) {
	helper := &CountItTags{}
	err := helper.ValidateConfig(map[string]interface{}{"tag_id": "tag123"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCountItTags_ValidateConfig_MissingTagID(t *testing.T) {
	helper := &CountItTags{}
	err := helper.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing tag_id")
	}
}

func TestCountItTags_Execute_BasicCount(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{
			Total: 50,
			Contacts: []connectors.NormalizedContact{
				{ID: "123"}, {ID: "456"},
			},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"tag_id": "tag123"},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success")
	}
	if output.ModifiedData["total_count"] != 50 {
		t.Errorf("Expected count 50, got: %v", output.ModifiedData["total_count"])
	}
}

func TestCountItTags_Execute_ThresholdMet(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{Total: 100, Contacts: []connectors.NormalizedContact{}},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":             "tag123",
			"threshold":          50.0,
			"threshold_met_tag":  "met_tag",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "met_tag" {
		t.Error("Expected threshold_met_tag to be applied")
	}
}

func TestCountItTags_Execute_ThresholdNotMet(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{Total: 30, Contacts: []connectors.NormalizedContact{}},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":                 "tag123",
			"threshold":              50.0,
			"threshold_not_met_tag":  "not_met_tag",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "not_met_tag" {
		t.Error("Expected threshold_not_met_tag to be applied")
	}
}

func TestCountItTags_Execute_NoTag(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{Total: 100, Contacts: []connectors.NormalizedContact{}},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":            "tag123",
			"threshold":         50.0,
			"threshold_met_tag": "no_tag",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(mock.tagsApplied) != 0 {
		t.Error("Expected no tag to be applied when set to 'no_tag'")
	}
}

func TestCountItTags_Execute_SaveCountTo(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{Total: 75, Contacts: []connectors.NormalizedContact{}},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":        "tag123",
			"save_count_to": "count_field",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["count_field"] != "75" {
		t.Errorf("Expected count_field to be '75', got: %v", mock.fieldsSet["count_field"])
	}
}

func TestCountItTags_Execute_SavePositionTo(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{
			Total: 5,
			Contacts: []connectors.NormalizedContact{
				{ID: "111"}, {ID: "123"}, {ID: "333"},
			},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":            "tag123",
			"save_position_to":  "position_field",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["position_field"] != "2" {
		t.Errorf("Expected position_field to be '2', got: %v", mock.fieldsSet["position_field"])
	}
	if output.ModifiedData["position"] != 2 {
		t.Error("Expected position to be 2")
	}
}

func TestCountItTags_Execute_ContactNotInList(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList: &connectors.ContactList{
			Total: 3,
			Contacts: []connectors.NormalizedContact{
				{ID: "111"}, {ID: "222"},
			},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "999",
		Config: map[string]interface{}{"tag_id": "tag123"},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if output.ModifiedData["position"] != 0 {
		t.Error("Expected position to be 0 when contact not in list")
	}
}

func TestCountItTags_Execute_GetContactsError(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		getContactsError: fmt.Errorf("query error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{"tag_id": "tag123"},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetContacts failure")
	}
}

func TestCountItTags_Execute_ApplyTagError(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList:   &connectors.ContactList{Total: 100, Contacts: []connectors.NormalizedContact{}},
		applyTagError: fmt.Errorf("tag error"),
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":            "tag123",
			"threshold":         50.0,
			"threshold_met_tag": "met_tag",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal("Expected no error (tag failure should be logged)")
	}
	if !output.Success {
		t.Error("Expected success even with tag failure")
	}
}

func TestCountItTags_Execute_SetFieldError(t *testing.T) {
	helper := &CountItTags{}
	mock := &mockConnectorForCountItTags{
		contactList:   &connectors.ContactList{Total: 50, Contacts: []connectors.NormalizedContact{}},
		setFieldError: fmt.Errorf("field error"),
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tag_id":        "tag123",
			"save_count_to": "count_field",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatal("Expected no error (field failure should be logged)")
	}
	if !output.Success {
		t.Error("Expected success even with field set failure")
	}
}
