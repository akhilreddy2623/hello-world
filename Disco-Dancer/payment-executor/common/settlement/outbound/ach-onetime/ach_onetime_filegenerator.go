package AchOnetime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
)

var (
	lineBuilder strings.Builder
	log         = logging.GetLogger("ach-onetime-filegenerator")
)

type DataHolder struct {
	EntryDetailCount int
	EntryHashTotal   string
	FileInfo         AchOnetimeFileData
	FileType         enums.PaymentRequestType
}

func GenerateAchOnetimeFile(settledAchPayments []app.SettlementPayment) (string, error) {

	log.Info(context.Background(), "starting ach onetime file creation")

	fileLines, err := createFileLines(settledAchPayments)
	if err != nil {
		log.Error(context.Background(), err, "error in creating file lines")
		return "", err
	}

	log.Info(context.Background(), "End of ach onetime file creation")

	return fileLines, nil
}

func (d *DataHolder) createFileHeader() string {

	fileCreationDate := time.Now().Format("060102")
	fileCreationTime := time.Now().Format("1504")
	d.FileInfo = GetFileInfo(enums.InsuranceAutoAuctions)

	lineBuilder.WriteString(PadInt(int(enums.FileHeader), 1))
	lineBuilder.WriteString(PriorityCode)
	lineBuilder.WriteString(d.FileInfo.ImmediateDestination)
	lineBuilder.WriteString(d.FileInfo.ImmediateOrigin)
	lineBuilder.WriteString(fileCreationDate)
	lineBuilder.WriteString(fileCreationTime)
	lineBuilder.WriteString(FileIdModifiler)
	lineBuilder.WriteString(RecordSize)
	lineBuilder.WriteString(BlockingFactor)
	lineBuilder.WriteString(FormatCode)
	lineBuilder.WriteString(ImmediateDestinationName)
	lineBuilder.WriteString(ImmediateOriginName)
	lineBuilder.WriteString(d.FileInfo.ReferenceCode)

	fileHeader := lineBuilder.String()
	lineBuilder.Reset()
	return fileHeader
}

func (d *DataHolder) createBatchHeader(batchIndex int) string {
	paymentDate := time.Now().Format("060102")

	lineBuilder.WriteString(PadInt(int(enums.BatchHeader), 1))
	lineBuilder.WriteString(ServiceClass)
	lineBuilder.WriteString(CompanyName)
	lineBuilder.WriteString(d.FileInfo.CompanyDiscretionaryData)
	lineBuilder.WriteString(d.FileInfo.CompanyId)
	lineBuilder.WriteString(d.FileInfo.StandardEntryClassCode)
	lineBuilder.WriteString(CompanyEntryDescription)
	lineBuilder.WriteString(paymentDate)
	lineBuilder.WriteString(paymentDate)
	lineBuilder.WriteString("   ")
	lineBuilder.WriteString("1")
	lineBuilder.WriteString(RoutingNumber)
	lineBuilder.WriteString(fmt.Sprintf("%07d", batchIndex))

	batchHeader := lineBuilder.String()
	lineBuilder.Reset()

	return batchHeader
}

func (d *DataHolder) createEntryDetailRecord(settledAchPayments []app.SettlementPayment) []string {

	var entryDetailRecords []string

	for _, payment := range settledAchPayments {
		d.EntryDetailCount++

		lineBuilder.WriteString(PadInt(int(enums.EntryDetail), 1))
		lineBuilder.WriteString(PadString(GetTransactionCode(enums.Checking), 2))
		lineBuilder.WriteString(PadString(payment.RoutingNumber, 9))
		lineBuilder.WriteString(PadString(payment.AccountIdentifier, 17))
		lineBuilder.WriteString(FormatAndPadAmount(payment.Amount, 10))
		lineBuilder.WriteString(payment.SettlementIdentifier) //Identifier is "C" for consolidated "P" for payment
		lineBuilder.WriteString(PadString(payment.AccountName, 22))
		lineBuilder.WriteString("  ")
		lineBuilder.WriteString("0") //Always Zero for Credit settlement file
		lineBuilder.WriteString(payment.SettlementIdentifier)

		entryDetailRecord := lineBuilder.String()

		entryDetailRecords = append(entryDetailRecords, entryDetailRecord)
		lineBuilder.Reset()

	}

	return entryDetailRecords
}

func (d *DataHolder) createBatchFooter(settledAchPayments []app.SettlementPayment, batchIndex int) (string, error) {

	entryHash, err := GetEntryHash(settledAchPayments)
	if err != nil {
		log.Error(context.Background(), err, "error in generating entry hash")
		return "", err
	}

	d.EntryHashTotal = entryHash

	lineBuilder.WriteString(PadInt(int(enums.BatchFooter), 1))
	lineBuilder.WriteString(ServiceClass)
	lineBuilder.WriteString(PadInt(d.EntryDetailCount, 6))
	lineBuilder.WriteString(d.EntryHashTotal)
	lineBuilder.WriteString("000000000000")
	lineBuilder.WriteString(ComputeTotalBatchCreditAmount(settledAchPayments))
	lineBuilder.WriteString(d.FileInfo.CompanyId)
	lineBuilder.WriteString("                         ")
	lineBuilder.WriteString(RoutingNumber)
	lineBuilder.WriteString(fmt.Sprintf("%07d", batchIndex))

	batchFooter := lineBuilder.String()
	lineBuilder.Reset()

	return batchFooter, nil
}

func (d *DataHolder) createFileFooter(settledAchPayments []app.SettlementPayment) string {

	lineBuilder.WriteString(PadInt(int(enums.Filefooter), 1))
	lineBuilder.WriteString("000001")
	lineBuilder.WriteString("000001")
	lineBuilder.WriteString(PadInt(d.EntryDetailCount, 8))
	lineBuilder.WriteString(d.EntryHashTotal)
	lineBuilder.WriteString("000000000000")
	lineBuilder.WriteString(ComputeTotalBatchCreditAmount(settledAchPayments))
	lineBuilder.WriteString("                                       ")

	batchFooter := lineBuilder.String()
	lineBuilder.Reset()

	return batchFooter
}

func createFileLines(settledAchPayments []app.SettlementPayment) (string, error) {

	var fileLines []string

	dh := DataHolder{}

	fileHeader := dh.createFileHeader()
	fileLines = append(fileLines, fileHeader)
	batchIndex := 1
	batchHeader := dh.createBatchHeader(batchIndex)
	fileLines = append(fileLines, batchHeader)
	entryDetails := dh.createEntryDetailRecord(settledAchPayments)
	fileLines = append(fileLines, entryDetails...)
	batchFooter, err := dh.createBatchFooter(settledAchPayments, batchIndex)
	if err != nil {
		log.Error(context.Background(), err, "error in creating ach one time batch footer")
		return "", err
	}

	fileLines = append(fileLines, batchFooter)
	fileFooter := dh.createFileFooter(settledAchPayments)
	fileLines = append(fileLines, fileFooter)
	fileLinesStr := strings.Join(fileLines, "\n")

	return fileLinesStr, nil
}
