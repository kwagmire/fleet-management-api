package migrations

import (
	"log"

	"github.com/pressly/goose/v3"

	"github.com/kwagmire/fleet-management-api/internal/pkg/db"
)

func RunMigrations() {
	db.ConnectDB()
	// Specify the directory where your migration files are located
	//goose.SetDir("./migrations")

	// Run the migrations
	if err := goose.Up(db.DB, "./migrations"); err != nil {
		log.Fatalf("goose: failed to run migrations: %v\n", err)
	}

	log.Println("Database migrations applied successfully.")
}
