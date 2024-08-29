package handlers

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	handlerMocks "geico.visualstudio.com/Billing/plutus/payment-administrator-api/handlers/mocks"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/models"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository/mocks"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/stretchr/testify/assert"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	appmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/app"
	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
)

type TestCase struct {
	name  string
	input *appmodels.PaymentRequest
	want  *commonAppModels.ErrorResponse
}

var testConfig = config.NewConfigBuilder().
	AddJsonFile("../../api/config/appsettings.json").
	AddJsonFile("../../api/config/secrets.json").Build()

func Test_ValidateMakePaymentInputs(t *testing.T) {
	paymentRepositoryInterface := mocks.PaymentRepositoryInterface{}
	paymentHandlerInterface := handlerMocks.PaymentHandlerInterface{}

	ValidTenantIds = testConfig.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	MaxPaymentAmountIaa = testConfig.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)
	testCases := []TestCase{
		{
			name: "TenantIdShouldBeValidPositiveBigInt",
			input: &appmodels.PaymentRequest{
				TenantId: -23},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "tenantId should be a valid positive big int",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "TenantIdShouldBeValid",
			input: &appmodels.PaymentRequest{
				TenantId: 100},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "unknown tenantId, please provide a valid tenantId",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "TenantRequestIdShouldBeValidPositiveBigInt",
			input: &appmodels.PaymentRequest{
				TenantId:        101,
				TenantRequestId: 0},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "tenantRequestId should be a valid positive big int",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "UserIdShouldNotBeEmpty",
			input: &appmodels.PaymentRequest{
				TenantId:        101,
				TenantRequestId: 123,
				UserId:          ""},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "useId should be a non empty string",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "ProductIdentifierShouldNotBeEmpty",
			input: &appmodels.PaymentRequest{
				TenantId:          101,
				TenantRequestId:   123,
				UserId:            "ABC123",
				ProductIdentifier: ""},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "productIdentifier should be a non empty string",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentFrequencyShouldBeValidValue",
			input: &appmodels.PaymentRequest{
				TenantId:          101,
				TenantRequestId:   123,
				UserId:            "ABC123",
				ProductIdentifier: "XYZ123",
				PaymentFrequency:  "invalidValue"},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "paymentFrequency should be onetime or recurring",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "TransactionTypeShouldBeValidValue",
			input: &appmodels.PaymentRequest{
				TenantId:          101,
				TenantRequestId:   123,
				UserId:            "ABC123",
				ProductIdentifier: "XYZ123",
				PaymentFrequency:  "onetime",
				TransactionType:   "invalidValue"},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "transactionType should be payin or payout",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "CallerAppShouldBeValidValue",
			input: &appmodels.PaymentRequest{
				TenantId:          101,
				TenantRequestId:   123,
				UserId:            "ABC123",
				ProductIdentifier: "XYZ123",
				PaymentFrequency:  "onetime",
				TransactionType:   "payin",
				CallerApp:         "invalidValue"},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "caller app should be atlas or mcp",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentRequestTypeShouldBeValidValue",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "invalidValue"},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "unknown paymentRequestType. Please provide a valid value.",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleShouldNotBeEmpty",
			input: &appmodels.PaymentRequest{
				TenantId:                  101,
				TenantRequestId:           123,
				UserId:                    "ABC123",
				ProductIdentifier:         "XYZ123",
				PaymentFrequency:          "onetime",
				TransactionType:           "payin",
				CallerApp:                 "atlas",
				PaymentRequestType:        "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "empty payment extraction schedule",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAndPaymentFrequencyCombinationShouldBeValid",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 12.32},
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 32.12},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "invalid payment extraction schedule for onetime paymentFrequency",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAmountShouldBeValid",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: -1},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "provide a valid payment amount greater than 0",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleNoAmount",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date: models.JsonDate(time.Now()),
					},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "provide a valid payment amount greater than 0",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAmountShouldBeValidForRecurring",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "recurring",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 200},
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 0},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "provide a valid payment amount greater than 0",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAmountShouldBeLessThanMaxLimitForOnetime",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "mcp",
				PaymentRequestType: "insuranceautoauctions",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 2000.01},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "payment amount exceeds the maximum allowed limit of $2000 for InsuranceAutoAuctions",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAndPaymentFrequencyCombinationShouldBeValid1",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "mcp",
				PaymentRequestType: "customerchoice",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 12.32},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "for customerchoice payment request, atlas should be the caller app",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "PaymentExtractionScheduleAndPaymentFrequencyCombinationShouldBeValid2",
			input: &appmodels.PaymentRequest{
				TenantId:           101,
				TenantRequestId:    123,
				UserId:             "ABC123",
				ProductIdentifier:  "XYZ123",
				PaymentFrequency:   "onetime",
				TransactionType:    "payin",
				CallerApp:          "atlas",
				PaymentRequestType: "insuranceautoauctions",
				PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
					{
						Date:   models.JsonDate(time.Now()),
						Amount: 12.32},
				}},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "for insuranceautoauctions payment request, mcp should be the caller app",
				StatusCode: http.StatusBadRequest},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := makePayment(testCase.input, &paymentRepositoryInterface, &paymentHandlerInterface)

			assert.Equal(t, testCase.want.Message, got.Message)
			assert.Equal(t, testCase.want.Type, got.Type)
			assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
		})
	}

}

func Test_ValidateMakePaymentWithNoError(t *testing.T) {

	paymentRepository := mocks.PaymentRepositoryInterface{}
	paymentHandlerInterface := handlerMocks.PaymentHandlerInterface{}
	ValidTenantIds = testConfig.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	MaxPaymentAmountIaa = testConfig.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)

	paymentRequest := appmodels.PaymentRequest{
		TenantId:           101,
		TenantRequestId:    123,
		UserId:             "ABC123",
		ProductIdentifier:  "XYZ123",
		PaymentFrequency:   "onetime",
		TransactionType:    "payin",
		CallerApp:          "atlas",
		PaymentRequestType: "customerchoice",
		PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
			{
				Date:   models.JsonDate(time.Now()),
				Amount: 12.32},
		}}
	populateEnums(&paymentRequest)
	payment := getPaymentModel(&paymentRequest)

	var paymentList = []dbmodels.Payment{
		{
			PaymentId:         1,
			RequestId:         1,
			TenantRequestId:   1,
			ProductIdentifier: "XYZ123",
			AccountId:         "1",
			UserId:            "ABC123",
			Amount:            12.32,
			Status:            enums.Accepted.EnumIndex(),
		},
	}

	paymentRepository.On("MakePayment", &payment).Return(paymentList, nil)
	paymentHandlerInterface.On("PublishPaymentBalancingEventMessage", paymentList, payment).Return(nil)

	errorResponse := makePayment(&paymentRequest, &paymentRepository, &paymentHandlerInterface)
	assert.Nil(t, errorResponse)
}

func Test_ValidateMakePaymentWithError(t *testing.T) {

	paymentRepository := mocks.PaymentRepositoryInterface{}
	paymentHandlerInterface := handlerMocks.PaymentHandlerInterface{}
	ValidTenantIds = testConfig.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	MaxPaymentAmountIaa = testConfig.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)

	paymentRequest := appmodels.PaymentRequest{
		TenantId:           101,
		TenantRequestId:    123,
		UserId:             "ABC123",
		ProductIdentifier:  "XYZ123",
		PaymentFrequency:   "onetime",
		TransactionType:    "payin",
		CallerApp:          "atlas",
		PaymentRequestType: "customerchoice",
		PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
			{
				Date:   models.JsonDate(time.Now()),
				Amount: 12.32},
		}}

	populateEnums(&paymentRequest)
	payment := getPaymentModel(&paymentRequest)

	paymentRepository.On("MakePayment", &payment).Return(nil, errors.New(repository.ErrorUnableToGetPaymentPreferences))

	errorResponse := makePayment(&paymentRequest, &paymentRepository, &paymentHandlerInterface)
	assert.NotNil(t, errorResponse)
	assert.Equal(t, errorResponse.Message, repository.ErrorUnableToGetPaymentPreferences)
	assert.Equal(t, errorResponse.Type, invalidField)
	assert.Equal(t, errorResponse.StatusCode, 404)
}

func Test_ValidateMakePaymentWithOldDate(t *testing.T) {

	paymentRepository := mocks.PaymentRepositoryInterface{}
	paymentHandlerInterface := handlerMocks.PaymentHandlerInterface{}
	ValidTenantIds = testConfig.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	MaxPaymentAmountIaa = testConfig.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)

	paymentRequest := appmodels.PaymentRequest{
		TenantId:           101,
		TenantRequestId:    123,
		UserId:             "ABC123",
		ProductIdentifier:  "XYZ123",
		PaymentFrequency:   "onetime",
		TransactionType:    "payin",
		CallerApp:          "atlas",
		PaymentRequestType: "customerchoice",
		PaymentExtractionSchedule: []models.PaymentExtractionSchedule{
			{
				Date:   models.JsonDate(time.Now().AddDate(0, 0, -1)),
				Amount: 12.32},
		}}
	populateEnums(&paymentRequest)
	payment := getPaymentModel(&paymentRequest)

	var paymentList = []dbmodels.Payment{
		{
			PaymentId:         1,
			RequestId:         1,
			TenantRequestId:   1,
			ProductIdentifier: "XYZ123",
			AccountId:         "1",
			UserId:            "ABC123",
			Amount:            12.32,
			Status:            enums.Accepted.EnumIndex(),
		},
	}

	paymentRepository.On("MakePayment", &payment).Return(paymentList, nil)
	paymentHandlerInterface.On("PublishPaymentBalancingEventMessage", paymentList, payment).Return(nil)

	errorResponse := makePayment(&paymentRequest, &paymentRepository, &paymentHandlerInterface)
	assert.NotNil(t, errorResponse)
	assert.Equal(t, errorResponse.Type, invalidField)
	assert.Equal(t, errorResponse.StatusCode, 400)

}
