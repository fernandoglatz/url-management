package request

import (
	redirecttype "fernandoglatz/url-management/internal/core/entity/redirect"
)

type RedirectRequest struct {
	DNS         string            `json:"dns,omitempty"`
	Destination string            `json:"destination,omitempty"`
	Type        redirecttype.Type `json:"type" swaggertype:"string" enums:"PROXY,REDIRECT,IFRAME"`
}
