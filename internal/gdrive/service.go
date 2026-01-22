package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

// Service provides Google Drive backup functionality for workspaces.
// It handles OAuth authentication and file uploads with automatic token sanitization.
type Service interface {
	// Authenticate initiates the OAuth flow and returns an authorization URL.
	// The user should visit this URL to grant permissions, then call CompleteAuth
	// with the authorization code. Uses client ID and client secret for server-side OAuth.
	Authenticate(ctx context.Context, clientID, clientSecret, redirectURI string) (authURL string, err error)

	// CompleteAuth completes the OAuth flow using the authorization code.
	// It exchanges the authorization code for an access token using client ID and client secret.
	// Returns the access token that can be used for subsequent API calls.
	CompleteAuth(ctx context.Context, clientID, clientSecret, redirectURI, authCode string) (accessToken string, err error)

	// BackupWorkspace uploads a sanitized version of the workspace to Google Drive.
	// The workspace is automatically sanitized to remove all sensitive headers and form data
	// before upload. Returns the file ID of the uploaded file.
	BackupWorkspace(ctx context.Context, accessToken string, workspace *domain.Workspace, filename string) (fileID string, err error)

	// ListBackups lists all workspace backup files in Google Drive.
	ListBackups(ctx context.Context, accessToken string) ([]BackupFile, error)

	// RestoreWorkspace downloads and restores a workspace from Google Drive.
	RestoreWorkspace(ctx context.Context, accessToken string, fileID string) (*domain.Workspace, error)
}

// BackupFile represents a workspace backup file in Google Drive.
type BackupFile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedTime time.Time `json:"createdTime"`
	Size        int64     `json:"size"`
}

// GoogleDriveService implements the Service interface using the Google Drive API.
type GoogleDriveService struct {
	// OAuth2Config can be set for custom OAuth configuration
	// If nil, uses default OAuth2 flow
}

// NewService creates a new Google Drive service instance.
func NewService() Service {
	return &GoogleDriveService{}
}

// Authenticate generates an OAuth authorization URL using client ID and client secret.
// This uses server-side OAuth flow where the backend exchanges the authorization code for tokens.
func (s *GoogleDriveService) Authenticate(ctx context.Context, clientID, clientSecret, redirectURI string) (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("client ID is required")
	}
	if clientSecret == "" {
		return "", fmt.Errorf("client secret is required")
	}
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/api/gdrive/callback"
	}

	// Create OAuth2 config
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"https://www.googleapis.com/auth/drive.file"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Generate authorization URL with state for security
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	return authURL, nil
}

// CompleteAuth exchanges the authorization code for an access token using client ID and client secret.
// This implements server-side OAuth flow where the backend securely exchanges the code for tokens.
func (s *GoogleDriveService) CompleteAuth(ctx context.Context, clientID, clientSecret, redirectURI, authCode string) (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("client ID is required")
	}
	if clientSecret == "" {
		return "", fmt.Errorf("client secret is required")
	}
	if authCode == "" {
		return "", fmt.Errorf("authorization code is required")
	}
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/api/gdrive/callback"
	}

	// Create OAuth2 config
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"https://www.googleapis.com/auth/drive.file"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Exchange authorization code for token
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Return the access token
	return token.AccessToken, nil
}

// BackupWorkspace uploads a sanitized workspace to Google Drive.
func (s *GoogleDriveService) BackupWorkspace(ctx context.Context, accessToken string, workspace *domain.Workspace, filename string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("access token is required")
	}
	if workspace == nil {
		return "", fmt.Errorf("workspace is required")
	}
	if filename == "" {
		filename = fmt.Sprintf("constrictor-workspace-%s.json", time.Now().Format("2006-01-02"))
	}

	// Sanitize the workspace to remove all sensitive information
	sanitized := domain.SanitizeWorkspace(workspace)

	// Marshal to JSON
	data, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal workspace: %w", err)
	}

	// Create Drive service with access token
	service, err := drive.NewService(ctx, option.WithTokenSource(
		&staticTokenSource{token: accessToken},
	))
	if err != nil {
		return "", fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Create file metadata
	fileMetadata := &drive.File{
		Name:     filename,
		MimeType: "application/json",
		Parents:  []string{"root"}, // Upload to root folder
	}

	// Upload file
	file, err := service.Files.Create(fileMetadata).
		Media(bytes.NewReader(data)).
		Fields("id,name,createdTime,size").
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return file.Id, nil
}

// ListBackups lists all workspace backup files in Google Drive.
func (s *GoogleDriveService) ListBackups(ctx context.Context, accessToken string) ([]BackupFile, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	service, err := drive.NewService(ctx, option.WithTokenSource(
		&staticTokenSource{token: accessToken},
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Search for files with name starting with "constrictor-workspace-"
	query := "name contains 'constrictor-workspace-' and mimeType = 'application/json' and trashed = false"
	files, err := service.Files.List().
		Q(query).
		Fields("files(id,name,createdTime,size)").
		OrderBy("createdTime desc").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	backups := make([]BackupFile, 0, len(files.Files))
	for _, file := range files.Files {
		createdTime := time.Now()
		if file.CreatedTime != "" {
			if parsed, err := time.Parse(time.RFC3339, file.CreatedTime); err == nil {
				createdTime = parsed
			}
		}
		// file.Size is already int64 in Google Drive API v3
		size := file.Size
		backups = append(backups, BackupFile{
			ID:          file.Id,
			Name:        file.Name,
			CreatedTime: createdTime,
			Size:        size,
		})
	}

	return backups, nil
}

// RestoreWorkspace downloads and restores a workspace from Google Drive.
func (s *GoogleDriveService) RestoreWorkspace(ctx context.Context, accessToken string, fileID string) (*domain.Workspace, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}
	if fileID == "" {
		return nil, fmt.Errorf("file ID is required")
	}

	service, err := drive.NewService(ctx, option.WithTokenSource(
		&staticTokenSource{token: accessToken},
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Download file
	resp, err := service.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Read file content
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal workspace
	var workspace domain.Workspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace: %w", err)
	}

	return &workspace, nil
}

// Helper type for token source

type staticTokenSource struct {
	token string
}

func (s *staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: s.token,
		TokenType:   "Bearer",
	}, nil
}
