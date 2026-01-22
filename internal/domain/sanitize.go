package domain

// SensitiveHeaderKeys is a set of header keys that should be removed from backups
// for security reasons. These typically contain authentication tokens, API keys,
// or other sensitive credentials.
var SensitiveHeaderKeys = map[string]bool{
	"authorization":       true,
	"x-api-key":          true,
	"x-auth-token":       true,
	"api-key":            true,
	"apikey":             true,
	"bearer":             true,
	"x-access-token":     true,
	"x-auth":             true,
	"cookie":             true,
	"x-csrf-token":       true,
	"x-requested-with":   true,
	"x-session-token":    true,
	"x-token":            true,
	"authentication":     true,
	"proxy-authorization": true,
}

// SensitiveFormDataKeys is a set of form data keys that should be removed from backups
// for security reasons.
var SensitiveFormDataKeys = map[string]bool{
	"password":           true,
	"token":              true,
	"api_key":            true,
	"apikey":             true,
	"access_token":       true,
	"refresh_token":     true,
	"secret":             true,
	"secret_key":         true,
	"auth_token":         true,
	"session_token":      true,
	"csrf_token":         true,
}

// SanitizeWorkspace creates a copy of the workspace with all sensitive information removed.
// This includes:
// - Headers with sensitive keys (case-insensitive matching)
// - Form data fields with sensitive keys (case-insensitive matching)
// - Body content is preserved (as it may contain non-sensitive data)
//
// This function is used before backing up the workspace to Google Drive or other
// cloud storage to ensure no credentials are leaked.
func SanitizeWorkspace(workspace *Workspace) *Workspace {
	if workspace == nil {
		return nil
	}

	sanitized := &Workspace{
		Version: workspace.Version,
		Items:   make([]WorkspaceItem, len(workspace.Items)),
	}

	for i, item := range workspace.Items {
		sanitized.Items[i] = SanitizeWorkspaceItem(item)
	}

	return sanitized
}

// SanitizeWorkspaceItem creates a sanitized copy of a workspace item with sensitive
// headers and form data removed.
func SanitizeWorkspaceItem(item WorkspaceItem) WorkspaceItem {
	sanitized := WorkspaceItem{
		ID:        item.ID,
		Name:      item.Name,
		Type:      item.Type,
		ParentID:  item.ParentID,
		CreatedAt: item.CreatedAt,
		Method:    item.Method,
		URL:       item.URL,
		BodyType:  item.BodyType,
		Body:      item.Body, // Body is preserved - may contain non-sensitive data
	}

	// Sanitize headers
	if len(item.Headers) > 0 {
		sanitized.Headers = make([]Header, 0, len(item.Headers))
		for _, header := range item.Headers {
			keyLower := toLower(header.Key)
			if !SensitiveHeaderKeys[keyLower] {
				sanitized.Headers = append(sanitized.Headers, header)
			}
			// Sensitive headers are simply omitted
		}
	}

	// Sanitize form data
	if len(item.FormData) > 0 {
		sanitized.FormData = make([]FormDataItem, 0, len(item.FormData))
		for _, formItem := range item.FormData {
			keyLower := toLower(formItem.Key)
			if !SensitiveFormDataKeys[keyLower] {
				sanitized.FormData = append(sanitized.FormData, formItem)
			}
			// Sensitive form data fields are simply omitted
		}
	}

	return sanitized
}

// toLower is a helper function that safely converts a string to lowercase.
// It handles empty strings and nil cases.
func toLower(s string) string {
	if s == "" {
		return ""
	}
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}
