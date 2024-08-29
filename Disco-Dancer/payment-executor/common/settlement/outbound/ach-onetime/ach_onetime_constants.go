package AchOnetime

import (
	"geico.visualstudio.com/Billing/plutus/enums"
)

const (
	//File
	PriorityCode             = "01"
	FileIdModifiler          = "1"
	RecordSize               = "094"
	BlockingFactor           = "10"
	FormatCode               = "1"
	ImmediateDestinationName = "JP MORGAN CHASE        "
	ImmediateOriginName      = "GEICO INSURANCE COMPANY"

	//Batch
	ServiceClass            = "200"
	CompanyName             = "GEICO INSURANCE "
	CompanyEntryDescription = "EFT       "
	StatusCode              = "1"
	RoutingNumber           = "02100002"
)

type AchOnetimeFileData struct {
	ImmediateDestination     string
	ImmediateOrigin          string
	ReferenceCode            string
	CompanyDiscretionaryData string
	StandardEntryClassCode   string
	CompanyId                string
}

func GetFileInfo(paymentRequestType enums.PaymentRequestType) AchOnetimeFileData {
	switch paymentRequestType {
	case enums.InsuranceAutoAuctions:
		return getInsuranceAutoAuctionsFileInfo()
	default:
		return AchOnetimeFileData{}
	}
}

// TODO - get the correct values for IAA
func getInsuranceAutoAuctionsFileInfo() AchOnetimeFileData {
	return AchOnetimeFileData{
		ImmediateDestination:     " 021000021", //RoutingNumber
		ImmediateOrigin:          "9893476200", //AccountNumber
		ReferenceCode:            "00009042",   //TODO: check with workday team -
		CompanyDiscretionaryData: "                9042",
		StandardEntryClassCode:   "CCD",
		CompanyId:                "9893476201",
	}
}

// TODO - Commenting out the below code. Will be enabled when enterpriserental will be up and running
// func getEnterpriseRentalFileInfo() AchOnetimeFileData {
// 	return AchOnetimeFileData{
// 		ImmediateDestination:     " 021000021", //RoutingNumber
// 		ImmediateOrigin:          "9893476200", //AccountNumber
// 		ReferenceCode:            "00009042",   //TODO: check with workday team -
// 		CompanyDiscretionaryData: "                9042",
// 		StandardEntryClassCode:   "CCD",
// 		CompanyId:                "9893476201",
// 	}
// }
