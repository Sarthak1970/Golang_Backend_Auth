package handler

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PublicProfileFields defines the fields to retrieve for PublicProfile
var PublicProfileFields = bson.M{
	"_id":            1,
	"email":          1,
	"name":           1,
	"bio":            1,
	"web_address":    1,
	"channel_name":   1,
	"area_of_expert": 1,
	"profile_picture": 1,
	"verified":       1,
	"location":       1,
	"provider":       1,
	"follower":       1,
	"following":      1,
}

// PublicProfileMedia2Fields defines the fields for PublicProfileMedia2
var PublicProfileMedia2Fields = bson.M{
	"_id":   1,
	"email": 1,
	"name":  1,
}

// Room represents a room document
type Room struct {
	ID        primitive.ObjectID `bson:"_id"`
	Creator   bson.M             `bson:"creator"`
	Live      bool               `bson:"live,omitempty"`
	Schedule  int64              `bson:"schedule,omitempty"`
	// add other fields according to documentation later
}

// Video represents a video document
type Video struct {
	ID      primitive.ObjectID `bson:"_id"`
	Profile bson.M             `bson:"profile"`
}

// ProfileHandler handles the /profile endpoint
func ProfileHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(string)
		if !ok {
			writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		collection := client.Database("auth").Collection("profile")
		var user User
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			writeJSONError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		err = collection.FindOne(context.Background(), bson.M{"_id": objID}, options.FindOne().SetProjection(PublicProfileFields)).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"_id":            user.ID.Hex(),
			"email":          user.Email,
			"name":           user.Name,
			"bio":            user.Bio,
			"web_address":    user.WebAddress,
			"channel_name":   user.ChannelName,
			"area_of_expert": user.AreaOfExpert,
			"profile_picture": user.ProfilePicture,
			"verified":       user.Verified,
			"location":       user.Location,
			"provider":       user.Provider,
			"follower_count": len(user.Followers),
			"following_count": len(user.Following),
		}

		response := Response{
			Data:    data,
			Message: "Profile Data Extracted",
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// PublicProfileHandler handles the /public endpoint
func PublicProfileHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSONError(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			writeJSONError(w, "Invalid id format", http.StatusBadRequest)
			return
		}

		collection := client.Database("auth").Collection("profile")
		var user User
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}, options.FindOne().SetProjection(PublicProfileFields)).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"_id":            user.ID.Hex(),
			"email":          user.Email,
			"name":           user.Name,
			"bio":            user.Bio,
			"web_address":    user.WebAddress,
			"channel_name":   user.ChannelName,
			"area_of_expert": user.AreaOfExpert,
			"profile_picture": user.ProfilePicture,
			"verified":       user.Verified,
			"location":       user.Location,
			"provider":       user.Provider,
			"follower_count": len(user.Followers),
			"following_count": len(user.Following),
		}
		if userID, ok := r.Context().Value("userID").(string); ok {
			isFollowing := false
			for _, follower := range user.Followers {
				if follower == userID {
					isFollowing = true
					break
				}
			}
			data["is_following"] = isFollowing
		}

		response := Response{
			Data:    data,
			Message: "Data Extracted",
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// PublicProfileMediaHandler handles the /profile/media endpoint
func PublicProfileMediaHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSONError(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			writeJSONError(w, "Invalid id format", http.StatusBadRequest)
			return
		}
		collection := client.Database("auth").Collection("profile")
		var user bson.M
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetching videos
		videoCollection := client.Database("videos").Collection("upload")
		videoCursor, err := videoCollection.Find(context.Background(), bson.M{"profile._id": objID})
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer videoCursor.Close(context.Background())

		var videos []bson.M
		if err := videoCursor.All(context.Background(), &videos); err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetch rooms
		roomCollection := client.Database("myspace").Collection("rooms")
		roomCursor, err := roomCollection.Find(context.Background(), bson.M{"creator._id": objID})
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer roomCursor.Close(context.Background())

		var rooms []bson.M
		if err := roomCursor.All(context.Background(), &rooms); err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"videos": videos,
			"rooms":  rooms,
		}

		response := Response{
			Data:    data,
			Message: "Media Extracted for " + objID.Hex(),
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// PublicProfileMediaLiveHandler handles the /profile/media/live endpoint
func PublicProfileMediaLiveHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSONError(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			writeJSONError(w, "Invalid id format", http.StatusBadRequest)
			return
		}

		// Check if user exists
		collection := client.Database("auth").Collection("profile")
		var user bson.M
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetch live rooms
		roomCollection := client.Database("myspace").Collection("rooms")
		cursor, err := roomCollection.Find(context.Background(), bson.M{"creator._id": objID, "live": true})
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var liveRooms []bson.M
		if err := cursor.All(context.Background(), &liveRooms); err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"rooms": liveRooms,
		}

		response := Response{
			Data:    data,
			Message: "Live Rooms of the User having ID " + objID.Hex(),
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// PublicProfileMediaUpcomingHandler handles the /profile/media/upcoming endpoint
func PublicProfileMediaUpcomingHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSONError(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			writeJSONError(w, "Invalid id format", http.StatusBadRequest)
			return
		}

		collection := client.Database("auth").Collection("profile")
		var user bson.M
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		roomCollection := client.Database("myspace").Collection("rooms")
		cursor, err := roomCollection.Find(context.Background(), bson.M{"creator._id": objID})
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var rooms []Room
		if err := cursor.All(context.Background(), &rooms); err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var upcomingRooms []Room
		now := time.Now()
		for _, room := range rooms {
			scheduleTime := time.Unix(room.Schedule/1000, 0)
			if scheduleTime.After(now) {
				upcomingRooms = append(upcomingRooms, room)
			}
		}

		data := map[string]interface{}{
			"rooms": upcomingRooms,
		}

		response := Response{
			Data:    data,
			Message: "Upcoming Rooms of the User having ID " + objID.Hex(),
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// PublicProfileMedia2Handler handles the /profile/media2 endpoint
func PublicProfileMedia2Handler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSONError(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			writeJSONError(w, "Invalid id format", http.StatusBadRequest)
			return
		}

		collection := client.Database("auth").Collection("profile")
		var user User
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}, options.FindOne().SetProjection(PublicProfileMedia2Fields)).Decode(&user)
		if err == mongo.ErrNoDocuments {
			writeJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if err != nil {
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"_id":   user.ID.Hex(),
			"email": user.Email,
			"name":  user.Name,
		}

		response := Response{
			Data:    data,
			Message: "Profile Data Extracted",
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}