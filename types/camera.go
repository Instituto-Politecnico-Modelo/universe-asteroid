package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type Camera struct {
	// The ID of the camera
	ID primitive.ObjectID `json:"id" bson:"_id"`

	// The name of the camera
	Name string `json:"name" bson:"name"`

	// The RTSP URL of the camera
	URL string `json:"url" bson:"url"`

	// The location of the camera
	Location string `json:"location" bson:"location"`
}
