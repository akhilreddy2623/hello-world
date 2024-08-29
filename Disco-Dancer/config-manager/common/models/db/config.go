package db

type ConfigResponse struct {
	ConfigId    int
	Description string
	Scope       string
	Environment string
	Key         string
	Value       string
	Application string
	Tenant      string
	Product     string
	Vendor      string
}