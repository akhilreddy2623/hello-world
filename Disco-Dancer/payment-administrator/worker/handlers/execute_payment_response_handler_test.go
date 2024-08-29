package handlers

import (
	"context"
	"testing"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/enums"
	repositoryMock "geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository/mocks"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-worker/handlers/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_ExecutePaymentResponseHandlerSuccess(t *testing.T) {
	var tenantId, tenantRequestId int64 = 0001, 1234
	var paymentId int64 = 0001
	paymentResponse := commonMessagingModels.ExecutePaymentResponse{
		Version:         1,
		TenantId:        tenantId,
		TenantRequestId: tenantRequestId,
		Status:          enums.Completed,
	}
	executePaymentResponse := commonMessagingModels.ExecutePaymentResponse{
		Version:   1,
		PaymentId: paymentId,
		Status:    enums.Completed,
	}
	paymentRepositoryInterface := repositoryMock.PaymentRepositoryInterface{}
	executePaymentResponseHandlerInterface := mocks.ExecutePaymentResponseHandlerInterface{}
	paymentRepositoryInterface.On("UpdatePaymentStatus", executePaymentResponse).Return(nil)
	paymentRepositoryInterface.On("GetTenantInformationForPaymentId", executePaymentResponse.PaymentId).Return(tenantId, tenantRequestId, nil, nil)
	executePaymentResponseHandlerInterface.On("PublishPaymentResponse", paymentResponse).Return(nil)
	err := ProcessExecutePaymentResponse(context.Background(), executePaymentResponse, &paymentRepositoryInterface, &executePaymentResponseHandlerInterface)
	assert.Nil(t, err)
}

func Test_ExecutePaymentResponseHandlerWithErrorSuccess(t *testing.T) {
	var tenantId, tenantRequestId int64 = 0002, 1235
	var paymentId int64 = 0002
	paymentResponse := commonMessagingModels.ExecutePaymentResponse{
		Version:         1,
		TenantId:        tenantId,
		TenantRequestId: tenantRequestId,
		Status:          enums.Errored,
		ErrorDetails: &commonAppModels.ErrorResponse{
			Type:    "ForteValidationError",
			Message: "account number denied by forte",
		},
	}
	executePaymentResponse := commonMessagingModels.ExecutePaymentResponse{
		Version:   1,
		PaymentId: paymentId,
		Status:    enums.Errored,
		ErrorDetails: &commonAppModels.ErrorResponse{
			Type:    "ForteValidationError",
			Message: "account number denied by forte",
		},
	}
	paymentRepositoryInterface := repositoryMock.PaymentRepositoryInterface{}
	executePaymentResponseHandlerInterface := mocks.ExecutePaymentResponseHandlerInterface{}
	paymentRepositoryInterface.On("UpdatePaymentStatus", executePaymentResponse).Return(nil)
	paymentRepositoryInterface.On("GetTenantInformationForPaymentId", executePaymentResponse.PaymentId).Return(tenantId, tenantRequestId, nil, nil)
	executePaymentResponseHandlerInterface.On("PublishPaymentResponse", paymentResponse).Return(nil)
	err := ProcessExecutePaymentResponse(context.Background(), executePaymentResponse, &paymentRepositoryInterface, &executePaymentResponseHandlerInterface)
	assert.Nil(t, err)
}
