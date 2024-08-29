package app

type WorkdayFeed struct {
	VendorType       string  `json:"vendorType"`
	Amount           float32 `json:"amount"`
	PaymentDate      string  `json:"paymentDate"`
	AtlasCheckNumber string  `json:"atlasCheckNumber"`
	ClaimNumber      string  `json:"claimNumber"`
	PublicId         string  `json:"publicId"`
}

type WorkdayFeedMetaData struct {
	VendorType  string
	CheckNumber string
	ClaimNumber string
	PublicId    string
}
