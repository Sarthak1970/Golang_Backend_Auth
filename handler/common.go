package handler

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  bool        `json:"status"`
}

type User struct {
	ID             primitive.ObjectID `bson:"_id"`
	Email          string             `bson:"email"`
	Name           string             `bson:"name"`
	Bio            string             `bson:"bio"`
	WebAddress     string             `bson:"web_address"`
	ChannelName    string             `bson:"channel_name"`
	AreaOfExpert   []string           `bson:"area_of_expert"`
	ProfilePicture string             `bson:"profile_picture"`
	Verified       bool               `bson:"verified"`
	Location       string             `bson:"location"`
	Provider       string             `bson:"provider"`
	Followers      []string           `bson:"follower"`
	Following      []string           `bson:"following"`
}

func writeJSONResponse(w http.ResponseWriter, response Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"message": "Internal server error", "status": false}`, http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := Response{
		Message: message,
		Status:  false,
	}
	writeJSONResponse(w, response, statusCode)
}