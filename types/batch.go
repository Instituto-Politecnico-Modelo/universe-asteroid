package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Batch struct {
	// The ID of the batch
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}
