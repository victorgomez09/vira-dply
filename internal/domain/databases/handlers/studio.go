package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
	"github.com/mikrocloud/mikrocloud/pkg/containers/database/studio"
)

type StudioHandler struct {
	dbService     *service.DatabaseService
	clientFactory *studio.ClientFactory
	validator     *validator.Validate
}

func NewStudioHandler(dbService *service.DatabaseService) *StudioHandler {
	return &StudioHandler{
		dbService:     dbService,
		clientFactory: studio.NewClientFactory(),
		validator:     validator.New(),
	}
}

type ListTablesResponse struct {
	Tables []string `json:"tables"`
}

type ListSchemasResponse struct {
	Schemas []string `json:"schemas"`
}

type ExecuteQueryRequest struct {
	Query  string `json:"query" validate:"required"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type GetTableDataRequest struct {
	Limit   int             `json:"limit,omitempty"`
	Offset  int             `json:"offset,omitempty"`
	Filters []studio.Filter `json:"filters,omitempty"`
	Sorts   []studio.Sort   `json:"sorts,omitempty"`
}

type InsertRowRequest struct {
	Data map[string]any `json:"data" validate:"required"`
}

type UpdateRowRequest struct {
	PrimaryKey map[string]any `json:"primary_key" validate:"required"`
	Data       map[string]any `json:"data" validate:"required"`
}

type DeleteRowRequest struct {
	PrimaryKey map[string]any `json:"primary_key" validate:"required"`
}

type DatabaseInfoResponse struct {
	Version string `json:"version"`
	Size    int64  `json:"size"`
}

func (h *StudioHandler) validateDatabaseAccess(r *http.Request) (*databases.Database, error) {
	databaseIDStr := chi.URLParam(r, "database_id")
	databaseID, err := databases.DatabaseIDFromString(databaseIDStr)
	if err != nil {
		return nil, err
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, err
	}

	database, err := h.dbService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		return nil, err
	}

	if database.ProjectID() != projectID {
		return nil, err
	}

	if !h.clientFactory.SupportsStudio(database.Type()) {
		return nil, err
	}

	return database, nil
}

func (h *StudioHandler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	schemas, err := client.ListSchemas(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list schemas: "+err.Error())
		return
	}

	response := ListSchemasResponse{Schemas: schemas}
	utils.SendJSON(w, http.StatusOK, response)
}

func (h *StudioHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	schema := r.URL.Query().Get("schema")

	tables, err := client.ListTables(r.Context(), schema)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list tables: "+err.Error())
		return
	}

	response := ListTablesResponse{Tables: tables}
	utils.SendJSON(w, http.StatusOK, response)
}

func (h *StudioHandler) GetTableSchema(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	tableName := chi.URLParam(r, "table_name")
	schema := r.URL.Query().Get("schema")

	tableSchema, err := client.GetTableSchema(r.Context(), tableName, schema)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "schema_failed", "Failed to get table schema: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, tableSchema)
}

func (h *StudioHandler) GetTableData(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	tableName := chi.URLParam(r, "table_name")
	schema := r.URL.Query().Get("schema")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	var req GetTableDataRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Limit > 0 {
				limit = req.Limit
			}
			if req.Offset > 0 {
				offset = req.Offset
			}
		}
	}

	opts := studio.TableDataOptions{
		Limit:   limit,
		Offset:  offset,
		Filters: req.Filters,
		Sorts:   req.Sorts,
	}

	result, err := client.GetTableData(r.Context(), tableName, schema, opts)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "data_failed", "Failed to get table data: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, result)
}

func (h *StudioHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	var req ExecuteQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	if req.Limit == 0 {
		req.Limit = 1000
	}

	opts := studio.QueryOptions{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	result, err := client.ExecuteQuery(r.Context(), req.Query, opts)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "query_failed", "Failed to execute query: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, result)
}

func (h *StudioHandler) InsertRow(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	var req InsertRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	tableName := chi.URLParam(r, "table_name")
	schema := r.URL.Query().Get("schema")

	if err := client.InsertRow(r.Context(), tableName, schema, req.Data); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "insert_failed", "Failed to insert row: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, map[string]string{"message": "Row inserted successfully"})
}

func (h *StudioHandler) UpdateRow(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	var req UpdateRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	tableName := chi.URLParam(r, "table_name")
	schema := r.URL.Query().Get("schema")

	if err := client.UpdateRow(r.Context(), tableName, schema, req.PrimaryKey, req.Data); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "update_failed", "Failed to update row: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Row updated successfully"})
}

func (h *StudioHandler) DeleteRow(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	var req DeleteRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	tableName := chi.URLParam(r, "table_name")
	schema := r.URL.Query().Get("schema")

	if err := client.DeleteRow(r.Context(), tableName, schema, req.PrimaryKey); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "delete_failed", "Failed to delete row: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Row deleted successfully"})
}

func (h *StudioHandler) GetDatabaseInfo(w http.ResponseWriter, r *http.Request) {
	database, err := h.validateDatabaseAccess(r)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_database", "Invalid database or access denied")
		return
	}

	client, err := h.clientFactory.CreateClient(r.Context(), database)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "connection_failed", "Failed to connect to database: "+err.Error())
		return
	}
	defer func() {
		_ = client.Close()
	}()

	version, _ := client.GetDatabaseVersion(r.Context())
	size, _ := client.GetDatabaseSize(r.Context())

	response := DatabaseInfoResponse{
		Version: version,
		Size:    size,
	}

	utils.SendJSON(w, http.StatusOK, response)
}
