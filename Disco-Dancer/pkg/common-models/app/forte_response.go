package app

type ForteResponse struct {
	Transaction_Id       string   `json:"transaction_id"`
	Location_Id          string   `json:"location_id"`
	Customer_Id          string   `json:"customer_id"`
	Action               string   `json:"action"`
	Authorization_Amount float32  `json:"authorization_amount"`
	Authorization_Code   string   `json:"authorization_code"`
	Entered_By           string   `json:"entered_by"`
	ECheck               ECheck   `json:"echeck"`
	Response             Response `json:"response"`
}

type ECheck struct {
	Account_Holder        string `json:"account_holder"`
	Masked_Account_Number string `json:"masked_account_number"`
	Last_4_Account_Number string `json:"last_4_account_number"`
	Routing_Number        string `json:"routing_number"`
	Account_Type          string `json:"account_type"`
}

type Response struct {
	Environment        string `json:"environment"`
	Response_Type      string `json:"response_type"`
	Response_Code      string `json:"response_code"`
	Response_Desc      string `json:"response_desc"`
	Authorization_Code string `json:"authorization_code"`
	Preauth_Result     string `json:"preauth_result"`
	Preauth_desc       string `json:"preauth_desc"`
}
