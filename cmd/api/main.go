// @title Fleet Management System API
// @version 1.0
// @description This is a sample fleet management backend service.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/kwagmire/fleet-management-api/internal/app/handlers"
	"github.com/kwagmire/fleet-management-api/internal/pkg/auth"
	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
	"github.com/kwagmire/fleet-management-api/migrations"
	//_ "github.com/kwagmire/fleet-management-api/docs"
)

/*func addSAdmin(fullname, email, password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Errorf("couldn't hash")
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
	err = db.DB.QueryRow(query, fullname, string(hashedPassword), email, "superadmin").Scan(&userID)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok && dbError.Code.Name() == "unique_violation" {
			fmt.Errorf("Email already exists")
			return
		}
		fmt.Errorf("Failed to register user: " + err.Error())
		return
	}
}*/

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Assuming environment variables are set in the environment.")
	}

	migrations.RunMigrations()

	db.ConnectDB()

	//addSAdmin("Boluwatiwi Oyebamiji", "boluwatiwioyebamiji@gmail.com", "datamayor")

	mux := http.NewServeMux()

	/*mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
	))*/

	mux.HandleFunc("POST /register", handlers.RegisterUser)
	mux.HandleFunc("POST /login", handlers.LoginUser)

	mux.HandleFunc(
		"POST /vehicles",
		auth.AuthMiddleware(auth.RequirePermission("owner:create.vehicle", handlers.AddVehicle)),
	)
	mux.HandleFunc(
		"GET /owned_vehicles",
		auth.AuthMiddleware(auth.RequirePermission("owner:read.vehicle", handlers.GetMyVehicles)),
	)

	mux.HandleFunc(
		"GET /vehicles",
		auth.AuthMiddleware(auth.RequirePermission("admin:read.driver", handlers.GetAllDrivers)),
	)
	mux.HandleFunc(
		"GET /drivers",
		auth.AuthMiddleware(auth.RequirePermission("admin:read.driver", handlers.GetAllDrivers)),
	)
	mux.HandleFunc(
		"GET /vehicle_owners",
		auth.AuthMiddleware(auth.RequirePermission("admin:read.owner", handlers.GetAllOwners)),
	)
	/*mux.HandleFunc(
		"PUT /vehicles/{id}/assign_driver",
		auth.AuthMiddleware(auth.RequirePermission("admin:assign.vehicle", handlers.AssignDriverToVehicle)),
	)*/

	//mux.HandleFunc("PUT /todos/", auth.AuthMiddleware(handlers.UpdateTodo))
	//mux.HandleFunc("DELETE /todos/", auth.AuthMiddleware(handlers.DeleteTodo))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)
	serverPort := ":8080"

	fmt.Printf("Fleet Management API server starting on http://localhost%s...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, handler))
}
