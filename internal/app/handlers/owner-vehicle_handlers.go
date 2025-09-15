package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"
)

func AddVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	_, ok := auth.GetUserDetailsFromContext(r.Context())
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
