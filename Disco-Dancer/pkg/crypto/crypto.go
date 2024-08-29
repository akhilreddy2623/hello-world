package crypto

import (
	"context"

	cryptor_sidecar "artifactory-pd-infra.aks.aze1.cloud.geico.net/geico/IaaS-Cryptography-Services/_git/Cryptor-Sidecar.git"
	"geico.visualstudio.com/Billing/plutus/logging"
)

var cryptorClient *cryptor_sidecar.Client
var log = logging.GetLogger("crypto")
var version string = "6.0.0"

func Init(volatageSideCarAddress string) {
	cryptorClient = cryptor_sidecar.NewClient(volatageSideCarAddress)
}

func EncryptPaymentInfo(plainText string) (*string, error) {
	encryptPayload, err := cryptorClient.EncryptString(context.Background(), plainText, cryptor_sidecar.FormatPaymentInfo, version, "")
	if err != nil {
		log.Error(context.Background(), err, "error in encrypting payment info '%s'", plainText)
		return nil, err
	}

	return &encryptPayload.Ciphertext, nil
}

func DecryptPaymentInfo(cipherText string) (*string, error) {
	decryptPayload, err := cryptorClient.DecryptString(context.Background(), cipherText, cryptor_sidecar.FormatPaymentInfo, version)
	if err != nil {
		log.Error(context.Background(), err, "error in decrypting payment info '%s'", cipherText)
		return nil, err
	}

	return &decryptPayload.Plaintext, nil
}
