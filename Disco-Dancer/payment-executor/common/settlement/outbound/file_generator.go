package settlement

import (
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	AchOnetime "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound/ach-onetime"
)

type FileGenerator struct {
}

func (fileGenerator FileGenerator) GenerateFile(settledAchPayments []app.SettlementPayment) (string, error) {

	fileLines, err := AchOnetime.GenerateAchOnetimeFile(settledAchPayments)
	if err != nil {
		return "", err
	}

	return fileLines, nil
}
