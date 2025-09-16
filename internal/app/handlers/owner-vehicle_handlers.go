package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

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
		thisRequest.Year <= 0 || thisRequest.LicensePlate == "" {
		respondWithError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	thisVehicle := models.Vehicle{
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
			status,
			owner_id
		) VALUES ($1, $2, $3, $4, $5, $6
		) RETURNING id`
	err = db.DB.QueryRow(query, thisVehicle.Make, thisVehicle.Model,
		thisVehicle.Year, thisVehicle.LicensePlate, thisVehicle.Status, userDetails.UserID).Scan(&thisVehicle.ID)
	if err != nil {
		respondWithError(w, "Failed to add vehicle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query = `
		UPDATE vehicle_owners
		SET
			fleet_size = fleet_size + 1
		WHERE
			user_id = $1`
	_, err = db.DB.Exec(query, userDetails.UserID)
	if err != nil {
		respondWithError(w, "Failed to update user details"+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, thisVehicle)
}

func GetMyVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	userDetails, ok := auth.GetUserDetailsFromContext(r.Context())
	if !ok {
		respondWithError(w, "User details not found in context. Authentication is required", http.StatusUnauthorized)
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

	var fleet_size int
	query := `
		SELECT fleet_size FROM vehicle_owners
		WHERE
			user_id = $1`
	_, err = db.DB.QueryRow(query, userDetails.UserID).Scan(&fleet_size)
	if err != nil {
		respondWithError(w, "Failed to get fleet details"+err.Error(), http.StatusInternalServerError)
		return
	}

	query = `
		SELECT
			v.id,
			v.make,
			v.model,
			v.year,
			v.license_plate,
			v.status,
			u.fullname,
			u.email
		FROM vehicles AS v
		LEFT JOIN
			drivers AS d
			ON v.driver_id = d.user_id
		LEFT JOIN
			users AS u
			ON d.user_id = u.id
		WHERE v.owner_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3`
	rows, err := db.DB.Query(query, userDetails.UserID, limit, offset)
	if err != nil {
		respondWithError(w, "Failed to retrieve vehicles: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var thisVehicle models.Vehicle
		if err := rows.Scan(
			&thisVehicle.ID,
			&thisVehicle.Make,
			&thisVehicle.Model,
			&thisVehicle.Year,
			&thisVehicle.LicensePlate,
			&thisVehicle.Status,
			&thisVehicle.DriverName,
			&thisVehicle.DriverEmail,
		); err != nil {
			respondWithError(w, "Error scanning todo row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		vehicles = append(vehicles, thisVehicle)
	}
	if err := rows.Err(); err != nil {
		respondWithError(w, "Error iterating vehicle rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": vehicles, "page": page, "limit": limit, "total": len(vehicles), "fleet_size": fleet_size})
}
