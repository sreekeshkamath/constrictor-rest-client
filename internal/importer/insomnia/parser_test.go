package insomnia

import (
	"testing"
)

func TestParse_ValidYAML(t *testing.T) {
	yamlContent := `type: collection.insomnia.rest/5.0
schema_version: "5.1"
name: Test Collection
meta:
  id: wrk_test
  created: 1717508275743
  modified: 1717508275743
  description: ""
collection:
  - name: Test Folder
    meta:
      id: fld_test123
      created: 1717508294832
      modified: 1717508294832
      sortKey: -1717508294832
      description: ""
    children:
      - url: http://example.com/api/test
        name: Test Request
        meta:
          id: req_test123
          created: 1717508296892
          modified: 1717508296892
          isPrivate: false
          description: ""
          sortKey: -1717508296892
        method: GET
        body:
          mimeType: application/json
          text: ""
        headers:
          - name: Content-Type
            value: application/json
        authentication:
          type: bearer
          token: "test-token"
        settings:
          renderRequestBody: true
          encodeUrl: true
          followRedirects: global
          cookies:
            send: true
            store: true
          rebuildPath: true
`

	result, err := Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Type != "collection.insomnia.rest/5.0" {
		t.Errorf("expected type 'collection.insomnia.rest/5.0', got '%s'", result.Type)
	}

	if result.Name != "Test Collection" {
		t.Errorf("expected name 'Test Collection', got '%s'", result.Name)
	}

	if len(result.Collection) != 1 {
		t.Errorf("expected 1 collection item, got %d", len(result.Collection))
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	invalidYAML := `invalid yaml content that is not properly formatted {`

	_, err := Parse([]byte(invalidYAML))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParse_MissingType(t *testing.T) {
	yamlContent := `schema_version: "5.1"
name: Test
collection: []
`

	_, err := Parse([]byte(yamlContent))
	if err != ErrMissingType {
		t.Errorf("expected ErrMissingType, got: %v", err)
	}
}

func TestParse_WrongType(t *testing.T) {
	yamlContent := `type: some.other.type
name: Test
collection: []
`

	_, err := Parse([]byte(yamlContent))
	if err == nil {
		t.Fatal("expected error for wrong type, got nil")
	}
}

func TestParse_NestedFolders(t *testing.T) {
	yamlContent := `type: collection.insomnia.rest/5.0
schema_version: "5.1"
name: Nested Test
meta:
  id: wrk_nested
  created: 1717508275743
  modified: 1717508275743
collection:
  - name: Root Folder
    meta:
      id: fld_root
      created: 1717508294832
      modified: 1717508294832
    children:
      - name: Nested Folder
        meta:
          id: fld_nested
          created: 1717508295000
          modified: 1717508295000
        children:
          - url: http://example.com/deep
            name: Deep Request
            meta:
              id: req_deep
              created: 1717508296000
              modified: 1717508296000
            method: POST
            body:
              mimeType: application/json
              text: '{"test": true}'
            headers: []
`

	result, err := Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(result.Collection) != 1 {
		t.Errorf("expected 1 root item, got %d", len(result.Collection))
	}

	root := result.Collection[0]
	if root.Name != "Root Folder" {
		t.Errorf("expected root folder name 'Root Folder', got '%s'", root.Name)
	}

	if len(root.Children) != 1 {
		t.Errorf("expected 1 nested item, got %d", len(root.Children))
	}

	nested := root.Children[0]
	if nested.Name != "Nested Folder" {
		t.Errorf("expected nested folder name 'Nested Folder', got '%s'", nested.Name)
	}

	if len(nested.Children) != 1 {
		t.Errorf("expected 1 deep item, got %d", len(nested.Children))
	}

	deep := nested.Children[0]
	if deep.Name != "Deep Request" {
		t.Errorf("expected deep request name 'Deep Request', got '%s'", deep.Name)
	}
	if deep.Method != "POST" {
		t.Errorf("expected method 'POST', got '%s'", deep.Method)
	}
}
