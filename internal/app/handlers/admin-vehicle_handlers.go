package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"
)

func GetAllVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	_, ok := auth.GetUserDetailsFromContext(r.Context())
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

	query := `
		SELECT
			v.id,
			v.make,
			v.model,
			v.year,
			v.license_plate,
			v.status,
			u.fullname AS driver_name,
			u.email AS driver_email,
			uo.fullname AS owner_name,
			uo.email AS owner_email,
			COUNT(*) OVER () AS total_rows
		FROM vehicles AS v
		LEFT JOIN
			drivers AS d
			ON v.driver_id = d.user_id
		LEFT JOIN
			users AS u
			ON d.user_id = u.id
		LEFT JOIN
			users AS uo
			ON v.owner_id = uo.id
		ORDER BY id ASC
		LIMIT $1 OFFSET $2`
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		respondWithError(w, "Failed to retrieve vehicles: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	var fleet_size int
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
			&thisVehicle.OwnerName,
			&thisVehicle.OwnerEmail,
			&fleet_size,
		); err != nil {
			respondWithError(w, "Error scanning vehicle row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		vehicles = append(vehicles, thisVehicle)
	}
	if err := rows.Err(); err != nil {
		respondWithError(w, "Error iterating vehicle rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       vehicles,
		"page":       page,
		"limit":      limit,
		"total":      len(vehicles),
		"fleet_size": fleet_size,
	})
}

func AssignDriver(w http.ResponseWriter, r *http.Request) {
	// 1. Get vehicle ID from URL
	vehicleIDStr := r.PathValue("id")
	if vehicleIDStr == "" {
		respondWithError(w, "Vehicle ID not found in URL", http.StatusBadRequest)
		return
	}
	vehicleID, err := strconv.Atoi(vehicleIDStr)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	// 2. Decode the request body to get the driver ID
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.AssignRequest
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.DriverID == "" {
		respondWithError(w, "All fields are required", http.StatusBadRequest)
		return

	}

	var currentVehicleStatus string
	var currentDriverAssigned bool

	// Check if vehicle is available
	err = db.DB.QueryRow("SELECT status FROM vehicles WHERE id = $1", vehicleID).Scan(&currentVehicleStatus)
	if err != nil {
		respondWithError(w, "Vehicle not found", http.StatusNotFound)
		return
	}
	if currentVehicleStatus != "available" {
		respondWithError(w, "Vehicle is not available for assignment", http.StatusConflict)
		return
	}

	// Check if driver is unassigned
	err = db.DB.QueryRow("SELECT assigned FROM drivers WHERE user_id = $1", thisRequest.DriverID).Scan(&currentDriverAssigned)
	if err != nil {
		respondWithError(w, "Driver not found", http.StatusNotFound)
		return
	}
	if currentDriverAssigned {
		respondWithError(w, "Driver is already assigned to a vehicle", http.StatusConflict)
		return
	}

	// 5. Update both the vehicle and driver records
	_, err = db.DB.Exec("UPDATE vehicles SET driver_id = $1, status = 'in_use' WHERE id = $2", thisRequest.DriverID, vehicleID)
	if err != nil {
		log.Printf("Failed to update vehicle: %v", err)
		respondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE drivers SET assigned = true WHERE user_id = $1", thisRequest.DriverID)
	if err != nil {
		log.Printf("Failed to update driver: %v", err)
		respondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"message": "Vehicle assigned successfully"})
}
