package dto

type RedirectRetrieveResponseDTO struct {
	Purposes []struct {
		Id          string `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
		Status      string `json:"status"`
		PurposeType string `json:"purposeType"`

		Agreements []struct {
			Id            string `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			DisplayAs     string `json:"displayAs"`
			Required      bool   `json:"required"`
			SelectionType string `json:"selectionType"`

			Options []struct {
				Id      string `json:"id"`
				Label   string `json:"label"`
				Checked bool   `json:"checked"`
			} `json:"options"`
		} `json:"agreements"`
	} `json:"purposes"`
}
