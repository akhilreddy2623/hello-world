package mocks

import (
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"github.com/stretchr/testify/mock"
)

type MockPaymentMethodRepository struct {
	mock.Mock
}

func (m *MockPaymentMethodRepository) StorePaymentMethod(paymentMethod *db.PaymentMethod) error {
	args := m.Called(paymentMethod)
	if paymentMethod != nil {
		paymentMethod.PaymentMethodId = 758265852
	}
	return args.Error(0)
}
