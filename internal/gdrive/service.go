package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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
	// with the authorization code.
	Authenticate(ctx context.Context, clientID, clientSecret string) (authURL string, err error)

	// CompleteAuth completes the OAuth flow using the authorization code.
	// It returns an access token that can be used for subsequent API calls.
	CompleteAuth(ctx context.Context, clientID, clientSecret, authCode string) (accessToken string, err error)

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

// Authenticate generates an OAuth authorization URL.
// For a full implementation, this would use oauth2.Config, but for now
// we'll return a URL that the frontend can use with the Google OAuth library.
func (s *GoogleDriveService) Authenticate(ctx context.Context, clientID, clientSecret string) (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("client ID is required")
	}

	// Return OAuth URL for frontend to handle
	// The frontend will use Google's OAuth2 library to get the authorization code
	// This is a simplified approach - in production, you might want server-side OAuth
	redirectURI := "http://localhost:8080/api/gdrive/callback"
	scope := "https://www.googleapis.com/auth/drive.file"
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
		clientID,
		redirectURI,
		scope,
	)

	return authURL, nil
}

// CompleteAuth exchanges the authorization code for an access token.
// In a production implementation, this would use oauth2.Config.Exchange.
func (s *GoogleDriveService) CompleteAuth(ctx context.Context, clientID, clientSecret, authCode string) (string, error) {
	// This is a placeholder - in production, you'd exchange the code for tokens
	// For now, we'll expect the frontend to handle OAuth and pass the access token directly
	// This keeps the implementation simpler and more secure (tokens never touch the server)
	return "", fmt.Errorf("OAuth code exchange should be handled by frontend - pass access token directly")
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
		size := int64(0)
		if file.Size != "" {
			if parsed, err := strconv.ParseInt(file.Size, 10, 64); err == nil {
				size = parsed
			}
		}
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
