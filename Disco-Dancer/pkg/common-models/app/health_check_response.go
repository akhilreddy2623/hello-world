package app

type HealthCheckResponse struct {
	Error             bool          `json:"error" example:"false"`
	Message           string        `json:"message" example:"I am alive !!"`
	Data              any           `json:"data,omitempty"`
	DatabaseStatus    string        `json:"databasestatus,omitempty" example:"Up"`
	ConfigFileName    string        `json:"configfilename,omitempty" example:"appsettings.dv.json"`
	SeceretsAvailable string        `json:"seceretsavailable,omitempty" example:"yes"`
	ErrorResponse     ErrorResponse `json:"errorresponse,omitempty"`
}
