package internal

import (
	"fmt"
	"testing"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_GetValidPaymentPreferenceForPayIn(t *testing.T) {
	transactionTypeEnum := enums.PayIn
	paymentRequestTyoeEnum := enums.InsuranceAutoAuctions
	productIdentifier := "ABC123"
	validatePaymentPreference(t, productIdentifier, transactionTypeEnum, paymentRequestTyoeEnum)
}

func Test_GetValidPaymentPreferenceForPayOut(t *testing.T) {
	transactionTypeEnum := enums.PayOut
	paymentRequestTyoeEnum := enums.InsuranceAutoAuctions
	productIdentifier := "ABC123"
	validatePaymentPreference(t, productIdentifier, transactionTypeEnum, paymentRequestTyoeEnum)
}

func Test_GetValidPaymentPreferenceForAllProductIdentifier(t *testing.T) {
	transactionTypeEnum := enums.PayOut
	paymentRequestTyoeEnum := enums.InsuranceAutoAuctions
	productIdentifier := ""
	validatePaymentPreference(t, productIdentifier, transactionTypeEnum, paymentRequestTyoeEnum)
}

func Test_GetErrorForInvalidTransactionType(t *testing.T) {
	transactionTypeEnum := enums.NoneTransactionType //invalid
	paymentRequestTyoeEnum := enums.InsuranceAutoAuctions
	paymentPreferenceRepository := mocks.PaymentPreferenceRepositoryInterface{}
	paymentPreferenceRepository.On("GetPaymentPreference", "XYZ123", transactionTypeEnum, paymentRequestTyoeEnum).Return(nil, nil)

	response, err := getPaymentPreference("XYZ123", "ABC123", transactionTypeEnum, paymentRequestTyoeEnum, &paymentPreferenceRepository)
	assert.NotNil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprint(err), "invalid transaction type passed by client ''")
}

func validatePaymentPreference(
	t *testing.T,
	productIdentifier string,
	transactionTypeEnum enums.TransactionType,
	paymentRequestTypeEnum enums.PaymentRequestType) {
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
	paymentPreferenceRepository.On("GetPaymentPreference", "XYZ123", productIdentifier, transactionTypeEnum, paymentRequestTypeEnum).Return(paymentPreferenceList, nil)

	response, err := getPaymentPreference("XYZ123", productIdentifier, transactionTypeEnum, paymentRequestTypeEnum, &paymentPreferenceRepository)
	assert.Nil(t, err)
	assert.Equal(t, len(response.PaymentPreference), 2)
	assert.Equal(t, response.PaymentPreference[0].AccountIdentifier, "9234567")
	assert.Equal(t, response.PaymentPreference[0].Split, int32(50))
	assert.Equal(t, response.PaymentPreference[1].AccountIdentifier, "8234567")
	assert.Equal(t, response.PaymentPreference[1].Split, int32(50))
}
