package handlers

import (
	"net/http"
	"strconv"

	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/internal/pkg/models"
)

func GetAllOwners(w http.ResponseWriter, r *http.Request) {
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
			vo.fleet_size,
			COUNT(*) OVER () AS total_rows
		FROM users AS u
		JOIN
			vehicle_owners AS vo
			ON u.id = vo.user_id
		ORDER BY u.id ASC
		LIMIT $1 OFFSET $2`
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		respondWithError(w, "Failed to retrieve vehicle_owners: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var owners []models.VehicleOwner
	var no_of_owners int
	for rows.Next() {
		var thisOwner models.VehicleOwner
		if err := rows.Scan(
			&thisOwner.UserID,
			&thisOwner.Fullname,
			&thisOwner.Email,
			&thisOwner.FleetSize,
			&no_of_owners,
		); err != nil {
			respondWithError(w, "Error scanning owner row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		owners = append(owners, thisOwner)
	}

	if err := rows.Err(); err != nil {
		respondWithError(w, "Error iterating owner rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":                 owners,
		"page":                 page,
		"limit":                limit,
		"total":                len(owners),
		"total_vehicle_owners": no_of_owners,
	})
}
