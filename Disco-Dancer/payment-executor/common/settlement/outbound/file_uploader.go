package settlement

import (
	"context"

	"geico.visualstudio.com/Billing/plutus/enums"

	filmapiclient "geico.visualstudio.com/Billing/plutus/filmapi-client"
)

var Outboundfileprocessingpath string

type FileUploader struct {
}

func (FileUploader) UploadFile(paymentRequestType enums.PaymentRequestType, fileContent string) error {

	// TODO - paymentRequestType is coming as "All"
	// Need to discuss with the team about how to handle all when files generated will be different for each tenants
	// For Now, for testing, hardcoding the below file path. Uncomment the below code based on discussion

	// configHandler := commonFunctions.NewConfigHandler()
	// switch paymentRequestType {
	// case enums.insuranceautoauctions:
	// 	filePath = configHandler.GetString("outboundfileprocessingpath.insuranceautoauctions", "billing/paymentplatform/fileprocessing/JPMC/")

	// 	//TODO - For testing purpose appending the file name with current date
	// 	filePath = filePath + "GEICOCORP.ACH.NACHA.ARX.PGP." + time.Now().Format("20060102150405")
	// }

	//TODO - For testing purpose appending the file name with current date
	var filePath = Outboundfileprocessingpath + "GPP.ACH.NACHA.TXT"
	err := filmapiclient.UploadFile(filePath, fileContent)

	if err != nil {
		log.Error(context.Background(), err, "error uploading file using film api at path: %s", filePath)
	}

	return nil
}
