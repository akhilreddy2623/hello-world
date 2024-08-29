package inbound_fileprocessor

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/jackc/pgx/v5"
)

const (
	user = "inbound-fileprocessor"
)

func isFileHashPresent(fileHash string) (bool, error) {
	isFileHashPresent := true

	isFileHashPresentQuery := `SELECT EXISTS
                             (SELECT 1 
							  FROM public.file_deduplication 
							  WHERE "FileHash" = $1 AND "CreatedDate" > current_date - interval '3' day)`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		isFileHashPresentQuery,
		fileHash)
	err := row.Scan(&isFileHashPresent)
	if err != nil {
		log.Error(context.Background(), err, "error fetching file hash from file_deduplication table")
		return isFileHashPresent, err

	}

	return isFileHashPresent, nil
}

func isFileHashWithStatusPresent(fileHash string, status string) (bool, error) {
	isFileHashWithStatusPresent := true

	isFileHashWithStatusPresentQuery := `SELECT EXISTS
										(SELECT 1 
										FROM public.file_deduplication 
										WHERE "FileHash" = $1 AND "Status" = $2 AND "CreatedDate" > current_date - interval '3' day)`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		isFileHashWithStatusPresentQuery,
		fileHash,
		status)
	err := row.Scan(&isFileHashWithStatusPresent)
	if err != nil {
		log.Error(context.Background(), err, "error fetching file hash with status from file_deduplication table")
		return isFileHashWithStatusPresent, err

	}

	return isFileHashWithStatusPresent, nil
}

func addFileDeduplicationEntry(filePath string, businessFileType string, fileHash string, status string) (int, error) {

	var fileId int
	addFileDeduplicationQuery := `INSERT INTO public.file_deduplication
	("FileHash",
	 "FilePath",
	 "BusinessFileType", 
	 "Status", 
	 "CreatedDate",
	 "CreatedBy",
	 "UpdatedDate",
	 "UpdatedBy"
	)
	VALUES(
	$1, $2, $3,	$4, $5, $6, $7,	$8
	)
	RETURNING "FileId"`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		addFileDeduplicationQuery,
		fileHash,
		filePath,
		businessFileType,
		status,
		time.Now(),
		user,
		time.Now(),
		user)

	err := row.Scan(&fileId)
	if err != nil {
		log.Error(context.Background(), err, "error inserting file_deduplication table")
		return fileId, err
	}

	return fileId, nil
}

func updateFileDeduplicationEntry(fileHash string, status string) (int, error) {
	var fileId int
	updateFileDeduplicationQuery := `UPDATE public.file_deduplication
	                                 SET "Status"=$1, "UpdatedDate"=$2
									 WHERE "FileHash" = $3 AND "CreatedDate" > current_date - interval '3' day
									 RETURNING "FileId"`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		updateFileDeduplicationQuery,
		status,
		time.Now(),
		fileHash)

	err := row.Scan(&fileId)
	if err != nil {
		log.Error(context.Background(), err, "error updating file_deduplication table")
		return fileId, err
	}

	return fileId, nil
}

func updateFileDeduplicationEntryForGivenFileId(transaction pgx.Tx, fileId int, status string) error {

	updateFileStatusQuery := `UPDATE public.file_deduplication
	                            SET "Status"=$1, "UpdatedDate"=$2
								WHERE "FileId" = $3`
	_, err := transaction.Exec(
		context.Background(),
		updateFileStatusQuery,
		status,
		time.Now(),
		fileId)

	if err != nil {
		log.Error(context.Background(), err, "error updating file_deduplication table")
		return err
	}

	return nil
}
