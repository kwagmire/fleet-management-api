package models

import "database/sql"

type Vehicle struct {
	ID           int            `json:"id"`
	Make         string         `json:"make"`
	Model        string         `json:"model"`
	Year         int            `json:"year"`
	LicensePlate string         `json:"license_plate"`
	Status       string         `json:"status"`
	DriverName   sql.NullString `json:"driver_name"` // sql.NullInt64 for nullable foreign key
	DriverEmail  sql.NullString `json:"driver_email"`
	OwnerName    sql.NullString `json:"owner_name"`
	OwnerEmail   sql.NullString `json:"owner_email"`
}

// Similar structs for User, Driver, etc.
type VehicleOwner struct {
	UserID    int    `json:"id"`
	Fullname  string `json:"username"`
	Email     string `json:"email"`
	FleetSize int    `json:"fleet_size"`
}

type Driver struct {
	UserID              int            `json:"user_id"`
	Fullname            string         `json:"fullname"`
	Email               string         `json:"email"`
	LicenseID           string         `json:"license_id"`
	Assigned            bool           `json:"assigned"`
	VehicleMake         sql.NullString `json:"vehicle_make"`
	VehicleModel        sql.NullString `json:"vehicle_model"`
	VehicleYear         sql.NullInt64  `json:"vehicle_year"`
	VehicleLicensePlate sql.NullString `json:"vehicle_license_plate"`
}
