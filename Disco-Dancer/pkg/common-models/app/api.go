package app

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type APIRequest struct {
	Type          enums.APIRequestType
	Url           string
	Request       []byte
	Authorization string
	Header        string
	Timeout       time.Duration
	FilePath      string
}

type APIResponse struct {
	Status   string
	Response []byte
}

type BearerTokenRequest struct {
	Url          string
	ClientId     string
	ClientSecret string
	Scope        string
	GrantType    string
}

type BearerTokenResponse struct {
	Token_Type     string
	Expires_In     int
	Ext_Expires_In int
	Access_Token   string
}
