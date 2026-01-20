package domain

// Workspace represents the entire workspace structure
type Workspace struct {
	Version int         `json:"version"`
	Items   []WorkspaceItem `json:"items"`
}

// WorkspaceItem represents either a request or a folder
type WorkspaceItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // "request" or "folder"
	ParentID  *string `json:"parentId,omitempty"`
	CreatedAt int64   `json:"createdAt"`

	// Request-specific fields
	Method   *string     `json:"method,omitempty"`
	URL      *string     `json:"url,omitempty"`
	Headers  []Header    `json:"headers,omitempty"`
	BodyType *string     `json:"bodyType,omitempty"` // "none", "json", "form-data", "url-encoded"
	Body     *string     `json:"body,omitempty"`
	FormData []FormDataItem `json:"formData,omitempty"`
}

// Header represents an HTTP header
type Header struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// FormDataItem represents a form data field
type FormDataItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}
