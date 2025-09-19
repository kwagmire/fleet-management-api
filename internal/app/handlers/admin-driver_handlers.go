package handlers

import (
	"net/http"
	"strconv"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"
)

func GetAllDrivers(w http.ResponseWriter, r *http.Request) {
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
			u.id,
			u.fullname,
			u.email,
			d.license_id,
			d.assigned,
			v.make,
			v.model,
			v.year,
			v.license_plate,
			COUNT(*) OVER () AS total_rows
		FROM users AS u
		JOIN
			drivers AS d
			ON u.id = d.user_id
		LEFT JOIN
			vehicles AS v
			ON d.user_id = v.driver_id
		ORDER BY u.id ASC
		LIMIT $1 OFFSET $2`
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		respondWithError(w, "Failed to retrieve drivers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var drivers []models.Driver
	var no_of_drivers int
	for rows.Next() {
		var thisDriver models.Driver
		if err := rows.Scan(
			&thisDriver.UserID,
			&thisDriver.Fullname,
			&thisDriver.Email,
			&thisDriver.LicenseID,
			&thisDriver.Assigned,
			&thisDriver.VehicleMake,
			&thisDriver.VehicleModel,
			&thisDriver.VehicleYear,
			&thisDriver.VehicleLicensePlate,
			&no_of_drivers,
		); err != nil {
			respondWithError(w, "Error scanning driver row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		drivers = append(drivers, thisDriver)
	}

	if err := rows.Err(); err != nil {
		respondWithError(w, "Error iterating driver rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":          drivers,
		"page":          page,
		"limit":         limit,
		"total":         len(drivers),
		"total_drivers": no_of_drivers,
	})
}
