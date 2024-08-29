package handlers

// TODO : Fix the below test case

// func Test_executePaymentRequestProcess_WithNoError(t *testing.T) {

// 	request := commonMessagingModels.ProcessPaymentRequest{
// 		TenantId:            1234,
// 		PaymentId:           5678,
// 		PaymentFrequency:    "OneTime",
// 		TransactionType:     "payout",
// 		PaymentRequestType:  "insuranceautoauctions",
// 		PaymentMethodType:   "ach",
// 		PaymentExtendedData: "{\"accountname\": \"temp 2\",\"accounttype\": \"checking\",\"bankname\": \"chase\"}",
// 		AccountIdentifier:   "1234567890",
// 		RoutingNumber:       "021000021",
// 		Amount:              25.63,
// 		PaymentDate:         commonMessagingModels.JsonDate(time.Now()),
// 	}

// 	executePaymentRequestRepository := mocks.ExecutionRequestRepositoryInterface{}
// 	executePaymentRequestRepository.On("ExecutePayments", request).Return(nil)

// 	errorResponse := executePaymentRequestProcess(context.Background(), request, &executePaymentRequestRepository)

// 	assert.Nil(t, errorResponse)
// }
