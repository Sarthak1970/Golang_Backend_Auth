package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProfilePictureUploadHandler handles the /profile/picture endpoint for uploading profile pictures
func ProfilePictureUploadHandler(client *mongo.Client) http.HandlerFunc {
	// Initialize AWS S3 uploader
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		fmt.Printf("Failed to initialize AWS session: %v\n", err)
	}

	uploader := s3manager.NewUploader(awsSession)

	return func(w http.ResponseWriter, r *http.Request) {
		// Get userID from context (set by JWTMiddleware in main.go)
		userID, ok := r.Context().Value("userID").(string)
		if !ok {
			writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeJSONError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get file from form
		file, handler, err := r.FormFile("display_picture")
		if err != nil {
			writeJSONError(w, "Failed to get display_picture: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate file (basic check for image)
		if !isImage(handler.Header.Get("Content-Type")) {
			writeJSONError(w, "Invalid file type; must be an image", http.StatusBadRequest)
			return
		}

		// Generate unique file name
		currentTime := time.Now().Format("20060102150405")
		filePath := fmt.Sprintf("display-pictures/dp_%s_%s.jpg", userID, currentTime)

		// Check if uploader is initialized
		if uploader == nil {
			writeJSONError(w, "AWS S3 uploader not initialized", http.StatusInternalServerError)
			return
		}

		// Upload to S3
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
			Key:         aws.String(filePath),
			Body:        file,
			ACL:         aws.String("public-read"),
			ContentType: aws.String("image/jpeg"),
		})
		if err != nil {
			writeJSONError(w, "Failed to upload to S3: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate image URL
		imageURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", os.Getenv("AWS_BUCKET"), filePath)

		// Update MongoDB profile
		collection := client.Database("auth").Collection("profile")
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			writeJSONError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{"profile_picture": imageURL}},
		)
		if err != nil {
			writeJSONError(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success response
		response := Response{
			Data:    map[string]string{"image_url": imageURL},
			Message: "Profile picture updated",
			Status:  true,
		}
		writeJSONResponse(w, response, http.StatusOK)
	}
}

// isImage checks if the content type is an image
func isImage(contentType string) bool {
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	for _, t := range allowedTypes {
		if t == contentType {
			return true
		}
	}
	return false
}