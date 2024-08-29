//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/messaging"
	api "geico.visualstudio.com/Billing/plutus/task-manager-api/handlers"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
	"github.com/stretchr/testify/require"
)

var (
	testConfig config.AppConfiguration
)

func init() {

	testConfig = config.NewConfigBuilder().
		AddJsonFile("../api/config/appsettings.json").
		AddJsonFile("../worker/config/secrets.json").Build()

	fmt.Println("Initializing task manager integration test")

	database.Init(testConfig)
	err := messaging.InitKafka([]string{testConfig.GetString("PaymentPlatform.Kafka.Brokers", "")}, "dv.paymentplatform.groups.task-manager-api")
	if err != nil {
		panic(err)
	}

	commonFunctions.SetConfigHandler(testConfig)
}

func setupTestRequest(t *testing.T, reqBody []byte) *http.Response {
	err := SeedDataToPostgresTables(truncate_scheduled_task, insert_scheduled_task)
	require.NoError(t, err)
	ExecuteTask := api.Taskhandler{}.ExecuteTask
	req := httptest.NewRequest(http.MethodPost, "/task", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ExecuteTask(w, req)
	return w.Result()
}

func TestTaskIdNegativeValue(t *testing.T) {
	//Case Id : ITG020
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Task Id .
	// 3. Expected response from API should be - taskId should be a valid positive integer.

	expectedErrMsg := "taskId should be a valid positive integer"

	var reqBody = []byte(`{"TaskId": 0,"Year":   "2024","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "all","PaymentRequestType":   "all","PaymentFrequency":     "all"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusBadRequest {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedErrMsg) {
			require.Contains(t, string(body), expectedErrMsg)
		}
	}
	require.Contains(t, string(body), expectedErrMsg)

}

func TestExecuteTask_InvalidDate(t *testing.T) {
	// 	Case Id : ITG021
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Date parameter .
	// 3. Expected response from API should be - Year, Month and Date should be a valid date

	expectedErrMsg := "Year, Month and Date should be a valid date"

	var reqBody = []byte(`{"TaskId": 1,"Year":   "24","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "all","PaymentRequestType":   "all","PaymentFrequency":     "all"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusBadRequest {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedErrMsg) {
			require.Contains(t, string(body), expectedErrMsg)
		}
	}
	require.Contains(t, string(body), expectedErrMsg)
}

func TestInvalidPaymentMethodType(t *testing.T) {
	// 	Case Id : ITG022
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Payment Method Type parameter .
	// 3. Expected response from API should be - PaymentMethodType should be ach, card or all

	expectedErrMsg := "PaymentMethodType should be ach, card or all"

	var reqBody = []byte(`{"TaskId": 1,"Year":   "2024","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "visaCard","PaymentRequestType":   "all","PaymentFrequency":     "all"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusBadRequest {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedErrMsg) {
			require.Contains(t, string(body), expectedErrMsg)
		}
	}
	require.Contains(t, string(body), expectedErrMsg)
}

func TestInvalidPaymentRequestType(t *testing.T) {
	// Case Id : ITG023
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Payment Request Type parameter .
	// 3. Expected response from API should be - PaymentRequestType should be customerchoice, insuranceautoauctions, sweep, incentive, commission or all

	expectedErrMsg := "PaymentRequestType should be customerchoice, insuranceautoauctions, sweep, incentive, commission or all"

	var reqBody = []byte(`{"TaskId": 1,"Year":   "2024","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "all","PaymentRequestType":   "ABC","PaymentFrequency":     "all"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusBadRequest {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedErrMsg) {
			require.Contains(t, string(body), expectedErrMsg)
		}
	}
	require.Contains(t, string(body), expectedErrMsg)
}

func TestInvalidPaymentFrequency(t *testing.T) {
	// Case Id : ITG024
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Payment Request Type parameter .
	// 3. Expected response from API should be - PaymentRequestType should be customerchoice, insuranceautoauctions, sweep, incentive, commission or all

	expectedErrMsg := "PaymentFrequency should be onetime, recurring or all"

	var reqBody = []byte(`{"TaskId": 1,"Year":   "2024","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "all","PaymentRequestType":   "all","PaymentFrequency":     "ABC"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusBadRequest {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedErrMsg) {
			require.Contains(t, string(body), expectedErrMsg)
		}
	}
	require.Contains(t, string(body), expectedErrMsg)
}

func TestExecuteTaskAPI(t *testing.T) {
	// Case Id : ITG025
	// 1. Construct the payload as per the Test case, this is the prerequisite for the test case.
	// 2. Make a call to the /task endpoint using invalid Payment Frequency  parameter .
	// 3. Expected result make sure for valid call task is stored in task_execution table

	expectedMsg := "\"status\": \"accepted\""

	var reqBody = []byte(`{"TaskId": 1,"Year":   "2024","Month":  "04","Day":    "16","ExecutionParameters": {"PaymentMethodType":    "all","PaymentRequestType":   "all","PaymentFrequency":     "all"}}`)
	res := setupTestRequest(t, reqBody)
	defer res.Body.Close()
	body, resErr := io.ReadAll(res.Body)

	require.Nil(t, resErr)

	if res.StatusCode != http.StatusOK {
		require.EqualValues(t, res.StatusCode, http.StatusBadRequest)
	} else {
		if !strings.Contains(string(body), expectedMsg) {
			require.Contains(t, string(body), expectedMsg)
		}
		require.Contains(t, string(body), expectedMsg)
	}

	executeTaskRequestTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")

	requestTopic := make(chan *kafkamessaging.Message)
	messaging.KafkaSubscribe(executeTaskRequestTopic, func(ctx context.Context, message *kafkamessaging.Message) error {
		requestTopic <- message
		return nil
	})
	response := <-requestTopic

	// consume message so that it can be soft deleted
	var taskRequest commonMessagingModels.ExecuteTaskRequest
	err := json.Unmarshal([]byte(*response.Body), &taskRequest)
	require.NoError(t, err)
	require.EqualValues(t, taskRequest.TaskExecutionId, 1)
}
