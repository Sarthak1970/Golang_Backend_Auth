package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"os"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"

	"Backend-Auth-Profiles/utils"
)

import model "Backend-Auth-Profiles/models"

// SocialAuthRequest defines the request structure for social login
type SocialAuthRequest struct {
	AuthToken string `json:"auth_token"` // Required
	DeviceID  string `json:"device_id,omitempty"` // Optional
	FCMToken  string `json:"fcm_token,omitempty"` // Optional
}

// RefreshTokenRequest defines the request structure for refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func GoogleLoginHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleSocialLogin(w, r, client, "google")
	}
}

func FacebookLoginHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleSocialLogin(w, r, client, "facebook")
	}
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	accessToken, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": accessToken,
	})
}

func handleSocialLogin(w http.ResponseWriter, r *http.Request, client *mongo.Client, provider string) {
	ctx := context.Background()
	var req SocialAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate auth_token is provided
	if req.AuthToken == "" {
		http.Error(w, "auth_token is required", http.StatusBadRequest)
		return
	}

	var userData map[string]interface{}
	var err error

	if provider == "google" {
		userData, err = validateGoogleToken(req.AuthToken)
	} else if provider == "facebook" {
		userData, err = validateFacebookToken(req.AuthToken)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID := userData["id"].(string)
	email := userData["email"].(string)
	name := userData["name"].(string)
	picture := userData["picture"].(string)

	collection := client.Database("authdb").Collection("profile")

	var user model.User
	err = collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		language := "en"
		areaOfInterest := map[string]map[string][]string{}
		profileOfInterest := []string{}
		username := generateUsername(name)

		user = model.User{
			UserID:            userID,
			Email:             email,
			Name:              name,
			ChannelName:       username,
			DeviceIDList:      []string{},
			AreaOfExpert:      []string{},
			AreaOfInterest:    areaOfInterest,
			Bio:               "",
			Language:          language,
			WebAddress:        "",
			Location:          "",
			Follower:          []string{},
			Following:         []string{},
			Verified:          false,
			ProfilePicture:    picture,
			ProfileOfInterest: profileOfInterest,
			FCMToken:          req.FCMToken,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			RoomsCreated:      0,
			Live:              false,
			Provider:          provider,
		}

		// Only add device_id if provided
		if req.DeviceID != "" {
			user.DeviceIDList = []string{req.DeviceID}
		}

		res, err := collection.InsertOne(ctx, user)
		if err != nil {
			http.Error(w, "User creation failed", http.StatusInternalServerError)
			return
		}
		user.ID = res.InsertedID.(primitive.ObjectID)
	} else if err == nil {
		update := bson.M{
			"$set": bson.M{
				"fcm_token":   req.FCMToken,
				"updated_at": time.Now(),
			},
		}
		// Only update device_id_list if device_id is provided
		if req.DeviceID != "" {
			update["$addToSet"] = bson.M{"device_id_list": req.DeviceID}
		}
		_, err := collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
		if err != nil {
			http.Error(w, "User update failed", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := utils.GenerateTokens(user.UserID)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func validateGoogleToken(token string) (map[string]interface{}, error) {
	ctx := context.Background()
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		log.Println("Error: GOOGLE_CLIENT_ID not set in .env")
		return nil, fmt.Errorf("server configuration error: GOOGLE_CLIENT_ID not set")
	}

	payload, err := idtoken.Validate(ctx, token, clientID)
	if err != nil {
		log.Printf("Error validating Google id_token: %v", err)
		return nil, fmt.Errorf("invalid Google token: %v", err)
	}

	sub, ok := payload.Claims["sub"].(string)
	if !ok || sub == "" {
		log.Println("Error: Missing or invalid 'sub' claim in id_token")
		return nil, fmt.Errorf("invalid Google token: missing sub claim")
	}
	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		log.Println("Error: Missing or invalid 'email' claim in id_token")
		return nil, fmt.Errorf("invalid Google token: missing email claim")
	}
	name, ok := payload.Claims["name"].(string)
	if !ok {
		name = "" 
	}
	picture, ok := payload.Claims["picture"].(string)
	if !ok {
		picture = "" 
	}
	return map[string]interface{}{
		"id":      sub,
		"email":   email,
		"name":    name,
		"picture": picture,
	}, nil
}

func validateFacebookToken(token string) (map[string]interface{}, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)&access_token=" + token)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid Facebook token")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Facebook response")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("Failed to parse Facebook response")
	}

	return map[string]interface{}{
		"id":      result["id"].(string),
		"email":   result["email"].(string),
		"name":    result["name"].(string),
		"picture": result["picture"].(map[string]interface{})["data"].(map[string]interface{})["url"].(string),
	}, nil
}

func generateUsername(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "")) + fmt.Sprintf("%d", time.Now().Unix()%1000)
}