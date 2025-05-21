package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"Backend-Auth-Profiles/controller"
	"Backend-Auth-Profiles/handlers"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	} else {
		log.Println(".env file loaded successfully")
	}
	var client *mongo.Client = controller.GetMongoClient()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running!")
	})

	http.HandleFunc("/auth/google", handlers.GoogleLoginHandler(client))
	http.HandleFunc("/auth/facebook", handlers.FacebookLoginHandler(client))
	http.HandleFunc("/auth/refresh", handlers.RefreshTokenHandler)

	http.HandleFunc("/profile", handlers.ProfileHandler(client))
	http.HandleFunc("/public", handlers.PublicProfileHandler(client))
	http.HandleFunc("/profile/media", handlers.PublicProfileMediaHandler(client))
	http.HandleFunc("/profile/media/live", handlers.PublicProfileMediaLiveHandler(client))
	http.HandleFunc("/profile/media/upcoming", handlers.PublicProfileMediaUpcomingHandler(client))
	http.HandleFunc("/profile/media2", handlers.PublicProfileMedia2Handler(client))

	port := "8080"
	fmt.Println("Server started at http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}