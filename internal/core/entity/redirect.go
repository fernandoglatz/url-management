package entity

import (
	"time"
)

type Redirect struct {
	ID        string    `json:"id" bson:"id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`

	DNS         string `json:"dns,omitempty" bson:"dns,omitempty"`
	Destination string `json:"destination,omitempty" bson:"destination,omitempty"`
	Proxy       bool   `json:"proxy" bson:"proxy"`
}
