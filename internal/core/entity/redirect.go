package entity

import (
	"time"

	redirecttype "fernandoglatz/url-management/internal/core/entity/redirect"
)

type Redirect struct {
	ID        string    `json:"id" bson:"id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`

	DNS         string            `json:"dns,omitempty" bson:"dns,omitempty"`
	Destination string            `json:"destination,omitempty" bson:"destination,omitempty"`
	Type        redirecttype.Type `json:"type" bson:"type" swaggertype:"string" enums:"PROXY,REDIRECT,IFRAME"`
}
