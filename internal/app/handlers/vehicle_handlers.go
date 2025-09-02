package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"
)

func AddVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userDetails, ok := auth.GetUserDetailsFromContext(r.Context())
	if !ok {
		respondWithError(w, "User details not found in context. Authentication is required", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.AddVehicleRequest
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Make == "" || thisRequest.Model == "" ||
		thisRequest.Year == "" || thisRequest.LicensePlate == "" {
		respondWithError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	thisVehicle := Vehicle{
		Make:         thisRequest.Make,
		Model:        thisRequest.Model,
		Year:         thisRequest.Year,
		LicensePlate: thisRequest.LicensePlate,
		Status:       "available",
	}

	query := `
		INSERT INTO vehicles (
			make,
			model,
			year,
			license_plate,
			status
		) VALUES ($1, $2, $3, $4, $5
		) RETURNING id`
	err = db.DB.QueryRow(query, thisVehicle.Make, thisVehicle.Model,
		thisVehicle.Year, thisVehicle.LicensePlate, thisVehicle.Status).Scan(&thisVehicle.ID)
	if err != nil {
		respondWithError(w, "Failed to add vehicle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, thisVehicle)
}

func GetAllVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserDetailsFromContext(r.Context())
	if !ok {
		respondWithError(w, "User ID not found in context. Authentication is required", http.StatusUnauthorized)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, title, description
		FROM todos
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3`
	rows, err := db.DB.Query(query, userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve todos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.TodoItem
	for rows.Next() {
		var todo models.TodoItem
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Desc); err != nil {
			http.Error(w, "Error scanning todo row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating todo rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": todos, "page": page, "limit": limit, "total": len(todos)})
}

// @Summary Update a ToDo item
// @Description Edit an existing to-do item for the authenticated user
// @Tags todos
// @Security ApiKeyAuth
// @Accept  json
// @Produce json,plain
// @Param   todo  body  models.CreateRequest  true  "New details for the to-do item"
// @Success 200 {object} models.TodoItem
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Todo doesn't exist"
// @Router /todos/{id} [put]
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context. Authentication is required", http.StatusUnauthorized)
		return
	}

	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 3 || pathSegments[2] == "" {
		http.Error(w, "Todo ID missing in URL path", http.StatusBadRequest)
		return
	}
	todoID, err := strconv.Atoi(pathSegments[2])
	if err != nil {
		http.Error(w, "Invalid todo ID format. Must be an integer.", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.CreateRequest
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Title == "" || thisRequest.Desc == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	updatedTodo := models.TodoItem{
		Title: thisRequest.Title,
		Desc:  thisRequest.Desc,
	}

	query := `
		UPDATE todos
		SET title = $1, description = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id`
	err = db.DB.QueryRow(query, updatedTodo.Title, updatedTodo.Desc, todoID, userID).Scan(&updatedTodo.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "Todo not found", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedTodo)
}

// @Summary Delete a ToDo item
// @Description Delete an existing to-do item for the authenticated user
// @Tags todos
// @Security ApiKeyAuth
// @Accept  json
// @Produce json,plain
// @Success 204
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Todo doesn't exist"
// @Router /todos/{id} [delete]
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context. Authentication is required", http.StatusUnauthorized)
		return
	}

	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 3 || pathSegments[2] == "" {
		http.Error(w, "Todo ID missing in URL path", http.StatusBadRequest)
		return
	}
	todoID, err := strconv.Atoi(pathSegments[2])
	if err != nil {
		http.Error(w, "Invalid todo ID format. Must be an integer.", http.StatusBadRequest)
		return
	}

	query := `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
		RETURNING id`
	var deletedTodoID int
	err = db.DB.QueryRow(query, todoID, userID).Scan(&deletedTodoID)
	if err == sql.ErrNoRows {
		http.Error(w, "Todo not found", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get ToDo items
// @Description Retrieve to-do items for the authenticated user
// @Tags todos
// @Security ApiKeyAuth
// @Produce json,plain
// @Param   page  query integer true "The page to view"
// @Param   limit  query integer true "Number of items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string "Unauthorized"
// @Router /todos [get]
func GetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context. Authentication is required", http.StatusUnauthorized)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, title, description
		FROM todos
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3`
	rows, err := db.DB.Query(query, userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve todos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.TodoItem
	for rows.Next() {
		var todo models.TodoItem
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Desc); err != nil {
			http.Error(w, "Error scanning todo row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating todo rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": todos, "page": page, "limit": limit, "total": len(todos)})
}
