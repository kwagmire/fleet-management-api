package models

type RegisterRequest struct {
	Name      string  `json:"fullname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
	LicenseID *string `json:"license_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AddVehicleRequest struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
}

type AssignRequest struct {
	DriverID string `json:"driver_id"`
}
