package request

type RedirectRequest struct {
	DNS         string `json:"dns,omitempty"`
	URI         string `json:"uri,omitempty"`
	Destination string `json:"destination,omitempty"`
	Proxy       bool   `json:"proxy"`
}
