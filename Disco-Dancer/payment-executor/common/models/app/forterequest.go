package app

type ForteRequest struct {
	Action               string  `json:"action"`
	Customer_id          string  `json:"customer_id"`
	Authorization_Amount float32 `json:"authorization_amount"`
	Echeck               Echeck  `json:"echeck"`
}

type Echeck struct {
	Account_Holder string `json:"account_holder"`
	Account_Number string `json:"account_number"`
	Routing_Number string `json:"routing_number"`
	Account_Type   string `json:"account_type"`
}
