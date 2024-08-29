package clients

import (
	"context"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	proto "geico.visualstudio.com/Billing/plutus/proto/paymentpreference"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetPaymentPreference(userId string, productIdentifier string, transactionType string, paymentRequestType string) ([]*proto.PaymentPreference, error) {

	isDummyResponse := commonFunctions.GetConfigHandler().GetBool("Dummy.PaymentPreference.Response", false)

	if isDummyResponse {
		dummyPaymentPreference := []*proto.PaymentPreference{
			{
				PaymentMethodType:      "ach",
				AccountIdentifier:      "Q\"j\"=F]K^",
				RoutingNumber:          "011100106",
				PaymentExtendedData:    "{\"FirstName\":\"Martin\",\"LastName\":\"Martin\",\"ACHAccountType\":\"checking\"}",
				WalletStatus:           true,
				PaymentMethodStatus:    "active",
				AccountValidationDate:  &timestamppb.Timestamp{},
				WalletAccess:           true,
				AutoPayPreference:      false,
				Split:                  100,
				Last4AccountIdentifier: "1234",
			},
		}
		return dummyPaymentPreference, nil
	} else {
		conn, err := getPaymentVaultConnection()
		if err != nil {
			return nil, err
		}
		client := proto.NewPaymentPreferenceServiceClient(conn)

		response, err := client.GetPaymentPreference(context.Background(),
			&proto.PaymentPreferenceRequest{
				UserId:             userId,
				ProductIdentifier:  productIdentifier,
				TransactionType:    transactionType,
				PaymentRequestType: paymentRequestType})
		if err != nil {
			log.Error(context.Background(), err, "business error recieved from grpc server")
			return nil, err
		}
		defer conn.Close()
		return response.PaymentPreference, nil
	}
}
