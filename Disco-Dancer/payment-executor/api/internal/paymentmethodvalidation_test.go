package internal

import (
	"errors"
	"testing"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/external/mocks"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name         string
	input        app.PaymentMethodValidationRequest
	wantError    error
	wantResponse *app.PaymentMethodValidationResponse
}

func Test_ValidatePaymentMethodValidationInputs(t *testing.T) {
	validationApiInterface := mocks.ValidationApiInterface{}

	testCases := []TestCase{
		{
			name: "UseIdShouldBeValidAlphanumeric",
			input: app.PaymentMethodValidationRequest{
				UserId: "@XYZ123"},
			wantError:    errors.New("userId should be alphanumeric"),
			wantResponse: nil,
		},
		{
			name: "AmountIdShouldBeGreaterThanZero",
			input: app.PaymentMethodValidationRequest{
				UserId: "XYZ123",
				Amount: 0},
			wantError:    errors.New("amount should be greater than 0"),
			wantResponse: nil,
		},
		{
			name: "AccountNumberShouldBeNumeric",
			input: app.PaymentMethodValidationRequest{
				UserId:        "XYZ123",
				Amount:        10,
				AccountNumber: "12345#FG"},
			wantError:    errors.New("account number should be numeric"),
			wantResponse: nil,
		},
		{
			name: "RoutingNumberShouldBeNumeric",
			input: app.PaymentMethodValidationRequest{
				UserId:        "XYZ123",
				Amount:        10,
				AccountNumber: "12345987",
				RoutingNumber: "09876GF"},
			wantError:    errors.New("routing number should be numeric"),
			wantResponse: nil,
		},
		{
			name: "RoutingNumberShouldBeNumeric",
			input: app.PaymentMethodValidationRequest{
				UserId:        "XYZ123",
				Amount:        10,
				AccountNumber: "12345987",
				RoutingNumber: "09876",
				AccountType:   enums.NoneACHAccountType},
			wantError:    errors.New("account type should be checking or savings"),
			wantResponse: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, gotError := validatePaymentMethod(testCase.input, &validationApiInterface)

			assert.Nil(t, got)
			assert.Equal(t, testCase.wantError, gotError)
		})
	}
}

func Test_ValidatPaymentMethodValidationWithApproveddResponse(t *testing.T) {

	validationApiInterface := mocks.ValidationApiInterface{}

	forteRequest := app.ForteRequest{
		Action:               "verify",
		Customer_id:          "1884290256",
		Authorization_Amount: 12.32,
		Echeck: app.Echeck{
			Account_Holder: "A B",
			Account_Number: "123",
			Routing_Number: "123",
			Account_Type:   "checking"},
	}

	forteResponse := commonAppModels.ForteResponse{
		Transaction_Id:       "trn_720b5d9c-d443-4d94-b939-c53b2f48439f",
		Location_Id:          "loc_129706",
		Customer_Id:          "1884290256",
		Action:               "verify",
		Authorization_Amount: 12.32,
		Authorization_Code:   "60839541",
		Entered_By:           "fda2dbcdb9db9882b1ba3bbfc97ff5f3",
		ECheck: commonAppModels.ECheck{
			Account_Holder:        "ABC",
			Masked_Account_Number: "***123",
			Last_4_Account_Number: "123",
			Routing_Number:        "123",
			Account_Type:          "checking",
		},
		Response: commonAppModels.Response{
			Environment:        "sandbox",
			Response_Type:      "A",
			Response_Code:      "A01",
			Response_Desc:      "APPROVED",
			Authorization_Code: "60839541",
			Preauth_Result:     "POS",
			Preauth_desc:       "P70:VALIDATED",
		},
	}
	validationApiInterface.On("ForteValidation", forteRequest).Return(&forteResponse, nil)

	paymentMethodValidationRequest := app.PaymentMethodValidationRequest{
		UserId:        "XYZ123",
		Amount:        12.32,
		AccountNumber: "123",
		RoutingNumber: "123",
		FirstName:     "A",
		LastName:      "B",
		AccountType:   enums.Checking}

	paymentMethodValidationResponse, err := validatePaymentMethod(paymentMethodValidationRequest, &validationApiInterface)
	assert.Nil(t, err)
	assert.Equal(t, paymentMethodValidationResponse.Status, enums.ApprovedPaymentMethodValidationStatus)

}

func Test_ValidatPaymentMethodValidationWithDeclinedResponse(t *testing.T) {

	validationApiInterface := mocks.ValidationApiInterface{}

	forteRequest := app.ForteRequest{
		Action:               "verify",
		Customer_id:          "1884290256",
		Authorization_Amount: 12.32,
		Echeck: app.Echeck{
			Account_Holder: "A B",
			Account_Number: "123",
			Routing_Number: "123",
			Account_Type:   "checking"},
	}

	forteResponse := commonAppModels.ForteResponse{
		Transaction_Id:       "trn_720b5d9c-d443-4d94-b939-c53b2f48439f",
		Location_Id:          "loc_129706",
		Customer_Id:          "1884290256",
		Action:               "verify",
		Authorization_Amount: 12.32,
		Authorization_Code:   "60839541",
		Entered_By:           "fda2dbcdb9db9882b1ba3bbfc97ff5f3",
		ECheck: commonAppModels.ECheck{
			Account_Holder:        "ABC",
			Masked_Account_Number: "***123",
			Last_4_Account_Number: "123",
			Routing_Number:        "123",
			Account_Type:          "checking",
		},
		Response: commonAppModels.Response{
			Environment:        "sandbox",
			Response_Type:      "D",
			Response_Code:      "U25",
			Response_Desc:      "Declined",
			Authorization_Code: "60839541",
			Preauth_Result:     "POS",
			Preauth_desc:       "P70:VALIDATED",
		},
	}
	validationApiInterface.On("ForteValidation", forteRequest).Return(&forteResponse, nil)

	paymentMethodValidationRequest := app.PaymentMethodValidationRequest{
		UserId:        "XYZ123",
		Amount:        12.32,
		AccountNumber: "123",
		RoutingNumber: "123",
		FirstName:     "A",
		LastName:      "B",
		AccountType:   enums.Checking}

	paymentMethodValidationResponse, err := validatePaymentMethod(paymentMethodValidationRequest, &validationApiInterface)
	assert.Nil(t, err)
	assert.Equal(t, paymentMethodValidationResponse.Status, enums.DeclinedPaymentMethodValidationStatus)

}
