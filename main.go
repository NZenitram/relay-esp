// main.go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nzenitram/relay-esp/controllers"
	"github.com/nzenitram/relay-esp/database"
	"github.com/nzenitram/relay-esp/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	// Connect to the database
	database.InitDB()
	db := database.GetDB()

	// Initialize router
	r := mux.NewRouter()

	// Initialize controllers
	userController := controllers.NewUserController(db)
	eventController := controllers.NewEventController(db)
	espController := controllers.NewESPController(db)

	// Protected routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.APIKeyAuth(db))
	api.HandleFunc("/users", userController.GetUsers).Methods("GET")
	api.HandleFunc("/users/{id}", userController.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userController.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userController.DeleteUser).Methods("DELETE")

	// Event routes
	api.HandleFunc("/events", eventController.GetEvents).Methods("GET")
	api.HandleFunc("/events/types", eventController.GetAvailableEventTypes).Methods("GET")
	api.HandleFunc("/events/{type}", eventController.GetEventsByType).Methods("GET")

	// ESP routes
	api.HandleFunc("/esps", espController.GetESPs).Methods("GET")

	// Start server
	log.Println("Server is running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
