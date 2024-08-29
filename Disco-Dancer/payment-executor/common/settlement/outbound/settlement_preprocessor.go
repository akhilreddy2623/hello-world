package settlement

import (
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
	ConsolidatedRepository "geico.visualstudio.com/Billing/plutus/payment-executor-common/repository"
	AchOnetime "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound/ach-onetime"
)

type SettlementPreProcessor struct {
}

func (SettlementPreProcessor) FilePreProcessor(settlementData []db.SettlementPayment) ([]db.SettlementPayment, error) {

	for i, settlementrecord := range settlementData {

		settlementData[i].SettlementIdentifier = AchOnetime.GetIndividualId(settlementrecord.ConsolidatedId, 15, "C")
	}

	err := ConsolidatedRepository.UpdateSelltlementIdentifer(settlementData)

	return settlementData, err
}
