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

	// Public routes
	r.HandleFunc("/login", userController.Login).Methods("POST")
	r.HandleFunc("/request-password-reset", userController.RequestPasswordReset).Methods("POST")
	r.HandleFunc("/reset-password", userController.ResetPassword).Methods("POST")

	// Protected routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.JWTAuth(db))
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
	api.HandleFunc("/esps", espController.CreateESP).Methods("POST")
	api.HandleFunc("/esps/{id}", espController.UpdateESP).Methods("PUT")
	api.HandleFunc("/esps/{id}", espController.DeleteESP).Methods("DELETE")

	// User event routes
	api.HandleFunc("/api/v1/event-stats", espController.GetUserEventStats).Methods("GET")
	// Start server
	log.Println("Server is running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
