package AchOnetime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
)

func ComputeTotalBatchCreditAmount(items []app.SettlementPayment) string {
	var total float64

	for _, item := range items {
		total += float64(item.Amount)
		total = math.Round(total*100) / 100
	}

	totalStr := fmt.Sprintf("%.2f", total)

	totalStr = strings.Replace(totalStr, ".", "", -1)

	if len(totalStr) < 12 {
		totalStr = fmt.Sprintf("%012s", totalStr)
	}
	return totalStr
}

func PadInt(value int, length int) string {

	paddedNum := fmt.Sprintf("%0*d", length, value)

	return paddedNum
}

func PadString(value string, length int) string {

	paddedStr := fmt.Sprintf("%-*s", length, value)

	return paddedStr
}

func FormatAndPadAmount(amount float32, length int) string {
	formattedAmount := fmt.Sprintf("%.2f", amount)

	formattedAmount = strings.Replace(formattedAmount, ".", "", -1)

	if len(formattedAmount) < length {
		return fmt.Sprintf("%0*s", length, formattedAmount)
	}
	return formattedAmount
}

func GetTransactionCode(accountType enums.ACHAccountType) string {

	if accountType == enums.Checking {
		return "22"
	}

	//If accountType is savings
	return "32"
}

func GetIndividualId(consolidationId int64, length int, identifierType string) string {

	//Get the current date
	currentDate := time.Now()

	//Prefix the consolidatedId with Zeros
	consolidatedIdStr := fmt.Sprintf("%08d", consolidationId)

	//Extract the last Eight of consolidatedId
	if len(consolidatedIdStr) > 8 {
		consolidatedIdStr = consolidatedIdStr[len(consolidatedIdStr)-8:]
	}

	individualId := fmt.Sprintf("%s%s%s", currentDate.Format("060102"), consolidatedIdStr, identifierType)

	individualId = fmt.Sprintf("%0*s", length, individualId)

	return individualId
}

func GetTraceNumber(consolidationId int) string {

	//Get the current date
	currentDate := time.Now()

	// "2" is Consolidation indicator
	return fmt.Sprintf("%s%s%d", currentDate.Format("060102"), "2", consolidationId)

}

func GetEntryHash(payments []app.SettlementPayment) (string, error) {
	var total int64 = 0

	for _, num := range payments {

		num := num.RoutingNumber

		if len(num) == 9 {

			firstEight := num[:8]

			value, err := strconv.ParseInt(firstEight, 10, 64)
			if err != nil {
				return "0000000000", err
			}

			total += value

		}
	}

	entryHash := strconv.FormatInt(total, 10)

	if len(entryHash) < 10 {
		entryHash = fmt.Sprintf("%010s", entryHash)
	} else if len(entryHash) > 10 {
		return entryHash[len(entryHash)-10:], nil
	}

	return entryHash, nil

}
