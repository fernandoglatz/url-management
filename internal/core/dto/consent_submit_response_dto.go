package dto

type RedirectSubmitResponseDTO struct {
	TransactionId string `json:"transactionId"`

	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}
