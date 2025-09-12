package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Register a new user
// @Description Creates a new user with the provided credentials
// @Security ApiKeyAuth
// @Accept  json
// @Produce json,plain
// @Param   user  body  models.RegisterRequest  true  "Credentials for new user"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Invalid request payload"
// @Failure 403 {string} string "User exists"
// @Router /register [post]
func RegisterDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.RegisterRequest
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Email == "" || thisRequest.Name == "" || thisRequest.Role == "" {
		respondWithError(w, "Email, name or role can't be empty", http.StatusBadRequest)
		return
	}
	if len(thisRequest.Password) < 8 {
		respondWithError(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}
	if thisRequest.Role == "driver" && (thisRequest.LicenseID == nil || *thisRequest.LicenseID == "") {
		respondWithError(w, "License ID is required to register as a driver", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(thisRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO users (
			fullname,
			password_hash,
			email,
			role
		) VALUES ($1, $2, $3, $4
		) RETURNING id`
	var userID int
	err = db.DB.QueryRow(query, thisRequest.Name, string(hashedPassword), thisRequest.Email, "driver").Scan(&userID)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok && dbError.Code.Name() == "unique_violation" {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		respondWithError(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if thisRequest.Role == "driver" {
		query = `
			INSERT INTO drivers (
				user_id,
				license_id
			) VALUES ($1, $2);`
		_, err = db.DB.Exec(query, userID, thisRequest.LicenseID)
		if err != nil {
			respondWithError(w, "Failed to register driver: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "User registration successful. Login to get started!"})
}

// @Summary Log a user in
// @Description Authenticate a user with provided credentials
// @Security ApiKeyAuth
// @Accept  json
// @Produce json,plain
// @Param   user  body  models.LoginRequest  true  "User login credentials"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Router /login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.LoginRequest
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Email == "" || thisRequest.Password == "" {
		respondWithError(w, "Input all fields to login", http.StatusBadRequest)
		return
	}

	query := `
		SELECT
			u.id,
			u.password_hash,
			r.name,
			array_agg(p.name) AS permission_list
		FROM users AS u
		JOIN roles AS r
		ON u.role = r.name
		JOIN role_permissions AS rp
		ON r.id = rp.role_id
		JOIN permissions AS p
		ON rp.permission_id = p.id
		WHERE u.email = $1
		GROUP BY
			u.id,
			r.name;
		`
	var userID int
	var hashedPassword string
	var role string
	var permissions []string
	err = db.DB.QueryRow(query, thisRequest.Email).Scan(&userID, &hashedPassword, &role, (*pq.StringArray)(&permissions))
	if err == sql.ErrNoRows {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondWithError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(thisRequest.Password)); err != nil {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(userID, role, permissions)
	if err != nil {
		respondWithError(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Login successful!", "token": token})
}
