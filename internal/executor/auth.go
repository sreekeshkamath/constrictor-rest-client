package executor

import (
	"encoding/base64"
	"fmt"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

// AuthResolver resolves authentication configuration and generates headers
type AuthResolver struct{}

// NewAuthResolver creates a new auth resolver
func NewAuthResolver() *AuthResolver {
	return &AuthResolver{}
}

// ResolveAuth resolves the authentication configuration for a workspace item
// by walking up the folder hierarchy if "inherit" is specified
func (r *AuthResolver) ResolveAuth(item *domain.WorkspaceItem, workspace *domain.Workspace) (*domain.AuthConfig, error) {
	// If no auth config, default to none
	if item.Auth == nil {
		return &domain.AuthConfig{Type: "none", Config: make(map[string]interface{})}, nil
	}

	// If not inherit, return the auth config directly
	if item.Auth.Type != "inherit" {
		return item.Auth, nil
	}

	// Walk up the folder hierarchy to find auth config
	visited := make(map[string]bool)
	currentParentID := item.ParentID
	for currentParentID != nil {
		if visited[*currentParentID] {
			return &domain.AuthConfig{Type: "none", Config: make(map[string]interface{})}, nil
		}
		visited[*currentParentID] = true

		// Find parent item
		var parentItem *domain.WorkspaceItem
		for i := range workspace.Items {
			if workspace.Items[i].ID == *currentParentID {
				parentItem = &workspace.Items[i]
				break
			}
		}

		if parentItem == nil {
			// Parent not found, default to none
			return &domain.AuthConfig{Type: "none", Config: make(map[string]interface{})}, nil
		}

		// If parent has auth config and it's not inherit, use it
		if parentItem.Auth != nil && parentItem.Auth.Type != "inherit" {
			return parentItem.Auth, nil
		}

		// If parent has inherit or no auth (nil), continue walking up
		if parentItem.Auth == nil || parentItem.Auth.Type == "inherit" {
			currentParentID = parentItem.ParentID
			continue
		}
	}

	// Reached root without finding auth, default to none
	return &domain.AuthConfig{Type: "none", Config: make(map[string]interface{})}, nil
}

// GenerateHeaders generates HTTP headers based on the authentication configuration
func (r *AuthResolver) GenerateHeaders(auth *domain.AuthConfig, reqMethod, reqURL string) (map[string]string, error) {
	if auth == nil || auth.Type == "none" {
		return make(map[string]string), nil
	}

	switch auth.Type {
	case "bearer":
		return r.generateBearerToken(auth)
	case "basic":
		return r.generateBasicAuth(auth)
	case "apikey":
		return r.generateAPIKey(auth)
	case "oauth2":
		return r.generateOAuth2(auth)
	case "oauth1":
		return r.generateOAuth1(auth, reqMethod, reqURL)
	case "digest":
		// Digest auth requires challenge-response, handled separately
		return r.generateDigestAuth(auth)
	case "ntlm":
		// NTLM requires multiple round trips, basic implementation
		return r.generateNTLM(auth)
	case "aws":
		return r.generateAWSAuth(auth, reqMethod, reqURL)
	case "hawk":
		return r.generateHawk(auth, reqMethod, reqURL)
	case "asap":
		return r.generateASAP(auth)
	case "netrc":
		// Netrc uses .netrc file, not headers
		return make(map[string]string), nil
	default:
		return make(map[string]string), nil
	}
}

// generateBearerToken generates Authorization header with Bearer token
func (r *AuthResolver) generateBearerToken(auth *domain.AuthConfig) (map[string]string, error) {
	if auth.Config == nil {
		return nil, fmt.Errorf("auth config is nil")
	}

	// Extract token from config - handle JSON unmarshaling which may produce different types
	tokenVal, exists := auth.Config["token"]
	if !exists {
		return nil, fmt.Errorf("bearer token is required (key 'token' not found in config)")
	}

	// Convert to string - handle various JSON types
	var token string
	switch v := tokenVal.(type) {
	case string:
		token = v
	case []byte:
		token = string(v)
	default:
		// Try fmt.Sprintf as fallback for other types
		token = fmt.Sprintf("%v", v)
	}

	if token == "" {
		return nil, fmt.Errorf("bearer token is required (token value is empty)")
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
	}, nil
}

// generateBasicAuth generates Authorization header with Basic auth
func (r *AuthResolver) generateBasicAuth(auth *domain.AuthConfig) (map[string]string, error) {
	username, ok1 := auth.Config["username"].(string)
	password, ok2 := auth.Config["password"].(string)
	if !ok1 || !ok2 || username == "" {
		return nil, fmt.Errorf("username and password are required for basic auth")
	}
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return map[string]string{
		"Authorization": "Basic " + credentials,
	}, nil
}

// generateAPIKey generates API key header
func (r *AuthResolver) generateAPIKey(auth *domain.AuthConfig) (map[string]string, error) {
	key, ok1 := auth.Config["key"].(string)
	value, ok2 := auth.Config["value"].(string)
	if !ok1 || !ok2 || key == "" || value == "" {
		return nil, fmt.Errorf("key and value are required for API key auth")
	}

	if loc, ok := auth.Config["location"].(string); ok && loc == "query" {
		return nil, fmt.Errorf("query-location API key auth not supported")
	}

	return map[string]string{
		key: value,
	}, nil
}

// generateOAuth2 generates Authorization header for OAuth 2.0
func (r *AuthResolver) generateOAuth2(auth *domain.AuthConfig) (map[string]string, error) {
	accessToken, ok1 := auth.Config["accessToken"].(string)
	if !ok1 || accessToken == "" {
		return nil, fmt.Errorf("accessToken is required for OAuth 2.0")
	}

	tokenType := "Bearer"
	if tt, ok := auth.Config["tokenType"].(string); ok && tt != "" {
		tokenType = tt
	}

	return map[string]string{
		"Authorization": tokenType + " " + accessToken,
	}, nil
}

// generateOAuth1 generates OAuth 1.0 signature (not implemented)
func (r *AuthResolver) generateOAuth1(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	consumerKey, ok1 := auth.Config["consumerKey"].(string)
	consumerSecret, ok2 := auth.Config["consumerSecret"].(string)
	if !ok1 || !ok2 || consumerKey == "" || consumerSecret == "" {
		return nil, fmt.Errorf("OAuth1 authentication not implemented: consumerKey and consumerSecret are required but the scheme requires proper HMAC-SHA1 signing")
	}
	return nil, fmt.Errorf("OAuth1 authentication not implemented: requires proper HMAC-SHA1 signing of the request signature base string")
}

// generateDigestAuth generates Digest authentication header (not implemented)
func (r *AuthResolver) generateDigestAuth(auth *domain.AuthConfig) (map[string]string, error) {
	username, ok1 := auth.Config["username"].(string)
	password, ok2 := auth.Config["password"].(string)
	if !ok1 || !ok2 || username == "" || password == "" {
		return nil, fmt.Errorf("Digest authentication not implemented: username and password are required but the scheme requires challenge-response handling with MD5/SHA hashing")
	}
	return nil, fmt.Errorf("Digest authentication not implemented: requires challenge-response handling to receive and respond to WWW-Authenticate challenge")
}

// generateNTLM generates NTLM authentication (not implemented)
func (r *AuthResolver) generateNTLM(auth *domain.AuthConfig) (map[string]string, error) {
	username, ok1 := auth.Config["username"].(string)
	password, ok2 := auth.Config["password"].(string)
	if !ok1 || !ok2 || username == "" || password == "" {
		return nil, fmt.Errorf("NTLM authentication not implemented: username and password are required but the scheme requires multiple round-trip handshake with NTLM protocol")
	}
	return nil, fmt.Errorf("NTLM authentication not implemented: requires multiple round-trip handshake with NTLM protocol negotiation")
}

// generateAWSAuth generates AWS Signature Version 4 headers (not implemented)
func (r *AuthResolver) generateAWSAuth(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	accessKeyID, ok1 := auth.Config["accessKeyId"].(string)
	secretAccessKey, ok2 := auth.Config["secretAccessKey"].(string)
	if !ok1 || !ok2 || accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("AWS authentication not implemented: accessKeyId and secretAccessKey are required but the scheme requires proper AWS Signature Version 4 signing")
	}
	return nil, fmt.Errorf("AWS authentication not implemented: requires proper AWS Signature Version 4 signing of the request with canonical URI, signed headers, and SHA256 hash")
}

// generateHawk generates Hawk authentication header (not implemented)
func (r *AuthResolver) generateHawk(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	authID, ok1 := auth.Config["authId"].(string)
	authKey, ok2 := auth.Config["authKey"].(string)
	if !ok1 || !ok2 || authID == "" || authKey == "" {
		return nil, fmt.Errorf("Hawk authentication not implemented: authId and authKey are required but the scheme requires proper HMAC-SHA256 signing with timestamp and nonce")
	}
	return nil, fmt.Errorf("Hawk authentication not implemented: requires proper HMAC-SHA256 signing of the request with timestamp, nonce, and hash of request payload")
}

// generateASAP generates Atlassian ASAP authentication header
func (r *AuthResolver) generateASAP(auth *domain.AuthConfig) (map[string]string, error) {
	token, ok := auth.Config["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required for ASAP auth")
	}

	return map[string]string{
		"Authorization": "Bearer " + token, // ASAP typically uses Bearer format
	}, nil
}
