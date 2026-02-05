package insomnia

type InsomniaExport struct {
	Type          string         `yaml:"type"`
	SchemaVersion string         `yaml:"schema_version"`
	Name          string         `yaml:"name"`
	Meta          InsomniaMeta   `yaml:"meta"`
	Collection    []InsomniaItem `yaml:"collection"`
}

type InsomniaMeta struct {
	ID          string `yaml:"id"`
	Created     int64  `yaml:"created"`
	Modified    int64  `yaml:"modified"`
	Description string `yaml:"description"`
}

type InsomniaItem struct {
	Name     string         `yaml:"name,omitempty"`
	Meta     InsomniaMeta   `yaml:"meta,omitempty"`
	Children []InsomniaItem `yaml:"children,omitempty"`

	URL            string            `yaml:"url,omitempty"`
	Method         string            `yaml:"method,omitempty"`
	Body           *InsomniaBody     `yaml:"body,omitempty"`
	Headers        []InsomniaHeader  `yaml:"headers,omitempty"`
	Authentication *InsomniaAuth     `yaml:"authentication,omitempty"`
	Parameters     []InsomniaParam   `yaml:"parameters,omitempty"`
	Settings       *InsomniaSettings `yaml:"settings,omitempty"`
}

type InsomniaBody struct {
	MimeType string          `yaml:"mimeType"`
	Text     string          `yaml:"text"`
	Params   []InsomniaParam `yaml:"params,omitempty"`
}

type InsomniaHeader struct {
	Name        string `yaml:"name"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
	Disabled    bool   `yaml:"disabled"`
}

type InsomniaAuth struct {
	Type        string `yaml:"type"`
	Token       string `yaml:"token,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
	Disabled    bool   `yaml:"disabled,omitempty"`
	UseISO88591 bool   `yaml:"useISO88591,omitempty"`
	Prefix      string `yaml:"prefix,omitempty"`
}

type InsomniaParam struct {
	Name        string `yaml:"name"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
	Disabled    bool   `yaml:"disabled"`
}

type InsomniaSettings struct {
	RenderRequestBody bool   `yaml:"renderRequestBody"`
	EncodeUrl         bool   `yaml:"encodeUrl"`
	FollowRedirects   string `yaml:"followRedirects"`
	Cookies           struct {
		Send  bool `yaml:"send"`
		Store bool `yaml:"store"`
	} `yaml:"cookies"`
	RebuildPath bool `yaml:"rebuildPath"`
}
