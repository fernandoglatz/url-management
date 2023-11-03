package dto

type RedirectSubmitRequestDTO struct {
	CollectionPointId string `json:"collectionPointId"`
	Channel           string `json:"channel"`

	Purposes []struct {
		Id string `json:"id"`

		Agreements []struct {
			Id string `json:"id"`

			Options []struct {
				Id string `json:"id"`
			} `json:"options"`
		} `json:"agreements"`
	} `json:"purposes"`
}
