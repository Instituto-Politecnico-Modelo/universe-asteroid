package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Batch struct {
	// The ID of the batch
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FetchCompleted bool               `json:"fetch_completed" bson:"fetch_completed"`
	Timestamp      time.Time          `json:"timestamp" bson:"timestamp"`
}
