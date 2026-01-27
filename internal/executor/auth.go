package executor

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

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
	currentParentID := item.ParentID
	for currentParentID != nil {
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

		// If parent has inherit, continue walking up
		if parentItem.Auth != nil && parentItem.Auth.Type == "inherit" {
			currentParentID = parentItem.ParentID
			continue
		}

		// Parent has no auth or it's none, default to none
		return &domain.AuthConfig{Type: "none", Config: make(map[string]interface{})}, nil
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
	token, ok := auth.Config["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("bearer token is required")
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

// generateAPIKey generates API key header or query parameter
func (r *AuthResolver) generateAPIKey(auth *domain.AuthConfig) (map[string]string, error) {
	key, ok1 := auth.Config["key"].(string)
	value, ok2 := auth.Config["value"].(string)
	if !ok1 || !ok2 || key == "" || value == "" {
		return nil, fmt.Errorf("key and value are required for API key auth")
	}

	if loc, ok := auth.Config["location"].(string); ok && loc == "query" {
		// Query params are handled separately, return empty headers
		return make(map[string]string), nil
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

// generateOAuth1 generates OAuth 1.0 signature (simplified implementation)
func (r *AuthResolver) generateOAuth1(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	consumerKey, _ := auth.Config["consumerKey"].(string)
	consumerSecret, _ := auth.Config["consumerSecret"].(string)
	token, _ := auth.Config["token"].(string)
	_ = auth.Config["tokenSecret"] // Reserved for future use

	if consumerKey == "" || consumerSecret == "" {
		return nil, fmt.Errorf("consumerKey and consumerSecret are required for OAuth 1.0")
	}

	// Simplified OAuth 1.0 - in production, this should use proper signature generation
	// For now, we'll use a basic implementation
	params := url.Values{}
	params.Set("oauth_consumer_key", consumerKey)
	if token != "" {
		params.Set("oauth_token", token)
	}
	params.Set("oauth_signature_method", "HMAC-SHA1")
	params.Set("oauth_version", "1.0")

	// Note: Full OAuth 1.0 signature generation requires proper HMAC-SHA1 signing
	// This is a placeholder - full implementation would need crypto/hmac
	oauthHeader := "OAuth " + strings.ReplaceAll(params.Encode(), "&", ", ")
	return map[string]string{
		"Authorization": oauthHeader,
	}, nil
}

// generateDigestAuth generates Digest authentication header (simplified)
func (r *AuthResolver) generateDigestAuth(auth *domain.AuthConfig) (map[string]string, error) {
	username, ok1 := auth.Config["username"].(string)
	_, ok2 := auth.Config["password"].(string)
	if !ok1 || !ok2 || username == "" {
		return nil, fmt.Errorf("username and password are required for digest auth")
	}

	// Digest auth requires challenge-response, so we store credentials
	// The actual digest header is generated after receiving 401 with WWW-Authenticate
	// For now, return empty - this would need to be handled in the executor
	return make(map[string]string), nil
}

// generateNTLM generates NTLM authentication (simplified)
func (r *AuthResolver) generateNTLM(auth *domain.AuthConfig) (map[string]string, error) {
	username, ok1 := auth.Config["username"].(string)
	_, ok2 := auth.Config["password"].(string)
	if !ok1 || !ok2 || username == "" {
		return nil, fmt.Errorf("username and password are required for NTLM auth")
	}

	// NTLM requires multiple round trips, similar to digest
	// This is a placeholder
	return make(map[string]string), nil
}

// generateAWSAuth generates AWS Signature Version 4 headers
func (r *AuthResolver) generateAWSAuth(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	accessKeyID, ok1 := auth.Config["accessKeyId"].(string)
	secretAccessKey, ok2 := auth.Config["secretAccessKey"].(string)
	region, _ := auth.Config["region"].(string)
	service, _ := auth.Config["service"].(string)

	if !ok1 || !ok2 || accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("accessKeyId and secretAccessKey are required for AWS auth")
	}

	if region == "" {
		region = "us-east-1"
	}
	if service == "" {
		service = "execute-api"
	}

	// AWS Signature Version 4 is complex and requires proper signing
	// This is a placeholder - full implementation would need AWS SDK or crypto
	// For now, return basic structure
	return map[string]string{
		"Authorization": fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/...", accessKeyID),
	}, nil
}

// generateHawk generates Hawk authentication header
func (r *AuthResolver) generateHawk(auth *domain.AuthConfig, method, reqURL string) (map[string]string, error) {
	authID, ok1 := auth.Config["authId"].(string)
	authKey, ok2 := auth.Config["authKey"].(string)
	if !ok1 || !ok2 || authID == "" || authKey == "" {
		return nil, fmt.Errorf("authId and authKey are required for Hawk auth")
	}

	// Hawk requires proper signature generation with timestamp, nonce, etc.
	// This is a placeholder
	return map[string]string{
		"Authorization": fmt.Sprintf("Hawk id=\"%s\", ...", authID),
	}, nil
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
