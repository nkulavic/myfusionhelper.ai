package connectors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewKeapConnector tests connector initialization
func TestNewKeapConnector(t *testing.T) {
	t.Run("success with all fields", func(t *testing.T) {
		config := ConnectorConfig{
			AccessToken: "test-token",
			BaseURL:     "https://custom-api.example.com",
		}

		connector, err := NewKeapConnector(config)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if connector == nil {
			t.Fatal("Expected connector to be created")
		}

		keap, ok := connector.(*KeapConnector)
		if !ok {
			t.Fatal("Expected KeapConnector type")
		}

		if keap.accessToken != "test-token" {
			t.Errorf("Expected access token 'test-token', got '%s'", keap.accessToken)
		}

		if keap.baseURL != "https://custom-api.example.com" {
			t.Errorf("Expected custom baseURL, got '%s'", keap.baseURL)
		}
	})

	t.Run("success with default baseURL", func(t *testing.T) {
		config := ConnectorConfig{
			AccessToken: "test-token",
		}

		connector, err := NewKeapConnector(config)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		keap := connector.(*KeapConnector)
		if keap.baseURL != keapBaseURL {
			t.Errorf("Expected default baseURL '%s', got '%s'", keapBaseURL, keap.baseURL)
		}
	})

	t.Run("error when access token missing", func(t *testing.T) {
		config := ConnectorConfig{}

		_, err := NewKeapConnector(config)
		if err == nil {
			t.Fatal("Expected error for missing access token")
		}

		if !strings.Contains(err.Error(), "access token is required") {
			t.Errorf("Expected error about access token, got '%s'", err.Error())
		}
	})
}

// TestKeapConnector_GetContacts tests contact list retrieval
func TestKeapConnector_GetContacts(t *testing.T) {
	t.Run("success with contacts", func(t *testing.T) {
		mockResponse := `{
			"contacts": [
				{
					"id": 123,
					"given_name": "John",
					"family_name": "Doe",
					"email_addresses": [{"email": "john@example.com", "field": "EMAIL1"}],
					"phone_numbers": [{"number": "+1234567890", "field": "PHONE1"}],
					"date_created": "2024-01-01T00:00:00Z",
					"last_updated": "2024-01-15T12:30:00Z",
					"tags": [],
					"custom_fields": []
				},
				{
					"id": 456,
					"given_name": "Jane",
					"family_name": "Smith",
					"email_addresses": [{"email": "jane@example.com", "field": "EMAIL1"}],
					"phone_numbers": [],
					"date_created": "2024-01-02T00:00:00Z",
					"last_updated": "2024-01-16T10:00:00Z",
					"tags": [],
					"custom_fields": []
				}
			],
			"count": 2,
			"next": "https://api.infusionsoft.com/crm/rest/v2/contacts?offset=25"
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			// Verify authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-token" {
				t.Errorf("Expected Bearer auth, got '%s'", authHeader)
			}

			// Verify query parameters
			limit := r.URL.Query().Get("limit")
			if limit != "10" {
				t.Errorf("Expected limit=10, got '%s'", limit)
			}

			// Return mock response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		opts := QueryOptions{Limit: 10}
		result, err := connector.GetContacts(context.Background(), opts)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Contacts) != 2 {
			t.Errorf("Expected 2 contacts, got %d", len(result.Contacts))
		}

		if result.Total != 2 {
			t.Errorf("Expected total=2, got %d", result.Total)
		}

		if !result.HasMore {
			t.Error("Expected hasMore=true")
		}

		// Verify first contact
		firstContact := result.Contacts[0]
		if firstContact.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", firstContact.ID)
		}
		if firstContact.FirstName != "John" {
			t.Errorf("Expected first name 'John', got '%s'", firstContact.FirstName)
		}
		if firstContact.LastName != "Doe" {
			t.Errorf("Expected last name 'Doe', got '%s'", firstContact.LastName)
		}
		if firstContact.Email != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%s'", firstContact.Email)
		}
		if firstContact.Phone != "+1234567890" {
			t.Errorf("Expected phone '+1234567890', got '%s'", firstContact.Phone)
		}
	})

	t.Run("success with empty list", func(t *testing.T) {
		mockResponse := `{"contacts": [], "count": 0}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		result, err := connector.GetContacts(context.Background(), QueryOptions{})

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Contacts) != 0 {
			t.Errorf("Expected 0 contacts, got %d", len(result.Contacts))
		}
	})

	t.Run("error on API failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid token"}`))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "invalid-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		_, err := connector.GetContacts(context.Background(), QueryOptions{})

		if err == nil {
			t.Fatal("Expected error for API failure")
		}

		connErr, ok := err.(*ConnectorError)
		if !ok {
			t.Fatalf("Expected ConnectorError, got %T", err)
		}

		if connErr.StatusCode != 401 {
			t.Errorf("Expected status code 401, got %d", connErr.StatusCode)
		}
	})
}

// TestKeapConnector_GetContact tests single contact retrieval
func TestKeapConnector_GetContact(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockResponse := `{
			"id": 123,
			"given_name": "John",
			"family_name": "Doe",
			"email_addresses": [{"email": "john@example.com", "field": "EMAIL1"}],
			"phone_numbers": [{"number": "+1234567890", "field": "PHONE1"}],
			"date_created": "2024-01-01T00:00:00Z",
			"last_updated": "2024-01-15T12:30:00Z",
			"tag_ids": [{"id": 1, "name": "VIP"}],
			"custom_fields": [{"id": 5, "content": "Premium"}]
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request path
			if !strings.HasSuffix(r.URL.Path, "/contacts/123") {
				t.Errorf("Expected path to end with '/contacts/123', got '%s'", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		contact, err := connector.GetContact(context.Background(), "123")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", contact.ID)
		}

		if len(contact.Tags) != 1 {
			t.Errorf("Expected 1 tag, got %d tags: %+v", len(contact.Tags), contact.Tags)
		} else if contact.Tags[0].Name != "VIP" {
			t.Errorf("Expected VIP tag name, got '%s'", contact.Tags[0].Name)
		}

		if contact.CustomFields["5"] != "Premium" {
			t.Error("Expected Premium custom field value")
		}
	})

	t.Run("error on not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Contact not found"}`))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		_, err := connector.GetContact(context.Background(), "999")

		if err == nil {
			t.Fatal("Expected error for not found")
		}

		connErr := err.(*ConnectorError)
		if connErr.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", connErr.StatusCode)
		}
	})
}

// TestKeapConnector_CreateContact tests contact creation
func TestKeapConnector_CreateContact(t *testing.T) {
	t.Run("success with all fields", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			// Parse request body
			json.NewDecoder(r.Body).Decode(&receivedBody)

			// Return mock created contact
			response := `{
				"id": 789,
				"given_name": "New",
				"family_name": "Contact",
				"email_addresses": [{"email": "new@example.com", "field": "EMAIL1"}],
				"phone_numbers": [{"number": "+1555555555", "field": "PHONE1"}],
				"company": {"company_name": "ACME Corp"},
				"date_created": "2024-02-09T00:00:00Z",
				"last_updated": "2024-02-09T00:00:00Z",
				"tags": [],
				"custom_fields": []
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		input := CreateContactInput{
			FirstName: "New",
			LastName:  "Contact",
			Email:     "new@example.com",
			Phone:     "+1555555555",
			Company:   "ACME Corp",
			CustomFields: map[string]interface{}{
				"10": "Value1",
			},
		}

		contact, err := connector.CreateContact(context.Background(), input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.ID != "789" {
			t.Errorf("Expected ID '789', got '%s'", contact.ID)
		}

		// Verify request body structure
		if receivedBody["given_name"] != "New" {
			t.Error("Expected given_name in request body")
		}
		if receivedBody["family_name"] != "Contact" {
			t.Error("Expected family_name in request body")
		}

		// Verify email_addresses array
		emails, ok := receivedBody["email_addresses"].([]interface{})
		if !ok || len(emails) == 0 {
			t.Error("Expected email_addresses array in request body")
		}

		// Verify phone_numbers array
		phones, ok := receivedBody["phone_numbers"].([]interface{})
		if !ok || len(phones) == 0 {
			t.Error("Expected phone_numbers array in request body")
		}

		// Verify company
		companyData, ok := receivedBody["company"].(map[string]interface{})
		if !ok || companyData["company_name"] != "ACME Corp" {
			t.Error("Expected company in request body")
		}

		// Verify custom_fields
		customFields, ok := receivedBody["custom_fields"].([]interface{})
		if !ok || len(customFields) == 0 {
			t.Error("Expected custom_fields array in request body")
		}
	})

	t.Run("success with minimal fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"id": 800,
				"given_name": "Minimal",
				"family_name": "User",
				"date_created": "2024-02-09T00:00:00Z",
				"last_updated": "2024-02-09T00:00:00Z",
				"tags": [],
				"custom_fields": []
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		input := CreateContactInput{
			FirstName: "Minimal",
			LastName:  "User",
		}

		contact, err := connector.CreateContact(context.Background(), input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.FirstName != "Minimal" {
			t.Errorf("Expected first name 'Minimal', got '%s'", contact.FirstName)
		}
	})
}

// TestKeapConnector_UpdateContact tests contact updates
func TestKeapConnector_UpdateContact(t *testing.T) {
	t.Run("success with partial update", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "PATCH" {
				t.Errorf("Expected PATCH request, got %s", r.Method)
			}

			// Parse request body
			json.NewDecoder(r.Body).Decode(&receivedBody)

			// Return updated contact
			response := `{
				"id": 123,
				"given_name": "UpdatedFirst",
				"family_name": "Doe",
				"email_addresses": [{"email": "john@example.com", "field": "EMAIL1"}],
				"phone_numbers": [{"number": "+1234567890", "field": "PHONE1"}],
				"date_created": "2024-01-01T00:00:00Z",
				"last_updated": "2024-02-09T12:00:00Z",
				"tags": [],
				"custom_fields": []
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		firstName := "UpdatedFirst"
		input := UpdateContactInput{
			FirstName: &firstName,
		}

		contact, err := connector.UpdateContact(context.Background(), "123", input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.FirstName != "UpdatedFirst" {
			t.Errorf("Expected updated first name, got '%s'", contact.FirstName)
		}

		// Verify only updated fields in request body
		if receivedBody["given_name"] != "UpdatedFirst" {
			t.Error("Expected given_name in request body")
		}
	})
}

// TestKeapConnector_DeleteContact tests contact deletion
func TestKeapConnector_DeleteContact(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE request, got %s", r.Method)
			}

			// Verify path
			if !strings.HasSuffix(r.URL.Path, "/contacts/123") {
				t.Errorf("Expected path to end with '/contacts/123', got '%s'", r.URL.Path)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.DeleteContact(context.Background(), "123")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("error on not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.DeleteContact(context.Background(), "999")

		if err == nil {
			t.Fatal("Expected error for not found")
		}
	})
}

// TestKeapConnector_GetMetadata tests metadata retrieval
func TestKeapConnector_GetMetadata(t *testing.T) {
	connector := &KeapConnector{
		baseURL: "https://api.infusionsoft.com/crm/rest/v2",
	}

	metadata := connector.GetMetadata()

	if metadata.PlatformName != "Keap (Infusionsoft)" {
		t.Errorf("Expected platform name 'Keap (Infusionsoft)', got '%s'", metadata.PlatformName)
	}

	if metadata.PlatformSlug != "keap" {
		t.Errorf("Expected platform slug 'keap', got '%s'", metadata.PlatformSlug)
	}

	if metadata.APIVersion != "v2" {
		t.Errorf("Expected API version 'v2', got '%s'", metadata.APIVersion)
	}

	if metadata.BaseURL != "https://api.infusionsoft.com/crm/rest/v2" {
		t.Errorf("Expected BaseURL to match connector baseURL, got '%s'", metadata.BaseURL)
	}
}

// TestKeapConnector_TestConnection tests connection verification
func TestKeapConnector_TestConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify it hits the userinfo endpoint
			if !strings.Contains(r.URL.Path, "/oauth/connect/userinfo") {
				t.Errorf("Expected /oauth/connect/userinfo path, got '%s'", r.URL.Path)
			}

			// Return mock user info
			response := `{"sub": "user123", "email": "test@example.com"}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.TestConnection(context.Background())

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("error on auth failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		connector := &KeapConnector{
			accessToken: "invalid-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.TestConnection(context.Background())

		if err == nil {
			t.Fatal("Expected error for auth failure")
		}
	})
}

// TestKeapContact_toNormalized tests contact normalization
func TestKeapContact_toNormalized(t *testing.T) {
	kc := keapContact{
		ID:        123,
		GivenName: "John",
		FamilyName: "Doe",
		Emails: []struct {
			Email string `json:"email"`
			Field string `json:"field"`
		}{
			{Email: "john@example.com", Field: "EMAIL1"},
		},
		Phones: []struct {
			Number string `json:"number"`
			Field  string `json:"field"`
		}{
			{Number: "+1234567890", Field: "PHONE1"},
		},
		Tags: []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{
			{ID: 1, Name: "VIP"},
		},
		CustomFields: []struct {
			ID      int         `json:"id"`
			Content interface{} `json:"content"`
		}{
			{ID: 5, Content: "Premium"},
		},
		DateCreated: "2024-01-01T00:00:00Z",
		LastUpdated: "2024-01-15T12:30:00Z",
	}

	normalized := kc.toNormalized()

	if normalized.ID != "123" {
		t.Errorf("Expected ID '123', got '%s'", normalized.ID)
	}
	if normalized.FirstName != "John" {
		t.Errorf("Expected first name 'John', got '%s'", normalized.FirstName)
	}
	if normalized.LastName != "Doe" {
		t.Errorf("Expected last name 'Doe', got '%s'", normalized.LastName)
	}
	if normalized.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", normalized.Email)
	}
	if normalized.Phone != "+1234567890" {
		t.Errorf("Expected phone '+1234567890', got '%s'", normalized.Phone)
	}
	if len(normalized.Tags) != 1 || normalized.Tags[0].Name != "VIP" {
		t.Error("Expected VIP tag")
	}
	if normalized.CustomFields["5"] != "Premium" {
		t.Error("Expected Premium custom field")
	}
	if normalized.CreatedAt == nil {
		t.Error("Expected CreatedAt timestamp")
	}
	if normalized.UpdatedAt == nil {
		t.Error("Expected UpdatedAt timestamp")
	}
}
