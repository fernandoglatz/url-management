package dto

type ApigeeTokenDTO struct {
	ExpiresIn   int    `json:"expires_in,string"`
	AccessToken string `json:"access_token"`
}
