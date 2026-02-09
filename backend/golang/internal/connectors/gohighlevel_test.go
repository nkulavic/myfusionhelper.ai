package connectors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewGoHighLevelConnector tests connector initialization
func TestNewGoHighLevelConnector(t *testing.T) {
	t.Run("success with all fields", func(t *testing.T) {
		config := ConnectorConfig{
			AccessToken: "test-token",
			BaseURL:     "https://custom-api.example.com",
			AccountID:   "location-123",
		}

		connector, err := NewGoHighLevelConnector(config)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if connector == nil {
			t.Fatal("Expected connector to be created")
		}

		ghl, ok := connector.(*GoHighLevelConnector)
		if !ok {
			t.Fatal("Expected GoHighLevelConnector type")
		}

		if ghl.accessToken != "test-token" {
			t.Errorf("Expected access token 'test-token', got '%s'", ghl.accessToken)
		}

		if ghl.baseURL != "https://custom-api.example.com" {
			t.Errorf("Expected custom baseURL, got '%s'", ghl.baseURL)
		}

		if ghl.locationID != "location-123" {
			t.Errorf("Expected locationID 'location-123', got '%s'", ghl.locationID)
		}
	})

	t.Run("success with default baseURL", func(t *testing.T) {
		config := ConnectorConfig{
			AccessToken: "test-token",
		}

		connector, err := NewGoHighLevelConnector(config)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		ghl := connector.(*GoHighLevelConnector)
		if ghl.baseURL != ghlBaseURL {
			t.Errorf("Expected default baseURL '%s', got '%s'", ghlBaseURL, ghl.baseURL)
		}
	})

	t.Run("error when access token missing", func(t *testing.T) {
		config := ConnectorConfig{}

		_, err := NewGoHighLevelConnector(config)
		if err == nil {
			t.Fatal("Expected error for missing access token")
		}

		if !strings.Contains(err.Error(), "access token is required") {
			t.Errorf("Expected error about access token, got '%s'", err.Error())
		}
	})
}

// TestGoHighLevelConnector_GetContacts tests contact list retrieval
func TestGoHighLevelConnector_GetContacts(t *testing.T) {
	t.Run("success with contacts", func(t *testing.T) {
		mockResponse := `{
			"contacts": [
				{
					"id": "contact-123",
					"firstName": "John",
					"lastName": "Doe",
					"email": "john@example.com",
					"phone": "+1234567890",
					"companyName": "ACME Corp",
					"dateAdded": "2024-01-01T00:00:00.000Z",
					"dateUpdated": "2024-01-15T12:30:00.000Z",
					"tags": ["VIP", "Customer"],
					"customFields": [{"id": "field1", "value": "Premium"}]
				},
				{
					"id": "contact-456",
					"firstName": "Jane",
					"lastName": "Smith",
					"email": "jane@example.com",
					"phone": "+1555555555",
					"dateAdded": "2024-01-02T00:00:00.000Z",
					"dateUpdated": "2024-01-16T10:00:00.000Z",
					"tags": [],
					"customFields": []
				}
			],
			"meta": {
				"total": 2,
				"nextPageUrl": "https://api.example.com/contacts?startAfterId=contact-456",
				"startAfterId": "contact-456"
			}
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

			// Verify API version header
			versionHeader := r.Header.Get("Version")
			if versionHeader != ghlAPIVersion {
				t.Errorf("Expected Version header '%s', got '%s'", ghlAPIVersion, versionHeader)
			}

			// Verify query parameters
			limit := r.URL.Query().Get("limit")
			if limit != "10" {
				t.Errorf("Expected limit=10, got '%s'", limit)
			}

			locationID := r.URL.Query().Get("locationId")
			if locationID != "location-123" {
				t.Errorf("Expected locationId='location-123', got '%s'", locationID)
			}

			// Return mock response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			locationID:  "location-123",
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

		if result.NextCursor != "contact-456" {
			t.Errorf("Expected nextCursor='contact-456', got '%s'", result.NextCursor)
		}

		// Verify first contact
		firstContact := result.Contacts[0]
		if firstContact.ID != "contact-123" {
			t.Errorf("Expected ID 'contact-123', got '%s'", firstContact.ID)
		}
		if firstContact.FirstName != "John" {
			t.Errorf("Expected first name 'John', got '%s'", firstContact.FirstName)
		}
		if firstContact.Email != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%s'", firstContact.Email)
		}
		if firstContact.Company != "ACME Corp" {
			t.Errorf("Expected company 'ACME Corp', got '%s'", firstContact.Company)
		}
		if len(firstContact.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(firstContact.Tags))
		}
	})

	t.Run("success with empty list", func(t *testing.T) {
		mockResponse := `{
			"contacts": [],
			"meta": {
				"total": 0,
				"nextPageUrl": "",
				"startAfterId": ""
			}
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
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

		if result.HasMore {
			t.Error("Expected hasMore=false")
		}
	})

	t.Run("error on API failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid token"}`))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
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

// TestGoHighLevelConnector_GetContact tests single contact retrieval
func TestGoHighLevelConnector_GetContact(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockResponse := `{
			"contact": {
				"id": "contact-123",
				"firstName": "John",
				"lastName": "Doe",
				"email": "john@example.com",
				"phone": "+1234567890",
				"companyName": "ACME Corp",
				"dateAdded": "2024-01-01T00:00:00.000Z",
				"dateUpdated": "2024-01-15T12:30:00.000Z",
				"tags": ["VIP"],
				"customFields": [{"id": "field1", "value": "Premium"}]
			}
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request path
			if !strings.HasSuffix(r.URL.Path, "/contacts/contact-123") {
				t.Errorf("Expected path to end with '/contacts/contact-123', got '%s'", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		contact, err := connector.GetContact(context.Background(), "contact-123")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.ID != "contact-123" {
			t.Errorf("Expected ID 'contact-123', got '%s'", contact.ID)
		}

		if len(contact.Tags) != 1 || contact.Tags[0].Name != "VIP" {
			t.Error("Expected VIP tag")
		}

		if contact.CustomFields["field1"] != "Premium" {
			t.Error("Expected Premium custom field value")
		}
	})

	t.Run("error on not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Contact not found"}`))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		_, err := connector.GetContact(context.Background(), "contact-999")

		if err == nil {
			t.Fatal("Expected error for not found")
		}

		connErr := err.(*ConnectorError)
		if connErr.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", connErr.StatusCode)
		}
	})
}

// TestGoHighLevelConnector_CreateContact tests contact creation
func TestGoHighLevelConnector_CreateContact(t *testing.T) {
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
				"contact": {
					"id": "contact-new",
					"firstName": "New",
					"lastName": "Contact",
					"email": "new@example.com",
					"phone": "+1555555555",
					"companyName": "ACME Corp",
					"dateAdded": "2024-02-09T00:00:00.000Z",
					"dateUpdated": "2024-02-09T00:00:00.000Z",
					"tags": ["New"],
					"customFields": []
				}
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			locationID:  "location-123",
			client:      server.Client(),
		}

		input := CreateContactInput{
			FirstName: "New",
			LastName:  "Contact",
			Email:     "new@example.com",
			Phone:     "+1555555555",
			Company:   "ACME Corp",
			Tags:      []string{"New"},
			CustomFields: map[string]interface{}{
				"field1": "Value1",
			},
		}

		contact, err := connector.CreateContact(context.Background(), input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.ID != "contact-new" {
			t.Errorf("Expected ID 'contact-new', got '%s'", contact.ID)
		}

		// Verify request body structure
		if receivedBody["firstName"] != "New" {
			t.Error("Expected firstName in request body")
		}
		if receivedBody["lastName"] != "Contact" {
			t.Error("Expected lastName in request body")
		}
		if receivedBody["email"] != "new@example.com" {
			t.Error("Expected email in request body")
		}
		if receivedBody["phone"] != "+1555555555" {
			t.Error("Expected phone in request body")
		}
		if receivedBody["companyName"] != "ACME Corp" {
			t.Error("Expected companyName in request body")
		}
		if receivedBody["locationId"] != "location-123" {
			t.Error("Expected locationId in request body")
		}

		// Verify tags
		tags, ok := receivedBody["tags"].([]interface{})
		if !ok || len(tags) == 0 {
			t.Error("Expected tags array in request body")
		}

		// Verify custom fields
		customFields, ok := receivedBody["customFields"].([]interface{})
		if !ok || len(customFields) == 0 {
			t.Error("Expected customFields array in request body")
		}
	})

	t.Run("success with minimal fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"contact": {
					"id": "contact-minimal",
					"firstName": "Minimal",
					"lastName": "User",
					"email": "minimal@example.com",
					"dateAdded": "2024-02-09T00:00:00.000Z",
					"dateUpdated": "2024-02-09T00:00:00.000Z",
					"tags": [],
					"customFields": []
				}
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		input := CreateContactInput{
			FirstName: "Minimal",
			LastName:  "User",
			Email:     "minimal@example.com",
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

// TestGoHighLevelConnector_UpdateContact tests contact updates
func TestGoHighLevelConnector_UpdateContact(t *testing.T) {
	t.Run("success with partial update", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "PUT" {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}

			// Parse request body
			json.NewDecoder(r.Body).Decode(&receivedBody)

			// Return updated contact
			response := `{
				"contact": {
					"id": "contact-123",
					"firstName": "UpdatedFirst",
					"lastName": "Doe",
					"email": "john@example.com",
					"phone": "+1234567890",
					"dateAdded": "2024-01-01T00:00:00.000Z",
					"dateUpdated": "2024-02-09T12:00:00.000Z",
					"tags": [],
					"customFields": []
				}
			}`

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		firstName := "UpdatedFirst"
		input := UpdateContactInput{
			FirstName: &firstName,
		}

		contact, err := connector.UpdateContact(context.Background(), "contact-123", input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if contact.FirstName != "UpdatedFirst" {
			t.Errorf("Expected updated first name, got '%s'", contact.FirstName)
		}

		// Verify only updated fields in request body
		if receivedBody["firstName"] != "UpdatedFirst" {
			t.Error("Expected firstName in request body")
		}
	})
}

// TestGoHighLevelConnector_DeleteContact tests contact deletion
func TestGoHighLevelConnector_DeleteContact(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE request, got %s", r.Method)
			}

			// Verify path
			if !strings.HasSuffix(r.URL.Path, "/contacts/contact-123") {
				t.Errorf("Expected path to end with '/contacts/contact-123', got '%s'", r.URL.Path)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.DeleteContact(context.Background(), "contact-123")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("error on not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			client:      server.Client(),
		}

		err := connector.DeleteContact(context.Background(), "contact-999")

		if err == nil {
			t.Fatal("Expected error for not found")
		}
	})
}

// TestGoHighLevelConnector_GetMetadata tests metadata retrieval
func TestGoHighLevelConnector_GetMetadata(t *testing.T) {
	connector := &GoHighLevelConnector{
		baseURL: "https://services.leadconnectorhq.com",
	}

	metadata := connector.GetMetadata()

	if metadata.PlatformName != "GoHighLevel" {
		t.Errorf("Expected platform name 'GoHighLevel', got '%s'", metadata.PlatformName)
	}

	if metadata.PlatformSlug != "gohighlevel" {
		t.Errorf("Expected platform slug 'gohighlevel', got '%s'", metadata.PlatformSlug)
	}

	if metadata.APIVersion != "v2" {
		t.Errorf("Expected API version 'v2', got '%s'", metadata.APIVersion)
	}

	if metadata.BaseURL != "https://services.leadconnectorhq.com" {
		t.Errorf("Expected BaseURL to match connector baseURL, got '%s'", metadata.BaseURL)
	}
}

// TestGoHighLevelConnector_TestConnection tests connection verification
func TestGoHighLevelConnector_TestConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify it hits the contacts endpoint with limit=1
			if !strings.Contains(r.URL.Path, "/contacts/") {
				t.Errorf("Expected /contacts/ path, got '%s'", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "1" {
				t.Error("Expected limit=1 for connection test")
			}

			// Return mock contacts response
			response := `{
				"contacts": [],
				"meta": {
					"total": 0,
					"nextPageUrl": "",
					"startAfterId": ""
				}
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		connector := &GoHighLevelConnector{
			accessToken: "test-token",
			baseURL:     server.URL,
			locationID:  "location-123",
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

		connector := &GoHighLevelConnector{
			accessToken: "invalid-token",
			baseURL:     server.URL,
			locationID:  "location-123",
			client:      server.Client(),
		}

		err := connector.TestConnection(context.Background())

		if err == nil {
			t.Fatal("Expected error for auth failure")
		}
	})
}

// TestGHLContact_toNormalized tests contact normalization
func TestGHLContact_toNormalized(t *testing.T) {
	gc := ghlContact{
		ID:          "contact-123",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		Phone:       "+1234567890",
		CompanyName: "ACME Corp",
		Tags:        []string{"VIP", "Customer"},
		CustomFields: []struct {
			ID    string      `json:"id"`
			Value interface{} `json:"value"`
		}{
			{ID: "field1", Value: "Premium"},
		},
		DateAdded:   "2024-01-01T00:00:00.000Z",
		DateUpdated: "2024-01-15T12:30:00.000Z",
	}

	normalized := gc.toNormalized()

	if normalized.ID != "contact-123" {
		t.Errorf("Expected ID 'contact-123', got '%s'", normalized.ID)
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
	if normalized.Company != "ACME Corp" {
		t.Errorf("Expected company 'ACME Corp', got '%s'", normalized.Company)
	}
	if len(normalized.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(normalized.Tags))
	}
	if normalized.CustomFields["field1"] != "Premium" {
		t.Error("Expected Premium custom field")
	}
	if normalized.CreatedAt == nil {
		t.Error("Expected CreatedAt timestamp")
	}
	if normalized.UpdatedAt == nil {
		t.Error("Expected UpdatedAt timestamp")
	}
}
