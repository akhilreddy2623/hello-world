package repository

import (
	"context"
	"errors"

	"geico.visualstudio.com/Billing/plutus/database"
)

const (
	user_exist = `SELECT EXISTS(SELECT 1 FROM public.Persona WHERE "UserId"=$1)`
)

func IsUserExists(userId string) error {

	var exists bool
	err := database.NewDbContext().Database.QueryRow(context.Background(), user_exist, userId).Scan(&exists)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in checking existence of user details")
		return errors.New(unhandledExceptionOccurred)
	}

	if !exists {
		return errors.New("user ID does not exist")
	}
	return nil

}
