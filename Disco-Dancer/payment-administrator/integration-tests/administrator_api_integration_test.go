//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"

	apiHandlers "geico.visualstudio.com/Billing/plutus/payment-administrator-api/handlers"

	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/stretchr/testify/require"
)

var (
	paymentEventsTopic string
	paymentEventChan          = make(chan *kafkamessaging.Message)
	newPaymentId       int64  = 2
	todayDate          string = time.Now().Format("2006-01-02")
)

type ExpectedPaymentRequestDbValues struct {
	RequestId          int64
	TenantId           int64
	TenantRequestId    int64
	UserId             string
	ProductIdentifier  string
	PaymentFrequency   enums.PaymentFrequency
	TransactionType    enums.TransactionType
	PaymentRequestType enums.PaymentRequestType
	CallerApp          enums.CallerApp
	Amount             float32
	Status             enums.PaymentStatus
}

type ExpectedPaymentDbValues struct {
	PaymentId         int64
	RequestId         int64
	TenantRequestId   int64
	UserId            string
	ProductIdentifier string
	PaymentFrequency  enums.PaymentFrequency
	Amount            float32
	Status            enums.PaymentStatus
}

func init() {
	var err error

	testConfig = config.NewConfigBuilder().
		AddJsonFile("../api/config/appsettings.json").
		AddJsonFile("../api/config/secrets.json").Build()

	fmt.Println("Initializing administrator api integration test")

	err = messaging.InitKafka([]string{testConfig.GetString("PaymentPlatform.Kafka.Brokers", "")}, "dv.paymentplatform.groups.administrator")
	if err != nil {
		panic(err)
	}

	apiHandlers.ValidTenantIds = testConfig.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	apiHandlers.MaxPaymentAmountIaa = testConfig.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)
	commonFunctions.SetConfigHandler(testConfig)
}

func cleanupTopicMessages(t *testing.T, callerApp enums.CallerApp) {
	err := commonFunctions.ProcessOneMesssageFromReader(paymentEventsReader, func(ctx context.Context, message *kafkamessaging.Message) error {
		var paymentEvent commonMessagingModels.PaymentEvent
		err := json.Unmarshal([]byte(*message.Body), &paymentEvent)
		require.NoError(t, err)
		require.EqualValues(t, paymentEvent.PaymentRequestType, callerApp)
		require.EqualValues(t, paymentEvent.EventType, enums.SavedInAdminstrator)
		return nil
	})

	require.NoError(t, err)
}

func setupPaymentRequestTests(t *testing.T, reqBody []byte) *http.Response {
	err := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, err)
	makePaymentHandler := apiHandlers.PaymentHandler{}.MakePaymentHandler
	req := httptest.NewRequest(http.MethodPost, "/payment", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	makePaymentHandler(w, req)
	return w.Result()
}

func checkPaymentResponseDb(t *testing.T, expectedValues ExpectedPaymentRequestDbValues) {
	// Check the values in the payment request table
	paymentRequest, amount, status, err := GetPaymentRequestFromDb(newPaymentId)
	require.NoError(t, err)
	require.EqualValues(t, paymentRequest.RequestId, expectedValues.RequestId)
	require.EqualValues(t, paymentRequest.TenantId, expectedValues.TenantId)
	require.EqualValues(t, paymentRequest.TenantRequestId, expectedValues.TenantRequestId)
	require.EqualValues(t, paymentRequest.UserId, expectedValues.UserId)
	require.EqualValues(t, paymentRequest.ProductIdentifier, expectedValues.ProductIdentifier)
	require.EqualValues(t, paymentRequest.PaymentFrequencyEnum, expectedValues.PaymentFrequency)
	require.EqualValues(t, paymentRequest.TransactionTypeEnum, expectedValues.TransactionType)
	require.EqualValues(t, paymentRequest.PaymentRequestTypeEnum, expectedValues.PaymentRequestType)
	require.EqualValues(t, paymentRequest.CallerAppEnum, expectedValues.CallerApp)
	require.EqualValues(t, amount, expectedValues.Amount)
	require.EqualValues(t, status, expectedValues.Status)
}

func checkPaymentDb(t *testing.T, expectedValues ExpectedPaymentDbValues) {
	// Check the values in the payment table
	payment, err := GetPaymentFromDb(newPaymentId)
	require.NoError(t, err)
	require.EqualValues(t, payment.PaymentId, expectedValues.PaymentId)
	require.EqualValues(t, payment.RequestId, expectedValues.RequestId)
	require.EqualValues(t, payment.TenantRequestId, expectedValues.TenantRequestId)
	require.EqualValues(t, payment.UserId, expectedValues.UserId)
	require.EqualValues(t, payment.ProductIdentifier, expectedValues.ProductIdentifier)
	require.EqualValues(t, payment.PaymentFrequencyEnum, expectedValues.PaymentFrequency)
	require.EqualValues(t, payment.Amount, expectedValues.Amount)
	require.EqualValues(t, payment.Status, expectedValues.Status)
}

func TestExecutePaymentRequest_SucceedsOnValidOneTimePayment(t *testing.T) {
	// ITG008
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a one time payment frequency.
	// 3. Expects no error responses and for the correct data to be inserted into the payment and incoming_payment_request tables
	expectedMsg := "\"status\": \"accepted\""
	expectedPaymentRequestDbValues := ExpectedPaymentRequestDbValues{
		RequestId:          newPaymentId,
		TenantId:           11,
		TenantRequestId:    11,
		UserId:             "IAA101",
		ProductIdentifier:  "ALL",
		PaymentFrequency:   enums.OneTime,
		TransactionType:    enums.PayIn,
		PaymentRequestType: enums.CustomerChoice,
		CallerApp:          enums.ATLAS,
		Amount:             float32(20.16),
		Status:             enums.Accepted,
	}
	expectedPaymentDbValues := ExpectedPaymentDbValues{
		PaymentId:         newPaymentId,
		RequestId:         newPaymentId,
		TenantRequestId:   11,
		UserId:            "IAA101",
		ProductIdentifier: "ALL",
		PaymentFrequency:  enums.OneTime,
		Amount:            float32(20.16),
		Status:            enums.Accepted,
	}
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusOK)
	require.Contains(t, string(body), expectedMsg)

	checkPaymentResponseDb(t, expectedPaymentRequestDbValues)
	checkPaymentDb(t, expectedPaymentDbValues)

	cleanupTopicMessages(t, enums.CallerApp(enums.CustomerChoice))
}

func TestExecutePaymentRequest_SucceedsOnValidRecurringPayment(t *testing.T) {
	// ITG009
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a recurring payment frequency.
	// 3. Expects no error responses and for the correct data to be inserted into the payment and incoming_payment_request tables
	expectedMsg := "\"status\": \"accepted\""
	expectedPaymentRequestDbValues := ExpectedPaymentRequestDbValues{
		RequestId:          newPaymentId,
		TenantId:           11,
		TenantRequestId:    11,
		UserId:             "IAA101",
		ProductIdentifier:  "ALL",
		PaymentFrequency:   enums.Recurrring,
		TransactionType:    enums.PayIn,
		PaymentRequestType: enums.CustomerChoice,
		CallerApp:          enums.ATLAS,
		Amount:             float32(20.16),
		Status:             enums.Accepted,
	}
	expectedPaymentDbValues := ExpectedPaymentDbValues{
		PaymentId:         newPaymentId,
		RequestId:         newPaymentId,
		TenantRequestId:   11,
		UserId:            "IAA101",
		ProductIdentifier: "ALL",
		PaymentFrequency:  enums.Recurrring,
		Amount:            float32(20.16),
		Status:            enums.Accepted,
	}
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"recurring","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusOK)
	require.Contains(t, string(body), expectedMsg)

	checkPaymentResponseDb(t, expectedPaymentRequestDbValues)
	checkPaymentDb(t, expectedPaymentDbValues)

	cleanupTopicMessages(t, enums.CallerApp(enums.CustomerChoice))
}

func TestExecutePaymentRequest_SucceedsOnValidMcpCallerAppAndInsuranceAutoAuctionsTypeRequest(t *testing.T) {
	// ITG0014, ITG0016
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a mcp caller app and Insurance Auto Auctions payment type.
	// 3. Expects no error responses and for the correct data to be inserted into the payment and incoming_payment_request tables

	expectedMsg := "\"status\": \"accepted\""
	expectedPaymentRequestDbValues := ExpectedPaymentRequestDbValues{
		RequestId:          newPaymentId,
		TenantId:           11,
		TenantRequestId:    11,
		UserId:             "IAA101",
		ProductIdentifier:  "ALL",
		PaymentFrequency:   enums.OneTime,
		TransactionType:    enums.PayIn,
		PaymentRequestType: enums.InsuranceAutoAuctions,
		CallerApp:          enums.MCP,
		Amount:             float32(88.88),
		Status:             enums.Accepted,
	}
	expectedPaymentDbValues := ExpectedPaymentDbValues{
		PaymentId:         newPaymentId,
		RequestId:         newPaymentId,
		TenantRequestId:   11,
		UserId:            "IAA101",
		ProductIdentifier: "ALL",
		PaymentFrequency:  enums.OneTime,
		Amount:            float32(88.88),
		Status:            enums.Accepted,
	}

	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":88.88}],"transactionType":"payin","callerApp":"mcp","paymentRequestType":"iaa"}`)
	pgerr := SeedDataToPostgresTables(truncate_Payment, truncate_PaymentRequest, insert_PaymentRequest, insert_Payment)
	require.NoError(t, pgerr)
	dbErr := SeedDataToVault(truncate_ProductDetails, truncate_PaymentMethod, truncate_PaymentPreference, insert_ProductDetails, insert_PaymentMethod, insert_PaymentPreference)
	require.NoError(t, dbErr)
	makePaymentHandler := apiHandlers.PaymentHandler{}.MakePaymentHandler
	req := httptest.NewRequest(http.MethodPost, "/payment", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	makePaymentHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusOK)
	require.Contains(t, string(body), expectedMsg)

	checkPaymentResponseDb(t, expectedPaymentRequestDbValues)
	checkPaymentDb(t, expectedPaymentDbValues)

	cleanupTopicMessages(t, enums.CallerApp(enums.MCP))
}
func TestExecutePaymentRequest_ErrorsOnDupTenantRequestId(t *testing.T) {
	// ITG001
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a tenantRequestId that already exists.
	// 3. Expects an error response for the duplicate tenantRequestId.
	expectedErrMsg := "unable to process payment, as previous payment request with same tenantid and tenantRequestId already exist"
	var reqBody = []byte(`{"tenantId":10,"tenantRequestId":10,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnNoVaultInfoForInsuranceAutoAuctions(t *testing.T) {
	// ITG002
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint for insuranceautoauctions without including the vault info.
	// 3. Expects an error response for the lack of vault info.
	expectedErrMsg := "unable to get payment preferences from payment vault"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"TestUser1","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"mcp","paymentRequestType":"insuranceautoauctions"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusNotFound)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnInvalidTenantId(t *testing.T) {
	// ITG003
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a tenantId that is invalid.
	// 3. Expects an error response for the invalid tenantId.
	expectedErrMsg := "tenantId should be a valid positive big int"
	var reqBody = []byte(`{"tenantId":-10,"tenantRequestId":10,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnInvalidTenantRequestId(t *testing.T) {
	// ITG004
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a tenantRequestId that is invalid.
	// 3. Expects an error response for the invalid tenantRequestId.
	expectedErrMsg := "tenantRequestId should be a valid positive big int"
	var reqBody = []byte(`{"tenantId":10,"tenantRequestId":-10,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnEmptyUserId(t *testing.T) {
	// ITG005
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using an empty string for the userId.
	// 3. Expects an error response for the empty userId.
	expectedErrMsg := "useId should be a non empty string"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnEmptyProductIdentifier(t *testing.T) {
	// ITG006
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using an empty string for the productIdentifier.
	// 3. Expects an error response for the empty productIdentifier.
	expectedErrMsg := "productIdentifier should be a non empty string"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnInvalidPaymentFrequency(t *testing.T) {
	// ITG007
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a paymentFrequency that is not onetime, recurring or subscription.
	// 3. Expects an error response for the invalid paymentFrequency.
	expectedErrMsg := "paymentFrequency should be onetime or recurring"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"never","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ErrorsOnInvalidTransactionType(t *testing.T) {
	// ITG010
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a transactionType that is not payin or payout.
	// 3. Expects an error response for the invalid transactionType.
	expectedErrMsg := "transactionType should be payin or payout"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"invalidetransactiontype","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := ioutil.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_FailsOnInvalidCallerApp(t *testing.T) {
	// ITG013
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a caller app that is not atlas or mcp.
	// 3. Expects an error response for the invalid caller app.
	expectedErrMsg := "caller app should be atlas or mcp"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"invalidCallerApp","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_FailsOnInvalidPaymentRequestType(t *testing.T) {
	// ITG015
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a payment request type that is not customerchoice or insurance auto auctions.
	// 3. Expects an error response for the invalid payment request type.
	expectedErrMsg := "unknown paymentRequestType. Please provide a valid value."
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"invalidPaymentRequestType"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_FailsOnEmptyExtractionSchedule(t *testing.T) {
	// ITG017
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using an empty extraction schedule.
	// 3. Expects an error response for the empty extraction schedule.
	expectedErrMsg := "empty payment extraction schedule"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_FailsOnOnetimePaymentFrequencyAndMultipleExtractionSchedule(t *testing.T) {
	// ITG018
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using a one time payment frequency and multiple extraction schedule.
	// 3. Expects an error response for invalid payment extraction schedule for one time payment frequency.
	expectedErrMsg := "invalid payment extraction schedule for onetime paymentFrequency"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16},{"date":"` + todayDate + `","amount":32.12}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_FailsOnInsuranceAutoAuctionsRequestTypeAndNonMcpCallerApp(t *testing.T) {
	// ITG019
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using insurance auto auctions payment request type and atlas caller app.
	// 3. Expects an error response for non mcp caller app for insurance auto auctions payment request.
	expectedErrMsg := "for insuranceautoauctions payment request, mcp should be the caller app"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"insuranceautoauctions"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecutePaymentRequest_ValidPayinTransactionType(t *testing.T) {
	// ITG011
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using valid PayIn transactionType
	// 3. Expects success response for the valid transactionType.

	expectedMsg := "\"status\": \"accepted\""
	//expectedEventType := "SAVEDINADMINSTRATOR"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payin","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusOK)
	require.Contains(t, string(body), expectedMsg)

	cleanupTopicMessages(t, enums.CallerApp(enums.CustomerChoice))
}

func TestExecutePaymentRequest_ValidPayOutTransactionType(t *testing.T) {
	// ITG012
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using valid PayOut transactionType
	// 3. Expects success response for the valid transactionType

	expectedMsg := "\"status\": \"accepted\""
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":20.16}],"transactionType":"payout","callerApp":"atlas","paymentRequestType":"customerchoice"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusOK)
	require.Contains(t, string(body), expectedMsg)
	cleanupTopicMessages(t, enums.CallerApp(enums.CustomerChoice))
}

func TestExecutePaymentRequest_AmountEqualToZero(t *testing.T) {
	// ITG046
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using Amount with 0$
	// 3. Expects Error response.

	expectedMsg := "provide a valid payment amount greater than 0"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":0.0}],"transactionType":"payin","callerApp":"mcp","paymentRequestType":"insuranceautoauctions"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedMsg)

}

func TestExecutePaymentRequest_AmountGreaterThanMaximumAmount(t *testing.T) {
	// ITG047
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using valid PayIn transactionType
	// 3. Expects Error response.

	expectedMsg := "payment amount exceeds the maximum allowed limit of $2000"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"` + todayDate + `","amount":5000.0}],"transactionType":"payin","callerApp":"mcp","paymentRequestType":"insuranceautoauctions"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedMsg)

}

func TestExecutePaymentRequest_ExtractionDateEarlierThanToday(t *testing.T) {
	// ITG048
	// 1. Setup the Data as per the prerequisite for the test case.
	// 2. Make a call to the /payment endpoint using valid PayIn transactionType
	// 3. Expects Error response.

	expectedMsg := "payment date cannot be earlier than today's date"
	var reqBody = []byte(`{"tenantId":11,"tenantRequestId":11,"userId":"IAA101","productIdentifier":"ALL","paymentFrequency":"onetime","paymentExtractionSchedule":[{"date":"2024-08-21","amount":100.0}],"transactionType":"payin","callerApp":"mcp","paymentRequestType":"insuranceautoauctions"}`)

	res := setupPaymentRequestTests(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)
	require.Nil(t, resErr)
	require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	require.Contains(t, string(body), expectedMsg)

}
