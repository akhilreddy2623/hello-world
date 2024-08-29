package handlers

import (
	"errors"
	"net/http"
	"testing"
	"time"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

type TestCaseGetPaymentPreference struct {
	name                    string
	inputUserId             string
	inputTransactionType    string
	inputPaymentRequestType string
	inputProductIdentifier  string
	want                    *commonAppModels.ErrorResponse
}

func Test_ValidateGetExternalPaymentPreferenceInputs(t *testing.T) {

	paymentPreferenceRepositoryInterface := mocks.PaymentPreferenceRepositoryInterface{}
	testCases := []TestCaseGetPaymentPreference{
		{
			name:                    "UserIdShouldBeAlphanumeric",
			inputUserId:             "UserId@123",
			inputTransactionType:    "payin",
			inputPaymentRequestType: "insuranceautoauctions",
			inputProductIdentifier:  "PQR123",
			want: &commonAppModels.ErrorResponse{
				Type:       invalidUserId,
				Message:    "UserId can only contain alphabets and numbers.",
				StatusCode: http.StatusBadRequest},
		},
		{
			name:                    "UserIdShouldBeAlphanumeric",
			inputUserId:             " ",
			inputTransactionType:    "payin",
			inputPaymentRequestType: "insuranceautoauctions",
			inputProductIdentifier:  "PQR123",
			want: &commonAppModels.ErrorResponse{
				Type:       invalidUserId,
				Message:    "UserId can only contain alphabets and numbers.",
				StatusCode: http.StatusBadRequest},
		},
		{
			name:                    "TransactionTypeShouldBeValid",
			inputUserId:             "UserId123",
			inputTransactionType:    "payinout",
			inputPaymentRequestType: "insuranceautoauctions",
			inputProductIdentifier:  "PQR123",
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "transactionType should be a valid enum value",
				StatusCode: http.StatusBadRequest},
		},
		{
			name:                    "PaymentRequestTypeShouldBeValid",
			inputUserId:             "UserId123",
			inputTransactionType:    "payin",
			inputPaymentRequestType: "insuranceautoauctionstype",
			inputProductIdentifier:  "PQR123",
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "paymentRequestType should be a valid enum value",
				StatusCode: http.StatusBadRequest},
		},
		{
			name:                    "ProductIdentifierShouldBeValid",
			inputUserId:             "UserId123",
			inputTransactionType:    "payin",
			inputPaymentRequestType: "insuranceautoauctions",
			inputProductIdentifier:  "**--",
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "ProductIdentifier can only contain alphabets and numbers.",
				StatusCode: http.StatusBadRequest},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := getPaymentPreference(testCase.inputUserId, testCase.inputTransactionType, testCase.inputPaymentRequestType, testCase.inputProductIdentifier, &paymentPreferenceRepositoryInterface)

			assert.Equal(t, testCase.want.Message, err.Message)
			assert.Equal(t, testCase.want.Type, err.Type)
			assert.Equal(t, testCase.want.StatusCode, err.StatusCode)
		})
	}
}

func Test_GetExternalPaymentPreference_WithNoError(t *testing.T) {

	paymentPreferenceRepository := mocks.PaymentPreferenceRepositoryInterface{}
	var paymentPreferenceList []repository.PaymentPreference

	paymentPreferenceList = append(paymentPreferenceList, repository.PaymentPreference{
		PaymentMethodType:     1,
		AccountIdentifier:     "9234567",
		RoutingNumber:         "123456",
		PaymentExtendedData:   "{\"accountname\": \"temp 2\",\"accounttype\": \"checking\",\"bankname\": \"chase\"}",
		WalletStatus:          true,
		PaymentMethodStatus:   0,
		AccountValidationDate: time.Now(),
		WalletAccess:          true,
		AutoPayPreference:     false,
		Split:                 50,
	})

	paymentPreferenceList = append(paymentPreferenceList, repository.PaymentPreference{
		PaymentMethodType:     1,
		AccountIdentifier:     "8234567",
		RoutingNumber:         "223456",
		PaymentExtendedData:   "{\"accountname\": \"temp 3\",\"accounttype\": \"checking\",\"bankname\": \"bofa\"}",
		WalletStatus:          true,
		PaymentMethodStatus:   0,
		AccountValidationDate: time.Now(),
		WalletAccess:          true,
		AutoPayPreference:     false,
		Split:                 50,
	})

	paymentPreferenceRepository.On("GetPaymentPreference", "XYZ123", "ABC123", enums.PayIn, enums.InsuranceAutoAuctions).Return(paymentPreferenceList, nil)

	response, err := getPaymentPreference("XYZ123", "payin", "insuranceautoauctions", "ABC123", &paymentPreferenceRepository)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, len(response), 2)
	assert.Equal(t, response[0].AccountIdentifier, "9234567")
	assert.Equal(t, response[0].Split, int32(50))
	assert.Equal(t, response[1].AccountIdentifier, "8234567")
	assert.Equal(t, response[1].Split, int32(50))
}

func Test_StoreAddPaymentPreference_ValidateRequest(t *testing.T) {

	type TestCaseStorePaymentPreference struct {
		name  string
		input *app.PaymentPreferenceRequest
		want  *commonAppModels.ErrorResponse
	}

	handler := new(PaymentPreferenceHandler)
	testCases := []TestCaseStorePaymentPreference{
		{
			name: "UserIDShouldNotBeEmpty",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.UserId = ""
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "userid can only contain alphabets and numbers",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "ProducttypeShouldNotBeEmpty",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.ProductType = enums.NoneProductType
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "product type should be commercial, and ppa",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "ProductSubtypeShouldNotBeEmpty",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.ProductSubType = enums.NoneProductSubType
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "product sub type should be auto, cycle, and umbrella",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "ProductRequesttypeShouldNotBeEmpty",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.PaymentRequestType = enums.NonePaymentRequestType
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "payment request type should be all, commission, customerchoice, iaa, incentive, and sweep",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "PaymentMethodidShouldBeValid",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.PayIn = []*app.PaymentPreferenceDetail{
					{PaymentMethodId: 0, Split: 100, AutoPayPreference: true},
				}
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "paymentmethodid is not valid",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "SplitPercentageShouldBe100",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.PayIn = []*app.PaymentPreferenceDetail{
					{PaymentMethodId: 123, Split: 10, AutoPayPreference: true},
					{PaymentMethodId: 124, Split: 10, AutoPayPreference: true},
				}
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "split total percentage is not 100 in PayIn, got 20",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "PayInORPayOutShouldHaveValue",
			input: func() *app.PaymentPreferenceRequest {
				request := createPaymentPreferenceRequest()
				request.PayIn = nil
				request.PayOut = nil
				return request
			}(),
			want: &commonAppModels.ErrorResponse{
				Type:       paymentPreferenceValidationError,
				Message:    "either PayIn or PayOut should have a value",
				StatusCode: http.StatusBadRequest,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepository := new(mocks.PaymentPreferenceRepositoryInterface)

			_, err := handler.storePaymentPreferenceDetails(testCase.input, mockRepository)

			assert.Equal(t, testCase.want.Message, err.Message)
			assert.Equal(t, testCase.want.Type, err.Type)
			assert.Equal(t, testCase.want.StatusCode, err.StatusCode)
		})
	}

}

func Test_StoreAddPaymentPreference_ValidRequest_Success(t *testing.T) {

	handler := new(PaymentPreferenceHandler)
	mockRepository := new(mocks.PaymentPreferenceRepositoryInterface)
	request := createPaymentPreferenceRequest()

	err := handler.validatePaymentPreferenceRequest(request)
	var model *db.StorePaymentPreference
	if err == nil {
		model = handler.convertToPaymentPreferenceDbModel(request)
	}

	mockRepository.On("StorePaymentPreference", model).Return(nil)

	_, errorResponse := handler.storePaymentPreferenceDetails(request, mockRepository)
	assert.Nil(t, errorResponse)

}

func Test_StoreAddPaymentPreference_InValidRequest_Failure(t *testing.T) {

	handler := new(PaymentPreferenceHandler)
	mockRepository := new(mocks.PaymentPreferenceRepositoryInterface)
	request := createPaymentPreferenceRequest()

	err := handler.validatePaymentPreferenceRequest(request)
	var model *db.StorePaymentPreference
	if err == nil {
		model = handler.convertToPaymentPreferenceDbModel(request)
	}

	mockRepository.On("StorePaymentPreference", model).Return(errors.New("PaymentMethodId 4 is not active"))
	_, errorResponse := handler.storePaymentPreferenceDetails(request, mockRepository)
	assert.Equal(t, "PaymentMethodId 4 is not active", errorResponse.Message)

}

func createPaymentPreferenceRequest() *app.PaymentPreferenceRequest {

	request := app.PaymentPreferenceRequest{
		UserId:             "XYZ123",
		ProductType:        enums.Commercial,
		ProductSubType:     enums.Auto,
		PaymentRequestType: enums.InsuranceAutoAuctions,
		ProductIdentifier:  "ALL",
		PayIn: []*app.PaymentPreferenceDetail{
			{
				PaymentMethodId:   4,
				Split:             40,
				AutoPayPreference: true,
			},
			{
				PaymentMethodId:   5,
				Split:             60,
				AutoPayPreference: false,
			},
		},
		PayOut: []*app.PaymentPreferenceDetail{
			{
				PaymentMethodId:   4,
				Split:             40,
				AutoPayPreference: true,
			},
			{
				PaymentMethodId:   7,
				Split:             60,
				AutoPayPreference: false,
			},
		},
	}
	return &request
}
