package clients

import (
	"context"

	proto "geico.visualstudio.com/Billing/plutus/proto/paymentmethodvalidation"
)

func ValidatePaymentMethod(
	userId string,
	amount float32,
	firstName string,
	lastName string,
	accountNumber string,
	rotingNumber string,
	accountType string) (*proto.PaymentMethodValidationResponse, error) {
	conn, err := getExecutorConnection()
	if err != nil {
		return nil, err
	}
	client := proto.NewPaymentMethodValidationServiceClient(conn)

	response, err := client.ValidatePaymentMethod(context.Background(),
		&proto.PaymentMethodValidationRequest{
			UserId:        userId,
			Amount:        amount,
			FirstName:     firstName,
			LastName:      lastName,
			AccountNumber: accountNumber,
			RoutingNumber: rotingNumber,
			AccountType:   accountType})
	if err != nil {
		log.Error(context.Background(), err, "business error recieved from grpc server")
		return nil, err
	}
	defer conn.Close()
	return response, nil
}
