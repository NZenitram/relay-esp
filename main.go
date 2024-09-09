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

	// Initialize user controller
	userController := controllers.NewUserController(db)

	// Define routes
	r.HandleFunc("/users", userController.CreateUser).Methods("POST")
	r.HandleFunc("/users", userController.GetUsers).Methods("GET")
	r.HandleFunc("/users/{id}", userController.GetUser).Methods("GET")
	r.HandleFunc("/users/{id}", userController.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", userController.DeleteUser).Methods("DELETE")

	// Start server
	log.Println("Server is running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
