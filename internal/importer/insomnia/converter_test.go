package insomnia

import (
	"testing"
)

func TestConvert_BasicRequest(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Test Request",
				Meta: InsomniaMeta{
					ID:       "req_123",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Headers: []InsomniaHeader{
					{Name: "Accept", Value: "application/json", Disabled: false},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.Type != "request" {
		t.Errorf("expected type 'request', got '%s'", item.Type)
	}
	if item.Name != "Test Request" {
		t.Errorf("expected name 'Test Request', got '%s'", item.Name)
	}
	if item.ID != "req_123" {
		t.Errorf("expected ID 'req_123', got '%s'", item.ID)
	}
	if item.URL == nil || *item.URL != "http://example.com/api" {
		t.Errorf("expected URL 'http://example.com/api', got '%v'", item.URL)
	}
	if item.Method == nil || *item.Method != "GET" {
		t.Errorf("expected method 'GET', got '%v'", item.Method)
	}
	if len(item.Headers) != 1 {
		t.Errorf("expected 1 header, got %d", len(item.Headers))
	}
	if item.Headers[0].Key != "Accept" {
		t.Errorf("expected header key 'Accept', got '%s'", item.Headers[0].Key)
	}
}

func TestConvert_RequestWithBody(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "POST Request",
				Meta: InsomniaMeta{
					ID:       "req_post",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "POST",
				Body: &InsomniaBody{
					MimeType: "application/json",
					Text:     `{"key": "value"}`,
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.BodyType == nil || *item.BodyType != "json" {
		t.Errorf("expected bodyType 'json', got '%v'", item.BodyType)
	}
	if item.Body == nil || *item.Body != `{"key": "value"}` {
		t.Errorf("expected body '{\"key\": \"value\"}', got '%v'", item.Body)
	}
}

func TestConvert_RequestWithFormData(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Form Request",
				Meta: InsomniaMeta{
					ID:       "req_form",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/submit",
				Method: "POST",
				Body: &InsomniaBody{
					MimeType: "multipart/form-data",
					Params: []InsomniaParam{
						{Name: "field1", Value: "value1", Disabled: false},
						{Name: "field2", Value: "value2", Disabled: true},
					},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.BodyType == nil || *item.BodyType != "form-data" {
		t.Errorf("expected bodyType 'form-data', got '%v'", item.BodyType)
	}
	if len(item.FormData) != 2 {
		t.Errorf("expected 2 form data items, got %d", len(item.FormData))
	}
	if item.FormData[0].Key != "field1" || item.FormData[0].Value != "value1" {
		t.Errorf("expected formData[0] 'field1=value1', got '%s=%s'", item.FormData[0].Key, item.FormData[0].Value)
	}
	if !item.FormData[0].Enabled {
		t.Error("expected formData[0] to be enabled")
	}
	if item.FormData[1].Enabled {
		t.Error("expected formData[1] to be disabled")
	}
}

func TestConvert_RequestWithBearerAuth(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Auth Request",
				Meta: InsomniaMeta{
					ID:       "req_auth",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Authentication: &InsomniaAuth{
					Type:  "bearer",
					Token: "secret-token",
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.Auth == nil {
		t.Fatal("expected auth to be set")
	}
	if item.Auth.Type != "bearer" {
		t.Errorf("expected auth type 'bearer', got '%s'", item.Auth.Type)
	}
	token, ok := item.Auth.Config["token"]
	if !ok {
		t.Error("expected 'token' in config")
	}
	if token != "secret-token" {
		t.Errorf("expected token 'secret-token', got '%v'", token)
	}
}

func TestConvert_RequestWithBasicAuth(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Basic Auth Request",
				Meta: InsomniaMeta{
					ID:       "req_basic",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Authentication: &InsomniaAuth{
					Type:     "basic",
					Username: "user",
					Password: "pass",
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.Auth == nil {
		t.Fatal("expected auth to be set")
	}
	if item.Auth.Type != "basic" {
		t.Errorf("expected auth type 'basic', got '%s'", item.Auth.Type)
	}
	username, ok := item.Auth.Config["username"]
	if !ok {
		t.Error("expected 'username' in config")
	}
	if username != "user" {
		t.Errorf("expected username 'user', got '%v'", username)
	}
}

func TestConvert_FolderHierarchy(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Folder 1",
				Meta: InsomniaMeta{
					ID:      "fld_1",
					Created: 1717508294832,
				},
				Children: []InsomniaItem{
					{
						Name: "Folder 2",
						Meta: InsomniaMeta{
							ID:      "fld_2",
							Created: 1717508295000,
						},
						Children: []InsomniaItem{
							{
								Name: "Nested Request",
								Meta: InsomniaMeta{
									ID:      "req_nested",
									Created: 1717508296000,
								},
								URL:    "http://example.com/nested",
								Method: "GET",
							},
						},
					},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items (folder, nested folder, request), got %d", len(items))
	}

	folder1 := items[0]
	if folder1.Type != "folder" {
		t.Errorf("expected item[0] to be folder, got '%s'", folder1.Type)
	}
	if folder1.Name != "Folder 1" {
		t.Errorf("expected folder1 name 'Folder 1', got '%s'", folder1.Name)
	}
	if folder1.ParentID != nil {
		t.Error("expected folder1 to have nil ParentID")
	}

	folder2 := items[1]
	if folder2.Type != "folder" {
		t.Errorf("expected item[1] to be folder, got '%s'", folder2.Type)
	}
	if folder2.ParentID == nil || *folder2.ParentID != folder1.ID {
		t.Errorf("expected folder2 ParentID to be '%s', got '%v'", folder1.ID, folder2.ParentID)
	}

	request := items[2]
	if request.Type != "request" {
		t.Errorf("expected item[2] to be request, got '%s'", request.Type)
	}
	if request.ParentID == nil || *request.ParentID != folder2.ID {
		t.Errorf("expected request ParentID to be '%s', got '%v'", folder2.ID, request.ParentID)
	}
}

func TestConvert_HeaderMapping(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Header Test",
				Meta: InsomniaMeta{
					ID:       "req_headers",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Headers: []InsomniaHeader{
					{Name: "Enabled-Header", Value: "enabled", Disabled: false},
					{Name: "Disabled-Header", Value: "disabled", Disabled: true},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if len(item.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(item.Headers))
	}

	if !item.Headers[0].Enabled {
		t.Error("expected first header to be enabled")
	}
	if item.Headers[0].Key != "Enabled-Header" || item.Headers[0].Value != "enabled" {
		t.Errorf("expected header 'Enabled-Header: enabled', got '%s: %s'", item.Headers[0].Key, item.Headers[0].Value)
	}

	if item.Headers[1].Enabled {
		t.Error("expected second header to be disabled")
	}
}

func TestConvert_URLEncodedBody(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "URL Encoded Request",
				Meta: InsomniaMeta{
					ID:       "req_urlenc",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/submit",
				Method: "POST",
				Body: &InsomniaBody{
					MimeType: "application/x-www-form-urlencoded",
					Params: []InsomniaParam{
						{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"},
					},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.BodyType == nil || *item.BodyType != "url-encoded" {
		t.Errorf("expected bodyType 'url-encoded', got '%v'", item.BodyType)
	}
	if len(item.FormData) != 2 {
		t.Errorf("expected 2 form data items, got %d", len(item.FormData))
	}
}

func TestConvert_ParametersAsQueryParams(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Param Request",
				Meta: InsomniaMeta{
					ID:       "req_params",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api?param1=value1",
				Method: "GET",
				Parameters: []InsomniaParam{
					{Name: "param2", Value: "value2"},
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if len(item.Headers) != 0 {
		t.Errorf("expected 0 headers from parameters, got %d", len(item.Headers))
	}
	if item.URL == nil {
		t.Fatal("expected URL to be set")
	}
	expectedURL := "http://example.com/api?param1=value1&param2=value2"
	if *item.URL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, *item.URL)
	}
}

func TestConvert_NoAuthentication(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "No Auth Request",
				Meta: InsomniaMeta{
					ID:       "req_noauth",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.Auth == nil {
		t.Fatal("expected auth to be set (to 'none')")
	}
	if item.Auth.Type != "none" {
		t.Errorf("expected auth type 'none', got '%s'", item.Auth.Type)
	}
}

func TestConvert_NilBody(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Nil Body Request",
				Meta: InsomniaMeta{
					ID:       "req_nilbody",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Body:   nil,
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.BodyType == nil || *item.BodyType != "none" {
		t.Errorf("expected bodyType 'none', got '%v'", item.BodyType)
	}
	if item.Body != nil {
		t.Errorf("expected nil body, got '%v'", item.Body)
	}
	if item.FormData != nil {
		t.Errorf("expected nil formData, got '%v'", item.FormData)
	}
}

func TestConvert_DisabledAuth(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Disabled Auth Request",
				Meta: InsomniaMeta{
					ID:       "req_disabledauth",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "GET",
				Authentication: &InsomniaAuth{
					Type:     "bearer",
					Token:    "token",
					Disabled: true,
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.Auth == nil {
		t.Fatal("expected auth to be set")
	}
	if item.Auth.Type != "none" {
		t.Errorf("expected auth type 'none', got '%s'", item.Auth.Type)
	}
}

func TestConvert_CustomBodyMimeType(t *testing.T) {
	export := &InsomniaExport{
		Type:          "collection.insomnia.rest/5.0",
		SchemaVersion: "5.1",
		Name:          "Test",
		Collection: []InsomniaItem{
			{
				Name: "Custom Mime Type Request",
				Meta: InsomniaMeta{
					ID:       "req_custommime",
					Created:  1717508296892,
					Modified: 1717508296892,
				},
				URL:    "http://example.com/api",
				Method: "POST",
				Body: &InsomniaBody{
					MimeType: "text/plain",
					Text:     "plain text body",
				},
			},
		},
	}

	items, err := Convert(export)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	item := items[0]
	if item.BodyType == nil || *item.BodyType != "none" {
		t.Errorf("expected bodyType 'none' for unknown mimeType, got '%v'", item.BodyType)
	}
	if item.Body == nil || *item.Body != "plain text body" {
		t.Errorf("expected body 'plain text body', got '%v'", item.Body)
	}
}
