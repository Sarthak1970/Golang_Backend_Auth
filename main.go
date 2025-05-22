package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"Backend-Auth-Profiles/handler"
	"Backend-Auth-Profiles/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	router := mux.NewRouter()

	router.HandleFunc("/auth/google", handler.GoogleLoginHandler(client)).Methods("POST")
	router.HandleFunc("/auth/facebook", handler.FacebookLoginHandler(client)).Methods("POST")
	router.HandleFunc("/auth/refresh", utils.JWTMiddleware(handlers.RefreshTokenHandler(client))).Methods("POST")

	router.HandleFunc("/profile", utils.JWTMiddleware(handlers.ProfileHandler(client))).Methods("GET")
	router.HandleFunc("/profile/picture", utils.JWTMiddleware(handlers.ProfilePictureUploadHandler(client))).Methods("PUT")

	router.HandleFunc("/public", utils.LooseJWTMiddleware(handlers.PublicProfileHandler(client))).Methods("GET")
	router.HandleFunc("/profile/media", utils.LooseJWTMiddleware(handlers.PublicProfileMediaHandler(client))).Methods("GET")
	router.HandleFunc("/profile/media/live", utils.LooseJWTMiddleware(handlers.PublicProfileMediaLiveHandler(client))).Methods("GET")
	router.HandleFunc("/profile/media/upcoming", utils.LooseJWTMiddleware(handlers.PublicProfileMediaUpcomingHandler(client))).Methods("GET")
	router.HandleFunc("/profile/media2", utils.LooseJWTMiddleware(handlers.PublicProfileMedia2Handler(client))).Methods("GET")

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(router)))
}