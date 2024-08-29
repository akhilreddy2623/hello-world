package internal

import (
	"testing"

	"geico.visualstudio.com/Billing/plutus/config-manager-common/models/db"
	"geico.visualstudio.com/Billing/plutus/config-manager-common/repository/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var applicationName = "payment-vault"
var Environment = "DV1"

func Test_Configuration_ValidateRequest(t *testing.T) {
	mockRepo := new(mocks.ConfigRepositoryInterface)
	expectedConfigDetails := getTestConfigDetails()
	// Set up mock repository response

	applicationName := "payment-vault"
	Environment := "DV1"

	mockRepo.On("GetConfig", applicationName, Environment).Return(&expectedConfigDetails, nil)

	// Call the function under test
	response, err := getConfigDetails(applicationName, Environment, mockRepo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Add assertions to check the response
	assert.NotNil(t, response, "The response should not be nil.")
	assert.Equal(t, len(response.ConfigDetails), 6, "The response should contain 7 config details.")
	assert.Equal(t, response.ConfigDetails[0].Key, "WebServer.Port", "The first config detail should have the key 'WebServer.Port'.")
}

func Test_Configuration_InValidateRequest_HandleNoRecordsFound(t *testing.T) {
	mockRepo := new(mocks.ConfigRepositoryInterface)

	// Set up mock repository response

	applicationName := "payment-vault"
	Environment := "LD1"

	expectedConfigDetails := getTestConfigDetails(true)

	mockRepo.On("GetConfig", applicationName, Environment).Return(&expectedConfigDetails, nil)

	// Call the function under test
	response, err := getConfigDetails(applicationName, Environment, mockRepo)

	if response != nil && response.ConfigDetails != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			assert.Equal(t, e.Code(), codes.NotFound, "The error code should be NotFound.")
			assert.Equal(t, e.Message(), "no records found for payment-vault", "The error message should be 'No records found.'.")
		}
	}
}

func getTestConfigDetails(shouldSendEmptyResponse ...bool) []db.ConfigResponse {

	var expectedConfigDetails = []db.ConfigResponse{}
	if len(shouldSendEmptyResponse) > 0 && shouldSendEmptyResponse[0] {
		return expectedConfigDetails
	}

	expectedConfigDetails = []db.ConfigResponse{
		{
			Key:         "WebServer.Port",
			Value:       "30000",
			Environment: "DV1",
			Application: "payment-vault",
		},
		{
			Key:         "db.host",
			Value:       "localhost",
			Environment: "DV1",
			Application: "payment-vault",
		},
		{
			Key:         "ACHOriginID",
			Value:       "543258664",
			Environment: "DV1",
			Application: "payment-vault",
			Tenant:      "Claims",
		},
		{
			Key:         "ACHOriginID",
			Value:       "743258664",
			Environment: "DV1",
			Application: "payment-vault",
			Tenant:      "Auto",
		},

		{
			Key:         "db.dbname",
			Value:       "config-manager",
			Environment: "DV1",
			Application: "payment-vault",
		},
		{
			Key:   "PaymentPlatform.Grpc.Address",
			Value: "0.0.0.0:5051",
		},
	}

	return expectedConfigDetails

}
