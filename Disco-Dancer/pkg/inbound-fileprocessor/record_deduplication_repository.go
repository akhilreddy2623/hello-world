package inbound_fileprocessor

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/jackc/pgx/v5"
)

const (
	updateRecordStatusQuery = `UPDATE public.record_deduplication
	                            SET "Status"=$1, "UpdatedDate"=$2
								WHERE "RecordId" = $3 AND "FileId" = $4`
)

func isRecordHashPresent(recordHash string, fileId int) bool {
	isRecordHashPresent := true

	isRecordHashPresentQuery := `SELECT EXISTS
                             (SELECT 1 
							  FROM public.record_deduplication 
							  WHERE "RecordHash" = $1 AND "FileId" = $2)`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		isRecordHashPresentQuery,
		recordHash,
		fileId)
	err := row.Scan(&isRecordHashPresent)
	if err != nil {
		log.Error(context.Background(), err, "error fetching record hash from record_deduplication table")
	}
	return isRecordHashPresent
}

func addRecordDeduplicationEntry(transaction pgx.Tx, fileId int, recordHash string, status string, rowNumber int) int64 {

	var recordId int64
	addRecordDeduplicationQuery := `INSERT INTO public.record_deduplication
	("FileId",
	 "RecordHash",
	 "Status", 
	 "RowNumber", 
	 "CreatedDate",
	 "CreatedBy",
	 "UpdatedDate",
	 "UpdatedBy"
	)
	VALUES(
	$1, $2, $3,	$4, $5, $6, $7,	$8
	)
	RETURNING "RecordId"`

	row := transaction.QueryRow(
		context.Background(),
		addRecordDeduplicationQuery,
		fileId,
		recordHash,
		status,
		rowNumber,
		time.Now(),
		user,
		time.Now(),
		user)

	err := row.Scan(&recordId)
	if err != nil {
		log.Error(context.Background(), err, "error inserting record_deduplication table")
	}

	return recordId
}

func updateRecordDeduplicationEntry(transaction pgx.Tx, recordId int64, fileId int, status string) error {

	_, err := transaction.Exec(
		context.Background(),
		updateRecordStatusQuery,
		status,
		time.Now(),
		recordId,
		fileId)

	if err != nil {
		log.Error(context.Background(), err, "error updating record_deduplication table")
		return err
	}

	return nil
}

func updateRecordDeduplicationEntryWithNoTransaction(recordId int64, fileId int, status string) error {

	_, err := database.NewDbContext().Database.Exec(
		context.Background(),
		updateRecordStatusQuery,
		status,
		time.Now(),
		recordId,
		fileId)

	if err != nil {
		log.Error(context.Background(), err, "error updating record_deduplication table")
		return err
	}

	return nil
}

func getRecordCountWithSpecificStatus(fileId int, status string) (int, error) {
	recordCount := 0

	getRecordCountForGivenFileIdQuery := `SELECT COUNT("RecordId") 
										 FROM public.record_deduplication 
										 WHERE "FileId" = $1 AND "Status" = $2`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getRecordCountForGivenFileIdQuery,
		fileId,
		status)
	err := row.Scan(&recordCount)
	if err != nil {
		log.Error(context.Background(), err, "error fetching count of records from record_deduplication table")
		return recordCount, err

	}

	return recordCount, nil
}
