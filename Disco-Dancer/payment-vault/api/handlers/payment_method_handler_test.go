package handlers

import (
	"errors"
	"net/http"
	"testing"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name  string
	input *app.PaymentMethodRequest
	want  *commonAppModels.ErrorResponse
}

func Test_PaymentMethod_ValidateRequest(t *testing.T) {

	testCases := []TestCase{

		{
			name: "UserIDShouldNotBeEmpty",
			input: &app.PaymentMethodRequest{
				UserID:            "",
				CallerApp:         enums.ATLAS,
				PaymentMethodType: enums.ACH,
			},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "userid cannot be empty",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name:  "PaymentMethodShouldBeACHORCARD",
			input: &app.PaymentMethodRequest{UserID: "324324", CallerApp: enums.ATLAS, PaymentMethodType: enums.NonePaymentMethodType},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "paymentmethodtype should be either ach or card",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name:  "CallerAppShouldNotBeEmpty",
			input: &app.PaymentMethodRequest{UserID: "324324", CallerApp: enums.NoneCallerApp, PaymentMethodType: enums.NonePaymentMethodType},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "callerapp should be valid",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "PaymentMethodDetailsShouldBePresent",
			input: &app.PaymentMethodRequest{
				UserID:            "testUser",
				PaymentMethodType: enums.ACH,
				FirstName:         "John",
				LastName:          "Doe",
				CallerApp:         enums.ATLAS,
				PaymentMethodDetails: app.Paymentmethoddetails{
					AccountNumber:  "",
					RoutingNumber:  "987654321",
					ACHAccountType: enums.Checking,
					NickName:       "MyAccount",
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "accountnumber must be between 1 to 17 digits",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "PaymentMethodDetailsShouldBePresent",
			input: &app.PaymentMethodRequest{
				UserID:            "testUser",
				PaymentMethodType: enums.ACH,
				FirstName:         "John",
				LastName:          "Doe",
				CallerApp:         enums.ATLAS,
				PaymentMethodDetails: app.Paymentmethoddetails{
					AccountNumber:  "54543",
					RoutingNumber:  "",
					ACHAccountType: enums.Checking,
					NickName:       "MyAccount",
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "routingnumber must be 9 digits",
				StatusCode: http.StatusBadRequest,
			},
		},

		{
			name: "PaymentMethodDetailsShouldBePresent",
			input: &app.PaymentMethodRequest{
				UserID:            "testUser",
				PaymentMethodType: enums.Card,
				FirstName:         "John",
				LastName:          "Doe",
				CallerApp:         enums.ATLAS,
				PaymentMethodDetails: app.Paymentmethoddetails{
					CardNumber: "",
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "cardnumber must be between 13 to 19 digits",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "PaymentMethodDetailsShouldBePresent",
			input: &app.PaymentMethodRequest{
				UserID:            "testUser",
				PaymentMethodType: enums.Card,
				FirstName:         "John",
				LastName:          "Doe",
				CallerApp:         enums.ATLAS,
				PaymentMethodDetails: app.Paymentmethoddetails{
					CardNumber:          "7894561258522535",
					ZipCode:             "",
					CVV:                 "026",
					NickName:            "MyAccount",
					CardExpirationMonth: "05",
					CardExpirationYear:  "2035",
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       "PaymentMethodValidationError",
				Message:    "zipcode must be 5 or 10 characters long",
				StatusCode: http.StatusBadRequest,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			mockRepository := new(mocks.MockPaymentMethodRepository)

			_, err := storePaymentMethodDetails(testCase.input, mockRepository)

			assert.Equal(t, testCase.want.Message, err.Message)
			assert.Equal(t, testCase.want.Type, err.Type)
			assert.Equal(t, testCase.want.StatusCode, err.StatusCode)
		})
	}

}

func Test_StoreAddPaymentMethod_ValidRequest_Success(t *testing.T) {

	mockRepository := new(mocks.MockPaymentMethodRepository)
	request := app.PaymentMethodRequest{
		UserID:            "23342",
		PaymentMethodType: enums.ACH,
		FirstName:         "Martin",
		LastName:          "Martin",
		CallerApp:         enums.ATLAS,
		PaymentMethodDetails: app.Paymentmethoddetails{
			AccountNumber:  "323243432",
			RoutingNumber:  "324324233",
			ACHAccountType: enums.Checking,
			NickName:       "Martin",
		},
	}

	err := validatePaymentMethodRequest(&request)
	var model db.PaymentMethod
	if err == nil {
		model = convertToPaymentMethodDBModel(&request)
	}

	mockRepository.On("StorePaymentMethod", &model).Return(nil)

	id, errorResponse := storePaymentMethodDetails(&request, mockRepository)
	assert.Nil(t, errorResponse)
	assert.Greater(t, *id, int64(0))

}

func Test_StoreAddPaymentMethod_ValidRequest_Failure(t *testing.T) {

	mockRepository := new(mocks.MockPaymentMethodRepository)
	request := app.PaymentMethodRequest{
		UserID:            "23342",
		PaymentMethodType: enums.ACH,
		FirstName:         "Martin",
		LastName:          "Martin",
		CallerApp:         enums.ATLAS,
		PaymentMethodDetails: app.Paymentmethoddetails{
			AccountNumber:  "323243432",
			RoutingNumber:  "324324233",
			ACHAccountType: enums.Checking,
			NickName:       "Martin",
		},
	}

	err := validatePaymentMethodRequest(&request)
	var model db.PaymentMethod
	if err == nil {
		model = convertToPaymentMethodDBModel(&request)
	}

	mockRepository.On("StorePaymentMethod", &model).Return(errors.New("Payment Method Already Exists"))
	_, errorResponse := storePaymentMethodDetails(&request, mockRepository)
	assert.NotNil(t, errorResponse)
	assert.Equal(t, "Payment Method Already Exists", errorResponse.Message)

}
