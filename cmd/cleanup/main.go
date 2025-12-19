package main

import (
	"log"
	"pdf_generator/pkg/database"
)

func main() {
	if err := database.InitDB("app.db"); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	db := database.GetDB()
	if err := db.Exec("DELETE FROM sessions").Error; err != nil {
		log.Fatalf("Failed to clear sessions: %v", err)
	}

	log.Println("Successfully cleared all sessions.")
}
