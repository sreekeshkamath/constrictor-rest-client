# Plan: Import API Collections from Insomnia

## Overview
Implement the ability to import API collections from Insomnia YAML exports into constrictor-rest-client. The import process creates a new workspace file, preserves the current workspace, and allows easy switching between workspaces via a dropdown UI.

## Requirements Confirmed

| Requirement | Decision |
|-------------|----------|
| Postman Support | Deferred to future iteration |
| Import Creates New Workspace | Yes - doesn't overwrite current |
| Auto-save Current | Current workspace saved before import |
| Workspace Naming | "Imported from Insomnia - {timestamp}" with UI rename option |
| Hierarchy Preservation | Converted to parentId references in flat items |
| Auth Preservation | Bearer/basic auth converted to Authorization headers |
| Workspace Switching | Dropdown UI for easy switching |

## Architecture

### Directory Structure
```text
data/
  workspace.json          # Workspace registry (v2 format)
  workspaces/
    default/              # Default workspace
      workspace.json
    imported_YYYYMMDD_HHMMSS/  # Imported workspaces
      workspace.json
```

### Workspace Registry Format
```json
{
  "version": 2,
  "activeWorkspaceId": "default",
  "workspaces": [
    {
      "id": "default",
      "name": "Default Workspace",
      "path": "workspaces/default/workspace.json",
      "createdAt": 1769071038217,
      "modifiedAt": 1769071038217
    }
  ]
}
```

6.12. Implement HandlePutWorkspace (update current workspace):
    ```go
    func (h *Handlers) HandlePutWorkspace(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]

        var workspace domain.Workspace
        if err := json.NewDecoder(r.Body).Decode(&workspace); err != nil {
            respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
            return
        }

        // Validate workspace ID matches the URL
        if workspace.ID != "" && workspace.ID != id {
            respondError(w, http.StatusBadRequest, "workspace ID mismatch", "")
            return
        }

        // Ensure ID is set
        workspace.ID = id

        if err := h.store.SaveWorkspace(&workspace); err != nil {
            respondError(w, http.StatusInternalServerError, "failed to save workspace", err.Error())
            return
        }

        respondJSON(w, http.StatusOK, map[string]bool{"success": true})
    }
    ```

**Testing**:
- Test `ListWorkspaces()` returns empty array initially
- Test `SaveWorkspaceAs()` creates directory and registry entry
- Test `GetWorkspace()` loads saved workspace
- Test `SetActiveWorkspace()` updates active ID
- Test `DeleteWorkspace()` prevents deleting default
- Test concurrent access with mutex

---

### Step 2: Create Workspace Registry Migration

**Goal**: Convert existing single-workspace format to v2 registry format

**Tasks**:
- Detect if `workspace.json` is v1 or v2 format
- If v1: Create `workspaces/default/`, move content there, create registry
- If v2: Use as-is
- Handle rollback on failure

**Files**:
- Create: `internal/storage/migrate.go`

**Output**: Automatic migration on startup, backward compatibility

**Detailed Tasks**:
2.1. Create migration function that runs at startup:
    ```go
    func MigrateIfNeeded(store *FileStore, dataDir string) error {
        // Check if already v2 format
        if isV2Format(store.GetPath()) {
            return nil
        }
        return migrateFromV1ToV2(store, dataDir)
    }
    ```

2.2. Implement `isV2Format(path string)`:
    - Read first bytes of workspace.json
    - Check if contains "workspaces" array (v2) or just "items" (v1)

2.3. Implement `migrateFromV1ToV2(store, dataDir)`:
    - Load existing workspace from store
    - Create `workspaces/default/` directory
    - Save workspace to `workspaces/default/workspace.json`
    - Create registry file with default workspace entry
    - Remove old workspace.json or back it up
    - Test read/write works before committing

2.4. Handle rollback on failure:
    - If migration fails, restore old file
    - Log error with details
    - Don't prevent application startup (use empty workspace on error)

**Testing**:
- Test migration creates correct directory structure
- Test registry file has correct format
- Test workspace loads correctly after migration
- Test old workspace.json is backed up or removed
- Test concurrent access during migration

---

### Step 3: Define Insomnia Format Models

**Goal**: Create Go structs matching Insomnia YAML export format

**Models Required**:
```go
package importmodels

// InsomniaCollection - Top-level structure
type InsomniaCollection struct {
    Type          string          `yaml:"type"`
    SchemaVersion string          `yaml:"schema_version"`
    Name          string          `yaml:"name"`
    Meta          InsomniaMeta    `yaml:"meta"`
    Collection    []InsomniaItem  `yaml:"collection"`
}

// InsomniaMeta - Collection metadata
type InsomniaMeta struct {
    ID          string `yaml:"id"`
    Created     int64  `yaml:"created"`
    Modified    int64  `yaml:"modified"`
    Description string `yaml:"description"`
}

// InsomniaItem - Folder or request (discriminated by Children vs URL)
type InsomniaItem struct {
    Name     string               `yaml:"name"`
    Meta     InsomniaItemMeta     `yaml:"meta"`
    Children []InsomniaItem       `yaml:"children,omitempty"` // Folder has children
    URL      string               `yaml:"url,omitempty"`      // Request has URL
    Method   string               `yaml:"method,omitempty"`
    Body     InsomniaBody         `yaml:"body,omitempty"`
    Headers  []InsomniaHeader     `yaml:"headers,omitempty"`
    Auth     InsomniaAuth         `yaml:"authentication,omitempty"`
    Settings map[string]interface{} `yaml:"settings,omitempty"`
}

// InsomniaItemMeta - Item metadata
type InsomniaItemMeta struct {
    ID          string `yaml:"id"`
    Created     int64  `yaml:"created"`
    Modified    int64  `yaml:"modified"`
    IsPrivate   bool   `yaml:"isPrivate"`
    Description string `yaml:"description"`
    SortKey     string `yaml:"sortKey"`
}

// InsomniaAuth - Authentication configuration
type InsomniaAuth struct {
    Type     string `yaml:"type"` // "bearer", "basic", "apikey", "aws_sig_v4", "netrc"
    Token    string `yaml:"token,omitempty"`
    Username string `yaml:"username,omitempty"`
    Password string `yaml:"password,omitempty"`
    Disabled bool   `yaml:"disabled,omitempty"`
    UseISO88591 bool `yaml:"useISO88591,omitempty"`
}

// InsomniaBody - Request body
type InsomniaBody struct {
    MIMEType string            `yaml:"mimeType"`
    Text     string            `yaml:"text,omitempty"`
    Params   []InsomniaParam   `yaml:"params,omitempty"` // For form-data
}

// InsomniaHeader - HTTP header
type InsomniaHeader struct {
    Name        string `yaml:"name"`
    Value       string `yaml:"value"`
    Description string `yaml:"description,omitempty"`
    Disabled    bool   `yaml:"disabled,omitempty"`
}

// InsomniaParam - Body parameter (for form-data)
type InsomniaParam struct {
    Name        string `yaml:"name"`
    Value       string `yaml:"value"`
    Description string `yaml:"description,omitempty"`
    Disabled    bool   `yaml:"disabled,omitempty"`
}

// InsomniaSettings - Request settings
type InsomniaSettings struct {
    RenderRequestBody bool     `yaml:"renderRequestBody"`
    EncodeUrl         bool     `yaml:"encodeUrl"`
    FollowRedirects   string   `yaml:"followRedirects"` // "global", "yes", "no"
    SendCookies       bool     `yaml:"cookies.send"`
    StoreCookies      bool     `yaml:"cookies.store"`
    RebuildPath       bool     `yaml:"rebuildPath"`
}
```

**Files**:
- Create: `internal/import/models.go`

**Output**: Structs for parsing Insomnia YAML format

**Detailed Tasks**:
3.1. Create the importmodels package structure
3.2. Define all structs with appropriate YAML tags
3.3. Add validation tags where applicable
3.4. Handle optional fields with omitempty
3.5. Test parsing with sample Insomnia YAML

---

### Step 4: Implement Insomnia Importer

**Goal**: Parse Insomnia YAML and convert to Constrictor workspace format

**Files**:
- Create: `internal/import/importers/insomnia.go`

**Output**: Complete importer with all Insomnia features supported

**Detailed Tasks**:
4.1. Create InsomniaImporter struct:
    ```go
    type InsomniaImporter struct {
        yamlPath string
    }
    ```

4.2. Implement Parse method:
    ```go
    func (i *InsomniaImporter) Parse(filePath string) (*importmodels.InsomniaCollection, error) {
        data, err := os.ReadFile(filePath)
        if err != nil {
            return nil, fmt.Errorf("failed to read file: %w", err)
        }
        var collection importmodels.InsomniaCollection
        if err := yaml.Unmarshal(data, &collection); err != nil {
            return nil, fmt.Errorf("failed to parse YAML: %w", err)
        }
        return &collection, nil
    }
    ```

4.3. Implement Convert method with recursive folder processing:
```go
func (i *InsomniaImporter) Convert(insomnia *importmodels.InsomniaCollection) (*domain.Workspace, error) {
    workspace := &domain.Workspace{
        Version: 2,
        Items:   make([]domain.WorkspaceItem, 0),
    }

    // Process collection recursively
    for _, item := range insomnia.Collection {
        i.processItem(item, nil, workspace)
    }

    return workspace, nil
}
```

4.4. Implement recursive processItem method:
```go
func (i *InsomniaImporter) processItem(
    item importmodels.InsomniaItem,
    parentID *string,
    workspace *domain.Workspace,
) {
    // Check if this is a folder (has Children) or request (has URL)
    if len(item.Children) > 0 {
        // It's a folder
        folderID := generateUUID()
            folder := domain.WorkspaceItem{
                ID:        folderID,
                Name:      item.Name,
                Type:      "folder",
                ParentID:  parentID,
                CreatedAt: item.Meta.Created,
            }
            workspace.Items = append(workspace.Items, folder)
            
            // Process children
            for _, child := range item.Children {
                i.processItem(child, &folderID, workspace)
            }
        } else if item.URL != "" {
            // It's a request
            workspace.Items = append(workspace.Items, i.convertRequest(item, parentID))
        }
    }
    ```

4.5. Implement convertRequest method with full conversion:
    ```go
    func (i *InsomniaImporter) convertRequest(
        req importmodels.InsomniaItem,
        parentID *string,
    ) domain.WorkspaceItem {
        // Convert regular headers
        headers := i.convertHeaders(req.Headers)

        // Convert auth headers and merge (auth headers take precedence)
        authHeaders := i.convertAuth(req.Auth)
        if len(authHeaders) > 0 {
            headers = append(authHeaders, headers...)
        }

        item := domain.WorkspaceItem{
            ID:        generateUUIDFromMeta(req.Meta.ID),
            Name:      req.Name,
            Type:      "request",
            ParentID:  parentID,
            CreatedAt: req.Meta.Created,
            Method:    &req.Method,
            URL:       &req.URL,
            Headers:   headers,
            BodyType:  i.mapBodyType(req.Body.MIMEType),
            Body:      i.getBodyText(req.Body),
            FormData:  i.convertFormData(req.Body.Params),
        }
        return item
    }
    ```

4.6. Implement header conversion with auth:
    ```go
    func (i *InsomniaImporter) convertHeaders(headers []importmodels.InsomniaHeader) []domain.Header {
        result := make([]domain.Header, 0)
        for _, h := range headers {
            if h.Disabled {
                continue
            }
            result = append(result, domain.Header{
                Key:     h.Name,
                Value:   h.Value,
                Enabled: !h.Disabled,
            })
        }
        return result
    }
    ```

4.7. Implement auth conversion to Authorization header:
    ```go
    func (i *InsomniaImporter) convertAuth(auth importmodels.InsomniaAuth) []domain.Header {
        if auth.Disabled || auth.Type == "" {
            return nil
        }
        
        switch auth.Type {
        case "bearer":
            return []domain.Header{
                {
                    Key:     "Authorization",
                    Value:   "Bearer " + auth.Token,
                    Enabled: true,
                },
            }
        case "basic":
            credentials := base64.StdEncoding.EncodeToString(
                []byte(auth.Username + ":" + auth.Password),
            )
            return []domain.Header{
                {
                    Key:     "Authorization",
                    Value:   "Basic " + credentials,
                    Enabled: true,
                },
            }
        case "apikey":
            // Note: API key authentication headers vary by API.
            // Using "X-API-Key" as a common default - you may need to adjust this
            // based on your target API (e.g., "Authorization: Bearer <token>",
            // "X-DreamFactory-Api-Key", "Api-Key", etc.).
            // Verify and update the header name after import if needed.
            return []domain.Header{
                {
                    Key:     "X-API-Key",
                    Value:   auth.Token,
                    Enabled: true,
                },
            }
        }
        return nil
    }
    ```

4.8. Implement body type mapping:
    ```go
    func (i *InsomniaImporter) mapBodyType(mimeType string) *string {
        typeMap := map[string]string{
            "application/json":               "json",
            "multipart/form-data":            "form-data",
            "application/x-www-form-urlencoded": "url-encoded",
            "":                               "none",
        }
        result := "none"
        if mapped, ok := typeMap[mimeType]; ok {
            result = mapped
        }
        return &result
    }
    ```

4.9. Implement form data conversion:
    ```go
    func (i *InsomniaImporter) convertFormData(params []importmodels.InsomniaParam) []domain.FormDataItem {
        result := make([]domain.FormDataItem, 0)
        for _, p := range params {
            if p.Disabled {
                continue
            }
            result = append(result, domain.FormDataItem{
                Key:     p.Name,
                Value:   p.Value,
                Enabled: !p.Disabled,
            })
        }
        return result
    }
    ```

4.10. Implement body text extraction:
    ```go
    func (i *InsomniaImporter) getBodyText(body importmodels.InsomniaBody) *string {
        text := body.Text
        return &text
    }
    ```

4.11. Implement UUID generation:
    ```go
    func generateUUID() string {
        return uuid.New().String()
    }
    
    func generateUUIDFromMeta(id string) string {
        if id != "" {
            // Insomnia IDs have prefix like "req_3b2fde2e"
            // We want a clean UUID format
            return uuid.New().String()
        }
        return generateUUID()
    }
    ```

**Testing**:
- Test parsing Insomnia YAML file
- Test folder hierarchy conversion (parentId)
- Test request conversion (method, url, headers)
- Test auth conversion (bearer, basic, apikey)
- Test body type mapping
- Test form data conversion
- Test disabled fields are skipped
- Test empty import creates empty workspace

---

### Step 5: Create Import Service Facade

**Goal**: Unified interface for importing and workspace creation

**Files**:
- Create: `internal/import/service.go`

**Output**: Clean service interface for importing

**Detailed Tasks**:
5.1. Create ImportService struct:
    ```go
    type ImportService struct {
        store    *storage.MultiWorkspaceStore
        importer *importers.InsomniaImporter
    }
    ```

5.2. Implement ImportInsomnia method:
    ```go
    func (s *ImportService) ImportInsomnia(filePath string) (*domain.Workspace, error) {
        // Step 1: Auto-save current workspace
        current, err := s.store.Load()
        if err != nil {
            return nil, fmt.Errorf("failed to load current workspace: %w", err)
        }
        if err := s.store.Save(current); err != nil {
            return nil, fmt.Errorf("failed to save current workspace: %w", err)
        }
        
        // Step 2: Parse Insomnia file
        insomnia, err := s.importer.Parse(filePath)
        if err != nil {
            return nil, fmt.Errorf("failed to parse Insomnia file: %w", err)
        }
        
        // Step 3: Convert to workspace
        workspace, err := s.importer.Convert(insomnia)
        if err != nil {
            return nil, fmt.Errorf("failed to convert workspace: %w", err)
        }
        
        // Step 4: Generate workspace ID and name
        timestamp := time.Now().Format("20060102_150405")
        workspaceID := fmt.Sprintf("imported_%s", timestamp)
        workspaceName := fmt.Sprintf("Imported from Insomnia - %s", time.Now().Format("2006-01-02 15:04"))
        
        // Step 5: Save as new workspace
        if err := s.store.SaveWorkspaceAs(workspaceID, workspace, workspaceName); err != nil {
            return nil, fmt.Errorf("failed to save workspace: %w", err)
        }
        
        // Step 6: Activate new workspace
        if err := s.store.SetActiveWorkspace(workspaceID); err != nil {
            return nil, fmt.Errorf("failed to activate workspace: %w", err)
        }
        
        return workspace, nil
    }
    ```

5.3. Implement statistics calculation:
    ```go
    func (s *ImportService) CalculateStatistics(workspace *domain.Workspace) ImportStatistics {
        stats := ImportStatistics{}
        for _, item := range workspace.Items {
            if item.Type == "request" {
                stats.Requests++
                stats.TotalItems++
                stats.Headers += len(item.Headers)
                if item.Body != nil {
                    stats.BodySize += len(*item.Body)
                }
            } else if item.Type == "folder" {
                stats.Folders++
                stats.TotalItems++
            }
        }
        return stats
    }
    ```

5.4. Create NewImportService factory:
    ```go
    func NewImportService(store *storage.MultiWorkspaceStore) *ImportService {
        return &ImportService{
            store:    store,
            importer: &importers.InsomniaImporter{},
        }
    }
    ```

5.5. Add ImportStatistics and ImportError types:
    ```go
    type ImportStatistics struct {
        TotalItems int `json:"totalItems"`
        Requests   int `json:"requests"`
        Folders    int `json:"folders"`
        Headers    int `json:"headers"`
        BodySize   int `json:"bodySize"`
    }
    
    type ImportError struct {
        Item   string `json:"item"`
        Message string `json:"message"`
        Line   int    `json:"line,omitempty"`
    }
    
    type ImportResult struct {
        Success     bool              `json:"success"`
        Workspace   *domain.Workspace `json:"workspace,omitempty"`
        Statistics  ImportStatistics  `json:"statistics,omitempty"`
        Errors      []ImportError     `json:"errors,omitempty"`
    }
    ```

**Testing**:
- Test auto-save before import
- Test import creates new workspace
- Test workspace ID generation
- Test statistics calculation
- Test error handling (file not found, parse error)

---

### Step 6: Add Workspace Management API

**Goal**: HTTP endpoints for workspace management and import

**Files**:
- Modify: `internal/httpapi/handlers.go`
- Modify: `internal/httpapi/router.go`

**Output**: Full REST API for workspace management

**Detailed Tasks**:
6.1. Add workspace-related types to handlers:
    ```go
    type WorkspaceListResponse struct {
        ActiveWorkspaceID string                  `json:"activeWorkspaceId"`
        Workspaces        []storage.WorkspaceInfo `json:"workspaces"`
    }
    
    type WorkspaceResponse struct {
        Workspace *domain.Workspace              `json:"workspace"`
        Metadata  storage.WorkspaceInfo          `json:"metadata"`
    }
    
    type CreateWorkspaceRequest struct {
        Name string `json:"name"`
    }
    
    type RenameWorkspaceRequest struct {
        Name string `json:"name"`
    }
    ```

6.2. Implement HandleListWorkspaces:
    ```go
    func (h *Handlers) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
        workspaces, err := h.store.ListWorkspaces()
        if err != nil {
            respondError(w, http.StatusInternalServerError, "failed to list workspaces", err.Error())
            return
        }
        
        respondJSON(w, http.StatusOK, WorkspaceListResponse{
            ActiveWorkspaceID: h.store.GetActiveWorkspaceID(),
            Workspaces:        workspaces,
        })
    }
    ```

6.3. Implement HandleGetWorkspace:
    ```go
    func (h *Handlers) HandleGetWorkspace(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        workspace, err := h.store.GetWorkspace(id)
        if err != nil {
            respondError(w, http.StatusNotFound, "workspace not found", err.Error())
            return
        }
        
        // Get workspace info
        workspaces, _ := h.store.ListWorkspaces()
        var metadata storage.WorkspaceInfo
        for _, ws := range workspaces {
            if ws.ID == id {
                metadata = ws
                break
            }
        }
        
        respondJSON(w, http.StatusOK, WorkspaceResponse{
            Workspace: workspace,
            Metadata:  metadata,
        })
    }
    ```

6.4. Implement HandleCreateWorkspace:
    ```go
    func (h *Handlers) HandleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
        var req CreateWorkspaceRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, http.StatusBadRequest, "invalid request", err.Error())
            return
        }
        
        workspace := &domain.Workspace{
            Version: 2,
            Items:   []domain.WorkspaceItem{},
        }
        
        workspaceID := fmt.Sprintf("workspace_%d", time.Now().UnixNano())
        if err := h.store.SaveWorkspaceAs(workspaceID, workspace, req.Name); err != nil {
            respondError(w, http.StatusInternalServerError, "failed to create workspace", err.Error())
            return
        }
        
        respondJSON(w, http.StatusCreated, map[string]string{
            "workspaceId": workspaceID,
        })
    }
    ```

6.5. Implement HandleDeleteWorkspace:
    ```go
    func (h *Handlers) HandleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        if id == "default" {
            respondError(w, http.StatusBadRequest, "cannot delete default workspace", "")
            return
        }
        
        if err := h.store.DeleteWorkspace(id); err != nil {
            respondError(w, http.StatusInternalServerError, "failed to delete workspace", err.Error())
            return
        }
        
        respondJSON(w, http.StatusOK, map[string]bool{"success": true})
    }
    ```

6.6. Implement HandleActivateWorkspace:
    ```go
    func (h *Handlers) HandleActivateWorkspace(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        if err := h.store.SetActiveWorkspace(id); err != nil {
            respondError(w, http.StatusNotFound, "workspace not found", err.Error())
            return
        }
        
        respondJSON(w, http.StatusOK, map[string]bool{"success": true})
    }
    ```

6.7. Implement HandleRenameWorkspace:
    ```go
    func (h *Handlers) HandleRenameWorkspace(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        var req RenameWorkspaceRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, http.StatusBadRequest, "invalid request", err.Error())
            return
        }
        
        // Reload, rename, save
        workspace, err := h.store.GetWorkspace(id)
        if err != nil {
            respondError(w, http.StatusNotFound, "workspace not found", err.Error())
            return
        }
        
        if err := h.store.SaveWorkspaceAs(id, workspace, req.Name); err != nil {
            respondError(w, http.StatusInternalServerError, "failed to rename workspace", err.Error())
            return
        }
        
        respondJSON(w, http.StatusOK, map[string]bool{"success": true})
    }
    ```

6.8. Implement HandleImportInsomnia:
    ```go
    type ImportInsomniaRequest struct {
        FilePath string `json:"filePath"`
    }
    
    func (h *Handlers) HandleImportInsomnia(w http.ResponseWriter, r *http.Request) {
        var req ImportInsomniaRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, http.StatusBadRequest, "invalid request", err.Error())
            return
        }
        
        workspace, err := h.importService.ImportInsomnia(req.FilePath)
        if err != nil {
            respondError(w, http.StatusInternalServerError, "import failed", err.Error())
            return
        }
        
        statistics := h.importService.CalculateStatistics(workspace)
        
        respondJSON(w, http.StatusOK, ImportResult{
            Success:    true,
            Workspace:  workspace,
            Statistics: statistics,
        })
    }
    ```

6.9. Implement HandleUploadFile for multipart file uploads:
    ```go
    import (
        "io"
        "mime/multipart"
        "net/http"
        "os"
        "path/filepath"
        "time"
    )

    const (
        maxFileSize      = 10 * 1024 * 1024 // 10MB
        allowedExtension = ".yaml"
    )

    func (h *Handlers) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
        // Parse multipart form with size limit
        if err := r.ParseMultipartForm(maxFileSize); err != nil {
            respondError(w, http.StatusBadRequest, "file too large or invalid", err.Error())
            return
        }

        // Get the file from the form
        file, header, err := r.FormFile("file")
        if err != nil {
            respondError(w, http.StatusBadRequest, "no file uploaded", err.Error())
            return
        }
        defer file.Close()

        // Validate file extension
        ext := filepath.Ext(header.Filename)
        if ext != allowedExtension {
            respondError(w, http.StatusBadRequest, "invalid file type", "only .yaml files are allowed")
            return
        }

        // Create temp file path
        tempDir := os.TempDir()
        tempFilePath := filepath.Join(tempDir, "import-"+header.Filename)

        // Create destination file
        dest, err := os.Create(tempFilePath)
        if err != nil {
            respondError(w, http.StatusInternalServerError, "failed to create temp file", err.Error())
            return
        }
        defer dest.Close()

        // Copy uploaded file to destination
        if _, err := io.Copy(dest, file); err != nil {
            respondError(w, http.StatusInternalServerError, "failed to save file", err.Error())
            return
        }

        log.Printf("Uploaded file saved to: %s", tempFilePath)

        respondJSON(w, http.StatusOK, map[string]string{
            "filePath": tempFilePath,
        })

        // Schedule cleanup of temp file after response is sent
        go func() {
            time.Sleep(5 * time.Minute) // Give time for import to complete
            if err := os.Remove(tempFilePath); err != nil {
                log.Printf("Failed to remove temp file %s: %v", tempFilePath, err)
            } else {
                log.Printf("Temp file cleaned up: %s", tempFilePath)
            }
        }()
    }
    ```

6.10. Update router to add new routes:
    ```go
    func SetupRoutes(router *mux.Router, handlers *Handlers) {
        api := router.PathPrefix("/api").Subrouter()

        // Current workspace endpoints
        api.HandleFunc("/workspace", handlers.HandleGetCurrentWorkspace).Methods("GET")
        api.HandleFunc("/workspace/{id}", handlers.HandlePutWorkspace).Methods("PUT")

        // Workspace management routes
        api.HandleFunc("/workspaces", handlers.HandleListWorkspaces).Methods("GET")
        api.HandleFunc("/workspaces", handlers.HandleCreateWorkspace).Methods("POST")
        api.HandleFunc("/workspaces/{id}", handlers.HandleGetWorkspace).Methods("GET")
        api.HandleFunc("/workspaces/{id}/activate", handlers.HandleActivateWorkspace).Methods("POST")
        api.HandleFunc("/workspaces/{id}/rename", handlers.HandleRenameWorkspace).Methods("PUT")
        api.HandleFunc("/workspaces/{id}", handlers.HandleDeleteWorkspace).Methods("DELETE")

        // File upload route
        api.HandleFunc("/upload", handlers.HandleUploadFile).Methods("POST")

        // Import route
        api.HandleFunc("/import/insomnia", handlers.HandleImportInsomnia).Methods("POST")
    }
    ```

6.11. Implement HandleGetCurrentWorkspace:
    ```go
    func (h *Handlers) HandleGetCurrentWorkspace(w http.ResponseWriter, r *http.Request) {
        workspace, err := h.store.Load()
        if err != nil {
            respondError(w, http.StatusInternalServerError, "failed to load workspace", err.Error())
            return
        }

        respondJSON(w, http.StatusOK, workspace)
    }
    ```

**Testing**:
- Test listing workspaces returns all
- Test creating new workspace
- Test getting specific workspace
- Test deleting non-default workspace
- Test activating workspace
- Test renaming workspace
- Test importing Insomnia file
- Test error handling for all endpoints

---

### Step 7: Add Workspace Metadata to Domain

**Goal**: Track workspace name, created/modified timestamps

**Files**:
- Modify: `internal/domain/workspace.go`

**Output**: Workspaces include metadata for display

**Detailed Tasks**:
7.1. Add WorkspaceMetadata struct:
    ```go
    type WorkspaceMetadata struct {
        ID         string `json:"id"`
        Name       string `json:"name"`
        CreatedAt  int64  `json:"createdAt"`
        ModifiedAt int64  `json:"modifiedAt"`
    }
    ```

7.2. Update Workspace struct:
    ```go
    type Workspace struct {
        Version  int                `json:"version"`
        Items    []WorkspaceItem    `json:"items"`
        Metadata *WorkspaceMetadata `json:"metadata,omitempty"`
    }
    ```

7.3. Add helper methods:
    ```go
    func (w *Workspace) WithMetadata(id, name string) *Workspace {
        now := time.Now().UnixMilli()
        w.Metadata = &WorkspaceMetadata{
            ID:         id,
            Name:       name,
            CreatedAt:  now,
            ModifiedAt: now,
        }
        return w
    }
    
    func (w *Workspace) Touch() {
        if w.Metadata != nil {
            w.Metadata.ModifiedAt = time.Now().UnixMilli()
        }
    }
    ```

7.4. Update MultiWorkspaceStore to track metadata on save:
    - When saving workspace, read existing metadata or create new
    - Update ModifiedAt timestamp
    - Preserve CreatedAt from existing metadata

**Testing**:
- Test metadata is saved with workspace
- test metadata is updated on modification
- Test metadata is returned in workspace list
- Test metadata is preserved across saves

---

### Step 8: Update Frontend Types

**Goal**: Add TypeScript types for workspaces and import

**Files**:
- Modify: `web/src/types.ts`

**Output**: TypeScript types for workspace operations

**Detailed Tasks**:
8.1. Add workspace-related types:
    ```typescript
    export interface WorkspaceInfo {
      id: string;
      name: string;
      path: string;
      createdAt: number;
      modifiedAt: number;
    }
    
    export interface WorkspaceListResponse {
      activeWorkspaceId: string;
      workspaces: WorkspaceInfo[];
    }
    
    export interface WorkspaceResponse {
      workspace: Workspace;
      metadata: WorkspaceInfo;
    }
    
    export interface ImportResult {
      success: boolean;
      workspace?: Workspace;
      statistics?: {
        totalItems: number;
        requests: number;
        folders: number;
        headers: number;
        bodySize: number;
      };
      errors?: Array<{ item: string; message: string }>;
    }
    
    export interface CreateWorkspaceRequest {
      name: string;
    }
    
    export interface RenameWorkspaceRequest {
      name: string;
    }
    ```

8.2. Update AppState to include workspace info:
    ```typescript
    interface AppState {
      // ... existing fields ...
      activeWorkspaceId: string | null;
      workspaces: WorkspaceInfo[];
    }
    ```

**Testing**:
- Verify TypeScript compilation
- Verify all types are used correctly in components

---

### Step 9: Create Import Modal Component

**Goal**: UI for importing Insomnia files

**Files**:
- Create: `web/src/components/ImportModal.tsx`

**Output**: Complete import modal UI

**Detailed Tasks**:
9.1. Create component structure:
    ```tsx
    import React, { useState, useRef } from 'react';
    import { ImportResult } from '../types';
    
    interface ImportModalProps {
      isOpen: boolean;
      onClose: () => void;
      onImportComplete: (workspaceId: string) => void;
    }
    
    const ImportModal: React.FC<ImportModalProps> = ({ isOpen, onClose, onImportComplete }) => {
      const [file, setFile] = useState<File | null>(null);
      const [isImporting, setIsImporting] = useState(false);
      const [result, setResult] = useState<ImportResult | null>(null);
      const [error, setError] = useState<string | null>(null);
      const fileInputRef = useRef<HTMLInputElement>(null);
      
      // ... implementation
    };
    
    export default ImportModal;
    ```

9.2. Implement file selection:
    ```tsx
    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (selectedFile) {
        if (!selectedFile.name.endsWith('.yaml') && !selectedFile.name.endsWith('.yml')) {
          setError('Please select a YAML file');
          return;
        }
        setFile(selectedFile);
        setError(null);
      }
    };
    ```

9.3. Implement drag and drop:
    ```tsx
    const handleDrop = (e: React.DragEvent) => {
      e.preventDefault();
      const droppedFile = e.dataTransfer.files[0];
      if (droppedFile) {
        setFile(droppedFile);
        setError(null);
      }
    };
    
    const handleDragOver = (e: React.DragEvent) => {
      e.preventDefault();
    };
    ```

9.4. Implement import handler:
    ```tsx
    const handleImport = async () => {
      if (!file) return;
      
      setIsImporting(true);
      setError(null);
      
      try {
        // Upload file first, then import
        const formData = new FormData();
        formData.append('file', file);
        
        // Save file to server
        const saveRes = await fetch('/api/upload', {
          method: 'POST',
          body: formData,
        });
        
        if (!saveRes.ok) throw new Error('Failed to save file');
        const { filePath } = await saveRes.json();
        
        // Import from file
        const importRes = await fetch('/api/import/insomnia', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ filePath }),
        });
        
        if (!importRes.ok) throw new Error('Import failed');

        const result = await importRes.json();
        setResult(result);

        // Safely extract workspace id with nil checks
        const workspaceId = result.workspace?.metadata?.id;
        if (workspaceId) {
          onImportComplete(workspaceId);
        } else {
          console.warn('Imported workspace missing metadata.id');
        }
      } catch (err: any) {
        setError(err.message);
      } finally {
        setIsImporting(false);
      }
    };
    ```

9.5. Implement UI with file drop zone:
    ```tsx
    return (
      <Modal isOpen={isOpen} onClose={onClose}>
        <h2>Import from Insomnia</h2>
        
        {!result ? (
          <>
            <div 
              className="drop-zone"
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onClick={() => fileInputRef.current?.click()}
            >
              {file ? (
                <div>
                  <p>Selected: {file.name}</p>
                  <p>Size: {(file.size / 1024).toFixed(2)} KB</p>
                </div>
              ) : (
                <p>Drop YAML file here or click to browse</p>
              )}
              <input
                ref={fileInputRef}
                type="file"
                accept=".yaml,.yml"
                onChange={handleFileSelect}
                style={{ display: 'none' }}
              />
            </div>
            
            {error && <p className="error">{error}</p>}
            
            <button 
              onClick={handleImport}
              disabled={!file || isImporting}
            >
              {isImporting ? 'Importing...' : 'Import'}
            </button>
          </>
        ) : (
          <div className="import-result">
            <h3>Import Successful!</h3>
            <ul>
              <li>Folders: {result.statistics?.folders}</li>
              <li>Requests: {result.statistics?.requests}</li>
              <li>Total Items: {result.statistics?.totalItems}</li>
            </ul>
            <button onClick={onClose}>Close</button>
          </div>
        )}
      </Modal>
    );
    ```

9.6. Add file upload endpoint:
    - Need to add `POST /api/upload` handler to save uploaded files temporarily

**Testing**:
- Test file picker works
- Test drag and drop works
- Test import success shows statistics
- Test error handling shows user-friendly messages
- Test loading state during import

---

### Step 10: Add Workspace Switcher to Sidebar

**Goal**: UI for switching and managing workspaces

**Files**:
- Modify: `web/src/components/Sidebar.tsx`

**Output**: Workspace switcher with rename/delete support

**Detailed Tasks**:
10.1. Add workspace switcher state:
    ```tsx
    const [showWorkspaceMenu, setShowWorkspaceMenu] = useState(false);
    const [renameWorkspaceId, setRenameWorkspaceId] = useState<string | null>(null);
    const [newWorkspaceName, setNewWorkspaceName] = useState('');
    ```

10.2. Add workspace switcher dropdown UI:
    ```tsx
    <div className="workspace-switcher">
      <button 
        className="workspace-button"
        onClick={() => setShowWorkspaceMenu(!showWorkspaceMenu)}
      >
        {currentWorkspaceName}
        <ChevronDownIcon />
      </button>
      
      {showWorkspaceMenu && (
        <div className="workspace-dropdown">
          {workspaces.map(ws => (
            <div 
              key={ws.id}
              className={`workspace-item ${ws.id === activeWorkspaceId ? 'active' : ''}`}
              onClick={() => switchToWorkspace(ws.id)}
            >
              {renameWorkspaceId === ws.id ? (
                <input
                  type="text"
                  value={newWorkspaceName}
                  onChange={(e) => setNewWorkspaceName(e.target.value)}
                  onBlur={() => renameWorkspace(ws.id, newWorkspaceName)}
                  onKeyDown={(e) => e.key === 'Enter' && renameWorkspace(ws.id, newWorkspaceName)}
                  autoFocus
                />
              ) : (
                <>
                  <span>{ws.name}</span>
                  <button onClick={(e) => { e.stopPropagation(); setRenameWorkspaceId(ws.id); setNewWorkspaceName(ws.name); }}>
                    ✏️
                  </button>
                  {ws.id !== 'default' && (
                    <button onClick={(e) => { e.stopPropagation(); deleteWorkspace(ws.id); }}>
                      🗑️
                    </button>
                  )}
                </>
              )}
            </div>
          ))}
          
          <div className="workspace-actions">
            <button onClick={() => openImportModal()}>
              📥 Import Insomnia
            </button>
            <button onClick={() => createWorkspace()}>
              ➕ Create New
            </button>
          </div>
        </div>
      )}
    </div>
    ```

10.3. Implement workspace operations:
    ```tsx
    const switchToWorkspace = async (id: string) => {
      await fetch(`/api/workspaces/${id}/activate`, { method: 'POST' });
      setShowWorkspaceMenu(false);
      // Reload workspace
      await loadWorkspace();
    };
    
    const renameWorkspace = async (id: string, newName: string) => {
      await fetch(`/api/workspaces/${id}/rename`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newName }),
      });
      setRenameWorkspaceId(null);
      await loadWorkspaces();
    };
    
    const deleteWorkspace = async (id: string) => {
      if (!confirm('Are you sure you want to delete this workspace?')) return;
      await fetch(`/api/workspaces/${id}`, { method: 'DELETE' });
      if (activeWorkspaceId === id) {
        await switchToWorkspace('default');
      } else {
        await loadWorkspaces();
      }
    };
    
    const createWorkspace = async () => {
      const name = prompt('Enter workspace name:', 'New Workspace');
      if (!name) return;
      
      const res = await fetch('/api/workspaces', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      });
      
      const { workspaceId } = await res.json();
      await switchToWorkspace(workspaceId);
    };
    ```

10.4. Update App.tsx to provide workspace data:
    ```tsx
    // Pass workspace info to Sidebar
    <Sidebar
      items={items}
      activeId={activeId}
      workspaces={workspaces}
      activeWorkspaceId={activeWorkspaceId}
      onSwitchWorkspace={switchWorkspace}
      // ... other props
    />
    ```

**Testing**:
- Test workspace dropdown opens
- Test clicking workspace switches to it
- Test rename button starts inline rename
- Test delete button removes workspace
- Test create new workspace
- Test import option opens modal

---

### Step 11: Handle Workspace Switching in App

**Goal**: Connect workspace switching to backend API

**Files**:
- Modify: `web/src/App.tsx`

**Output**: Full workspace switching integration

**Detailed Tasks**:
11.1. Add workspace state:
    ```tsx
    const [activeWorkspaceId, setActiveWorkspaceId] = useState<string | null>(null);
    const [workspaces, setWorkspaces] = useState<WorkspaceInfo[]>([]);
    ```

11.2. Add loadWorkspaces function:
    ```tsx
    const loadWorkspaces = async () => {
      try {
        const res = await fetch('/api/workspaces');
        if (!res.ok) throw new Error('Failed to load workspaces');
        const data = await res.json();
        setWorkspaces(data.workspaces);
        setActiveWorkspaceId(data.activeWorkspaceId);
      } catch (err) {
        console.error('Failed to load workspaces:', err);
      }
    };
    ```

11.3. Load workspaces on mount:
    ```tsx
    useEffect(() => {
      loadWorkspaces();
      loadWorkspace(); // Load current workspace
    }, []);
    ```

11.4. Add switchWorkspace function:
    ```tsx
    const switchWorkspace = async (id: string) => {
      try {
        // Save current workspace first
        await saveWorkspace();

        // Switch to new workspace
        const res = await fetch(`/api/workspaces/${id}/activate`, {
          method: 'POST',
        });

        if (!res.ok) throw new Error('Failed to switch workspace');

        // Reload workspace data
        await loadWorkspace();
        await loadWorkspaces();
      } catch (err) {
        console.error('Failed to switch workspace:', err);
      }
    };
    ```

11.5. Add import callback:
    ```tsx
    const handleImportComplete = async (newWorkspaceId: string) => {
      await loadWorkspaces();
      await loadWorkspace();
      setActiveWorkspaceId(newWorkspaceId);
    };
    ```

11.6. Update saveWorkspace to use active workspace:
    ```tsx
    const saveWorkspace = async () => {
      if (!activeWorkspaceId) return;

      try {
        const workspace = buildWorkspaceFromItems();
        const res = await fetch(`/api/workspace/${activeWorkspaceId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(workspace),
        });

        if (!res.ok) throw new Error('Failed to save workspace');
      } catch (err) {
        console.error('Failed to save workspace:', err);
      }
    };
    ```

11.7. Pass workspace data to Sidebar:
    ```tsx
    <Sidebar
      items={items}
      activeId={activeId}
      onItemSelect={setActiveId}
      onItemsChange={setItems}
      onExecute={executeRequest}
      onOpenSettings={() => setIsSettingsOpen(true)}
      workspaces={workspaces}
      activeWorkspaceId={activeWorkspaceId}
      onSwitchWorkspace={switchWorkspace}
      onOpenImport={() => setIsImportOpen(true)}
    />
    ```

**Testing**:
- Test workspaces load on page load
- Test switching workspaces saves current first
- Test import callback updates workspace list
- Test debounced auto-save works with workspace switching

---

### Step 12: Add Import Button to Sidebar

**Goal**: Easy access to import functionality

**Files**:
- Modify: `web/src/components/Sidebar.tsx`

**Output**: Easy-to-find import button

**Detailed Tasks**:
12.1. Add import button to Sidebar header:
    ```tsx
    <header className="sidebar-header">
      <h1>Constrictor</h1>
      <div className="sidebar-actions">
        <button 
          className="import-button"
          onClick={onOpenImport}
          title="Import from Insomnia"
        >
          <UploadIcon />
          Import
        </button>
        <button 
          className="settings-button"
          onClick={onOpenSettings}
          title="Settings"
        >
          <SettingsIcon />
        </button>
      </div>
    </header>
    ```

12.2. Update props interface:
    ```typescript
    interface SidebarProps {
      // ... existing props
      onOpenImport: () => void;
    }
    ```

12.3. Add ImportModal to App.tsx:
    ```tsx
    import ImportModal from './components/ImportModal';
    
    const App: React.FC = () => {
      // ... existing state
      
      return (
        <div className="app">
          <Sidebar
            // ... existing props
            onOpenImport={() => setIsImportOpen(true)}
          />
          <RequestEditor
            // ... existing props
          />
          <ResponseViewer
            // ... existing props
          />
          <SettingsModal
            isOpen={isSettingsOpen}
            onClose={() => setIsSettingsOpen(false)}
            settings={settings}
            onSettingsChange={setSettings}
          />
          <ImportModal
            isOpen={isImportOpen}
            onClose={() => setIsImportOpen(false)}
            onImportComplete={handleImportComplete}
          />
        </div>
      );
    };
    ```

12.4. Add simple icons (SVG or use library):
    - Use inline SVG icons for import, settings, etc.

**Testing**:
- Test import button is visible
- Test clicking opens ImportModal
- Test modal closes on backdrop click
- Test modal closes on successful import

---

## Files Summary

### Created Files
```text
internal/import/
  models.go              # Insomnia format models
  service.go             # Import service facade
  importers/
    insomnia.go          # Insomnia YAML importer

internal/storage/
  workspaces.go          # MultiWorkspaceStore implementation
  migrate.go             # v1 → v2 migration

web/src/components/
  ImportModal.tsx        # Import UI component
```

### Modified Files
```text
internal/
  domain/
    workspace.go         # Add metadata fields
  storage/
    store.go             # Add WorkspaceStore interface methods
  httpapi/
    handlers.go          # Add workspace + import handlers
    router.go            # Add workspace routes

web/src/
  types.ts               # Add workspace types
  components/
    Sidebar.tsx          # Add workspace switcher + import button
  App.tsx                # Handle workspace switching
```

---

## Testing Strategy

### Unit Tests
- Importer conversion logic (folders, auth, body types)
- Workspace metadata generation
- Form data parsing
- Storage registry operations

### Integration Tests
- MultiWorkspaceStore operations
- Migration from v1 to v2
- Full import flow with actual YAML file
- API endpoint responses

### API Tests
- List workspaces endpoint
- Switch workspace endpoint
- Import endpoint with validation
- Rename/delete operations

### UI Tests
- Import modal file selection
- Workspace switcher dropdown
- Rename workflow
- Delete workflow
- Import success/error display

### Manual E2E Test
1. Start with default workspace (1 request)
2. Click "Import Insomnia"
3. Select `Insomnia_2026-01-22.yaml`
4. Verify current workspace auto-saved
5. Verify new workspace created with timestamp name
6. Verify folder hierarchy via parentId
7. Verify auth headers preserved
8. Switch between workspaces
9. Rename imported workspace
10. Delete a workspace (not active)

---

## Edge Cases

| Case | Handling |
|------|----------|
| File not found | User-friendly error message in modal |
| Invalid YAML | Parse error with line number, show in modal |
| Missing fields | Default values or skip item, log warning |
| Duplicate folder names | Accept (IDs are unique) |
| Invalid URLs | Log warning, use placeholder URL, continue import |
| Large files | YAML is text-based, should handle 70K lines fine |
| Delete active workspace | Prevent or switch to default first |
| Empty import | Create empty workspace, show success |
| Import while unsaved changes | Auto-save first, then import |

---

## Progress Tracking

Use PROGRESS_IMPORT_API_COLLECTIONS.md for tracking each step.

**Entry Template**:
```markdown
## [YYYY-MM-DD HH:MM] - Step Name

**Completed:**
- What was done

**Files Changed:**
- path/to/file1.go
- path/to/file2.ts

**How to Verify:**
- Command to run or test to execute

**Next Steps:**
- What comes next

**Blockers/Notes:**
- Any issues or decisions made
```

**Steps**:
1. Step 1: Extend WorkspaceStore Interface
2. Step 2: Create Workspace Registry Migration
3. Step 3: Define Insomnia Format Models
4. Step 4: Implement Insomnia Importer
5. Step 5: Create Import Service Facade
6. Step 6: Add Workspace Management API
7. Step 7: Add Workspace Metadata to Domain
8. Step 8: Update Frontend Types
9. Step 9: Create Import Modal Component
10. Step 10: Add Workspace Switcher to Sidebar
11. Step 11: Handle Workspace Switching in App
12. Step 12: Add Import Button to Sidebar

---

## Plan Ready for Execution

This plan is ready to be executed by agents. Each step is independent and can be implemented in order. The implementation order can be adjusted based on dependencies (e.g., Step 1-2 must come before Step 5-6, but frontend steps can run in parallel with backend steps).

**Estimated Files**: 8 new, 6 modified
**Estimated Lines**: ~2000 Go, ~600 TypeScript
**Complexity**: Medium - standard patterns with careful edge case handling
