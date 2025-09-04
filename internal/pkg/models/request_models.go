package models

type RegisterRequest struct {
	Name      string `json:"fullname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	LicenseID string `json:"license_id"`
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
