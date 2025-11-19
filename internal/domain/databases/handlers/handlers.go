package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

type DatabaseHandler struct {
	dbService        *service.DatabaseService
	containerManager manager.ContainerManager
	validator        *validator.Validate
}

func NewDatabaseHandler(dbService *service.DatabaseService, containerManager manager.ContainerManager) *DatabaseHandler {
	return &DatabaseHandler{
		dbService:        dbService,
		containerManager: containerManager,
		validator:        validator.New(),
	}
}

// DatabaseResponse represents a database in API responses
type DatabaseResponse struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	Type             databases.DatabaseType   `json:"type"`
	ProjectID        string                   `json:"project_id"`
	EnvironmentID    string                   `json:"environment_id"`
	Config           databases.DatabaseConfig `json:"config"`
	Status           databases.DatabaseStatus `json:"status"`
	ConnectionString string                   `json:"connection_string"`
	Ports            map[string]int           `json:"ports"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

type CreateDatabaseRequest struct {
	Name          string                    `json:"name" validate:"required,min=1,max=63"`
	Description   string                    `json:"description,omitempty"`
	Type          databases.DatabaseType    `json:"type" validate:"required"`
	EnvironmentID string                    `json:"environment_id" validate:"required,uuid"`
	Config        *databases.DatabaseConfig `json:"config,omitempty"`
}

type UpdateDatabaseRequest struct {
	Description *string                   `json:"description,omitempty"`
	Config      *databases.DatabaseConfig `json:"config,omitempty"`
}

type DatabaseListItem struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	Type          databases.DatabaseType   `json:"type"`
	ProjectID     string                   `json:"project_id"`
	EnvironmentID string                   `json:"environment_id"`
	Status        databases.DatabaseStatus `json:"status"`
	CreatedAt     string                   `json:"created_at"`
}

type ListDatabasesResponse struct {
	Databases []DatabaseListItem `json:"databases"`
}

type DatabaseActionRequest struct {
	Action string `json:"action" validate:"required,oneof=start stop"`
}

type DatabaseTypesResponse struct {
	Types []databases.DatabaseType `json:"types"`
}

func maskDatabaseConfig(config databases.DatabaseConfig) databases.DatabaseConfig {
	masked := config

	if config.PostgreSQL != nil {
		pgCopy := *config.PostgreSQL
		pgCopy.Password = "********"
		masked.PostgreSQL = &pgCopy
	}

	if config.MySQL != nil {
		mysqlCopy := *config.MySQL
		mysqlCopy.Password = "********"
		mysqlCopy.RootPassword = "********"
		masked.MySQL = &mysqlCopy
	}

	if config.MariaDB != nil {
		mariadbCopy := *config.MariaDB
		mariadbCopy.Password = "********"
		mariadbCopy.RootPassword = "********"
		masked.MariaDB = &mariadbCopy
	}

	if config.Redis != nil {
		redisCopy := *config.Redis
		if redisCopy.Password != "" {
			redisCopy.Password = "********"
		}
		masked.Redis = &redisCopy
	}

	if config.KeyDB != nil {
		keydbCopy := *config.KeyDB
		if keydbCopy.Password != "" {
			keydbCopy.Password = "********"
		}
		masked.KeyDB = &keydbCopy
	}

	if config.Dragonfly != nil {
		dragonflyCopy := *config.Dragonfly
		if dragonflyCopy.Password != "" {
			dragonflyCopy.Password = "********"
		}
		masked.Dragonfly = &dragonflyCopy
	}

	if config.MongoDB != nil {
		mongodbCopy := *config.MongoDB
		mongodbCopy.Password = "********"
		masked.MongoDB = &mongodbCopy
	}

	if config.ClickHouse != nil {
		clickhouseCopy := *config.ClickHouse
		clickhouseCopy.Password = "********"
		masked.ClickHouse = &clickhouseCopy
	}

	return masked
}

// CreateDatabase creates a new database in a project
func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req CreateDatabaseRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// Parse environment ID
	environmentID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
		return
	}

	// Validate config if provided
	if req.Config != nil {
		if err := h.dbService.ValidateDatabaseConfig(req.Type, *req.Config); err != nil {
			utils.SendError(w, http.StatusBadRequest, "invalid_config", "Invalid database configuration: "+err.Error())
			return
		}
	}

	cmd := service.CreateDatabaseCommand{
		Name:          req.Name,
		Description:   req.Description,
		Type:          req.Type,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		Config:        req.Config,
	}

	database, err := h.dbService.CreateDatabase(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "creation_failed", "Failed to create database: "+err.Error())
		return
	}

	response := DatabaseResponse{
		ID:               database.ID().String(),
		Name:             database.Name().String(),
		Description:      database.Description(),
		Type:             database.Type(),
		ProjectID:        database.ProjectID().String(),
		EnvironmentID:    database.EnvironmentID().String(),
		Config:           maskDatabaseConfig(database.Config()),
		Status:           database.Status(),
		ConnectionString: database.ConnectionString(),
		Ports:            database.Ports(),
		CreatedAt:        database.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        database.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

// GetDatabase retrieves a specific database
func (h *DatabaseHandler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	// Get database ID from URL
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	// Get project ID from URL for validation
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	// Verify the database belongs to the project
	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	response := DatabaseResponse{
		ID:               database.ID().String(),
		Name:             database.Name().String(),
		Description:      database.Description(),
		Type:             database.Type(),
		ProjectID:        database.ProjectID().String(),
		EnvironmentID:    database.EnvironmentID().String(),
		Config:           maskDatabaseConfig(database.Config()),
		Status:           database.Status(),
		ConnectionString: database.ConnectionString(),
		Ports:            database.Ports(),
		CreatedAt:        database.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        database.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// ListDatabases lists all databases in a project
func (h *DatabaseHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	environmentIDStr := r.URL.Query().Get("environment_id")
	var databases []*databases.Database

	if environmentIDStr != "" {
		environmentID, parseErr := uuid.Parse(environmentIDStr)
		if parseErr != nil {
			utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
			return
		}
		databases, err = h.dbService.ListDatabasesByEnvironment(r.Context(), projectID, environmentID)
	} else {
		databases, err = h.dbService.ListDatabases(r.Context(), projectID)
	}

	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list databases: "+err.Error())
		return
	}

	items := make([]DatabaseListItem, len(databases))
	for i, db := range databases {
		items[i] = DatabaseListItem{
			ID:            db.ID().String(),
			Name:          db.Name().String(),
			Description:   db.Description(),
			Type:          db.Type(),
			ProjectID:     db.ProjectID().String(),
			EnvironmentID: db.EnvironmentID().String(),
			Status:        db.Status(),
			CreatedAt:     db.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := ListDatabasesResponse{
		Databases: items,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// UpdateDatabase updates an existing database
func (h *DatabaseHandler) UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	var req UpdateDatabaseRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get database ID from URL
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	// Get project ID from URL for validation
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the database exists and belongs to the project
	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	// Validate config if provided
	if req.Config != nil {
		if err := h.dbService.ValidateDatabaseConfig(database.Type(), *req.Config); err != nil {
			utils.SendError(w, http.StatusBadRequest, "invalid_config", "Invalid database configuration: "+err.Error())
			return
		}
	}

	cmd := service.UpdateDatabaseCommand{
		ID:          databaseID,
		Description: req.Description,
		Config:      req.Config,
	}

	updatedDatabase, err := h.dbService.UpdateDatabase(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update database: "+err.Error())
		return
	}

	response := DatabaseResponse{
		ID:               updatedDatabase.ID().String(),
		Name:             updatedDatabase.Name().String(),
		Description:      updatedDatabase.Description(),
		Type:             updatedDatabase.Type(),
		ProjectID:        updatedDatabase.ProjectID().String(),
		EnvironmentID:    updatedDatabase.EnvironmentID().String(),
		Config:           maskDatabaseConfig(updatedDatabase.Config()),
		Status:           updatedDatabase.Status(),
		ConnectionString: updatedDatabase.ConnectionString(),
		Ports:            updatedDatabase.Ports(),
		CreatedAt:        updatedDatabase.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedDatabase.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// DeleteDatabase deletes a database
func (h *DatabaseHandler) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	// Get database ID from URL
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	// Get project ID from URL for validation
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the database exists and belongs to the project
	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	if err := h.dbService.DeleteDatabase(r.Context(), databaseID); err != nil {
		utils.SendError(w, http.StatusBadRequest, "deletion_failed", "Failed to delete database: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusNoContent, nil)
}

// DatabaseAction handles start/stop actions for databases
func (h *DatabaseHandler) DatabaseAction(w http.ResponseWriter, r *http.Request) {
	var req DatabaseActionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get database ID from URL
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	// Get project ID from URL for validation
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the database exists and belongs to the project
	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	// Perform the requested action
	switch req.Action {
	case "start":
		if err := h.dbService.StartDatabase(r.Context(), databaseID); err != nil {
			utils.SendError(w, http.StatusBadRequest, "start_failed", "Failed to start database: "+err.Error())
			return
		}
	case "stop":
		if err := h.dbService.StopDatabase(r.Context(), databaseID); err != nil {
			utils.SendError(w, http.StatusBadRequest, "stop_failed", "Failed to stop database: "+err.Error())
			return
		}
	default:
		utils.SendError(w, http.StatusBadRequest, "invalid_action", "Invalid action. Must be 'start' or 'stop'")
		return
	}

	// Return updated database status
	updatedDatabase, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "status_update_failed", "Action completed but failed to get updated status")
		return
	}

	response := DatabaseResponse{
		ID:               updatedDatabase.ID().String(),
		Name:             updatedDatabase.Name().String(),
		Description:      updatedDatabase.Description(),
		Type:             updatedDatabase.Type(),
		ProjectID:        updatedDatabase.ProjectID().String(),
		EnvironmentID:    updatedDatabase.EnvironmentID().String(),
		Config:           maskDatabaseConfig(updatedDatabase.Config()),
		Status:           updatedDatabase.Status(),
		ConnectionString: updatedDatabase.ConnectionString(),
		Ports:            updatedDatabase.Ports(),
		CreatedAt:        updatedDatabase.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedDatabase.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// GetDatabaseTypes returns all supported database types
func (h *DatabaseHandler) GetDatabaseTypes(w http.ResponseWriter, r *http.Request) {
	types := h.dbService.GetSupportedDatabaseTypes()

	response := DatabaseTypesResponse{
		Types: types,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// GetDatabaseByName retrieves a database by name within a project
func (h *DatabaseHandler) GetDatabaseByName(w http.ResponseWriter, r *http.Request) {
	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// Get database name from URL
	databaseName := chi.URLParam(r, "database_name")
	if databaseName == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_database_name", "Database name is required")
		return
	}

	database, err := h.dbService.GetDatabaseByName(r.Context(), projectID, databaseName)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	response := DatabaseResponse{
		ID:               database.ID().String(),
		Name:             database.Name().String(),
		Description:      database.Description(),
		Type:             database.Type(),
		ProjectID:        database.ProjectID().String(),
		EnvironmentID:    database.EnvironmentID().String(),
		Config:           maskDatabaseConfig(database.Config()),
		Status:           database.Status(),
		ConnectionString: database.ConnectionString(),
		Ports:            database.Ports(),
		CreatedAt:        database.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        database.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// GetDefaultDatabaseConfig returns default configuration for a database type
func (h *DatabaseHandler) GetDefaultDatabaseConfig(w http.ResponseWriter, r *http.Request) {
	// Get database type from URL
	dbTypeStr := chi.URLParam(r, "type")
	dbType := databases.DatabaseType(dbTypeStr)

	// Validate database type
	supportedTypes := h.dbService.GetSupportedDatabaseTypes()
	var isSupported bool
	for _, supportedType := range supportedTypes {
		if supportedType == dbType {
			isSupported = true
			break
		}
	}

	if !isSupported {
		utils.SendError(w, http.StatusBadRequest, "unsupported_type", fmt.Sprintf("Unsupported database type: %s", dbType))
		return
	}

	// Generate default config
	var config databases.DatabaseConfig
	switch dbType {
	case databases.DatabaseTypePostgreSQL:
		config = databases.DatabaseConfig{
			Type:       dbType,
			PostgreSQL: databases.DefaultPostgreSQLConfig(),
		}
	case databases.DatabaseTypeMySQL:
		config = databases.DatabaseConfig{
			Type:  dbType,
			MySQL: databases.DefaultMySQLConfig(),
		}
	case databases.DatabaseTypeMariaDB:
		config = databases.DatabaseConfig{
			Type:    dbType,
			MariaDB: databases.DefaultMariaDBConfig(),
		}
	case databases.DatabaseTypeRedis:
		config = databases.DatabaseConfig{
			Type:  dbType,
			Redis: databases.DefaultRedisConfig(),
		}
	case databases.DatabaseTypeKeyDB:
		config = databases.DatabaseConfig{
			Type:  dbType,
			KeyDB: databases.DefaultKeyDBConfig(),
		}
	case databases.DatabaseTypeDragonfly:
		config = databases.DatabaseConfig{
			Type:      dbType,
			Dragonfly: databases.DefaultDragonflyConfig(),
		}
	case databases.DatabaseTypeMongoDB:
		config = databases.DatabaseConfig{
			Type:    dbType,
			MongoDB: databases.DefaultMongoDBConfig(),
		}
	case databases.DatabaseTypeClickHouse:
		config = databases.DatabaseConfig{
			Type:       dbType,
			ClickHouse: databases.DefaultClickHouseConfig(),
		}
	}

	utils.SendJSON(w, http.StatusOK, maskDatabaseConfig(config))
}

// GetDatabaseLogs streams logs from a database container
func (h *DatabaseHandler) GetDatabaseLogs(w http.ResponseWriter, r *http.Request) {
	// Get database ID from URL
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	// Get project ID from URL for validation
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the database exists and belongs to the project
	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	// Check if database has a container
	if database.ContainerID() == "" {
		utils.SendError(w, http.StatusBadRequest, "no_container", "Database has no running container")
		return
	}

	// Get follow parameter (default to false)
	follow := r.URL.Query().Get("follow") == "true"

	// Stream logs from container
	logStream, err := h.containerManager.StreamLogs(r.Context(), database.ContainerID(), follow)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "logs_failed", "Failed to get container logs: "+err.Error())
		return
	}
	defer func() {
		_ = logStream.Close()
	}()

	// Set appropriate headers for streaming
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if follow {
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	// Copy logs to response
	_, err = io.Copy(w, logStream)
	if err != nil {
		// Log error but don't send HTTP error since we've already started writing
		// This is common when client disconnects from a streaming endpoint
		return
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Rows uint   `json:"rows,omitempty"`
	Cols uint   `json:"cols,omitempty"`
}

func (h *DatabaseHandler) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database_id", "Invalid database ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found")
		return
	}

	if database.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "database_not_found", "Database not found in project")
		return
	}

	if database.ContainerID() == "" {
		utils.SendError(w, http.StatusBadRequest, "no_container", "Database has no running container")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	resizeChan := make(chan manager.TerminalSize, 10)

	errChan := make(chan error, 3)

	go func() {
		cmd := []string{"/bin/sh"}
		if database.Type() == databases.DatabaseTypePostgreSQL {
			cmd = []string{"/bin/bash"}
		}
		err := h.containerManager.ExecInteractive(r.Context(), database.ContainerID(), cmd, stdinReader, stdoutWriter, stderrWriter, resizeChan)
		errChan <- err
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stdoutReader.Read(buf)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				}
				return
			}
			if n > 0 {
				msg := TerminalMessage{
					Type: "output",
					Data: string(buf[:n]),
				}
				if err := conn.WriteJSON(msg); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stderrReader.Read(buf)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				}
				return
			}
			if n > 0 {
				msg := TerminalMessage{
					Type: "output",
					Data: string(buf[:n]),
				}
				if err := conn.WriteJSON(msg); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	go func() {
		for {
			var msg TerminalMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					errChan <- err
				}
				return
			}

			switch msg.Type {
			case "input":
				if _, err := stdinWriter.Write([]byte(msg.Data)); err != nil {
					errChan <- err
					return
				}
			case "resize":
				resizeChan <- manager.TerminalSize{
					Height: msg.Rows,
					Width:  msg.Cols,
				}
			}
		}
	}()

	<-errChan
	close(resizeChan)
	stdinWriter.Close()
	stdoutWriter.Close()
	stderrWriter.Close()
}
