package insomnia

import (
	"net/url"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/google/uuid"
)

func Convert(export *InsomniaExport) ([]domain.WorkspaceItem, error) {
	items := make([]domain.WorkspaceItem, 0)
	for _, item := range export.Collection {
		converted, err := convertItem(item, nil)
		if err != nil {
			return nil, err
		}
		items = append(items, converted...)
	}
	return items, nil
}

func convertItem(item InsomniaItem, parentID *string) ([]domain.WorkspaceItem, error) {
	items := make([]domain.WorkspaceItem, 0)

	if len(item.Children) > 0 {
		folderID := uuid.New().String()
		folder := domain.WorkspaceItem{
			ID:        folderID,
			Name:      item.Name,
			Type:      "folder",
			ParentID:  parentID,
			CreatedAt: item.Meta.Created,
		}
		items = append(items, folder)

		for _, child := range item.Children {
			converted, err := convertItem(child, &folderID)
			if err != nil {
				return nil, err
			}
			items = append(items, converted...)
		}
	} else {
		request := domain.WorkspaceItem{
			ID:        item.Meta.ID,
			Name:      item.Name,
			Type:      "request",
			ParentID:  parentID,
			CreatedAt: item.Meta.Created,
			Method:    &item.Method,
			URL:       &item.URL,
			Headers:   convertHeaders(item.Headers),
			BodyType:  convertBodyType(item.Body),
			Body:      convertBodyText(item.Body),
			FormData:  convertFormData(item.Body),
			Auth:      convertAuth(item.Authentication),
		}

		if item.Parameters != nil && request.URL != nil {
			url := mergeParametersIntoURL(*request.URL, item.Parameters)
			request.URL = &url
		}

		items = append(items, request)
	}

	return items, nil
}

func convertHeaders(headers []InsomniaHeader) []domain.Header {
	if len(headers) == 0 {
		return nil
	}
	result := make([]domain.Header, 0, len(headers))
	for _, h := range headers {
		result = append(result, domain.Header{
			Key:     h.Name,
			Value:   h.Value,
			Enabled: !h.Disabled,
		})
	}
	return result
}

func convertParametersToHeaders(params []InsomniaParam) []domain.Header {
	if len(params) == 0 {
		return nil
	}
	result := make([]domain.Header, 0, len(params))
	for _, p := range params {
		result = append(result, domain.Header{
			Key:     p.Name,
			Value:   p.Value,
			Enabled: !p.Disabled,
		})
	}
	return result
}

func convertBodyType(body *InsomniaBody) *string {
	if body == nil || body.MimeType == "" {
		none := "none"
		return &none
	}
	switch body.MimeType {
	case "application/json":
		json := "json"
		return &json
	case "multipart/form-data":
		formData := "form-data"
		return &formData
	case "application/x-www-form-urlencoded":
		urlEncoded := "url-encoded"
		return &urlEncoded
	default:
		none := "none"
		return &none
	}
}

func convertBodyText(body *InsomniaBody) *string {
	if body == nil {
		return nil
	}
	return &body.Text
}

func convertFormData(body *InsomniaBody) []domain.FormDataItem {
	if body == nil || body.Params == nil || len(body.Params) == 0 {
		return nil
	}
	result := make([]domain.FormDataItem, 0, len(body.Params))
	for _, p := range body.Params {
		result = append(result, domain.FormDataItem{
			Key:     p.Name,
			Value:   p.Value,
			Enabled: !p.Disabled,
		})
	}
	return result
}

func convertAuth(auth *InsomniaAuth) *domain.AuthConfig {
	if auth == nil || auth.Disabled || auth.Type == "" {
		none := "none"
		return &domain.AuthConfig{Type: none}
	}

	config := make(map[string]interface{})

	switch auth.Type {
	case "bearer":
		config["token"] = auth.Token
		if auth.Prefix != "" {
			config["prefix"] = auth.Prefix
		}
	case "basic":
		config["username"] = auth.Username
		config["password"] = auth.Password
	default:
		none := "none"
		return &domain.AuthConfig{Type: none}
	}

	return &domain.AuthConfig{
		Type:   auth.Type,
		Config: config,
	}
}

func mergeParametersIntoURL(originalURL string, params []InsomniaParam) string {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return originalURL
	}

	queryParams := parsedURL.Query()
	for _, p := range params {
		if !p.Disabled {
			queryParams.Set(url.QueryEscape(p.Name), url.QueryEscape(p.Value))
		}
	}
	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String()
}
