package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Snapshot struct {
	// The ID of the snapshot
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// The filename of the snapshot
	Filename string `json:"filename" bson:"filename"`
	// The url of the snapshot
	URL string `json:"url" bson:"url"`
	// The timestamp of the snapshot
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	// The ID of the camera that took the snapshot
	CameraID primitive.ObjectID `json:"camera_id" bson:"camera_id,omitempty"`
	// The ID of the batch that the snapshot belongs to
	BatchID primitive.ObjectID `json:"batch_id" bson:"batch_id,omitempty"`
}
