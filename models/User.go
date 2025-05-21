package model

import (
    "time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID                 primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
    UserID             string                   `bson:"user_id" json:"user_id"`
    DeviceIDList       []string                 `bson:"device_id_list" json:"device_id_list"`
    Email              string                   `bson:"email" json:"email"`
    ChannelName        string                   `bson:"channel_name" json:"channel_name"`
    AreaOfExpert       []string                 `bson:"area_of_expert" json:"area_of_expert"`
    AreaOfInterest     map[string]map[string][]string `bson:"area_of_interest" json:"area_of_interest"` // branch -> category -> subcategories
    Name               string                   `bson:"name" json:"name"`
    Provider           string                   `bson:"provider" json:"provider"`
    Bio                string                   `bson:"bio" json:"bio"`
    Language           string                   `bson:"language" json:"language"`
    WebAddress         string                   `bson:"web_address" json:"web_address"`
    Location           string                   `bson:"location" json:"location"`
    Follower           []string                 `bson:"follower" json:"follower"`
    Following          []string                 `bson:"following" json:"following"`
    Verified           bool                     `bson:"verified" json:"verified"`
    ProfilePicture     string                   `bson:"profile_picture" json:"profile_picture"`
    ProfileOfInterest  []string                 `bson:"profile_of_interest" json:"profile_of_interest"`
    FCMToken           string                   `bson:"fcm_token,omitempty" json:"fcm_token,omitempty"`
    CreatedAt          time.Time                `bson:"created_at" json:"created_at"`
    UpdatedAt          time.Time                `bson:"updated_at" json:"updated_at"`
    RoomsCreated       int                      `bson:"rooms_created" json:"rooms_created"`
    Live               bool                     `bson:"live" json:"live"`
}

