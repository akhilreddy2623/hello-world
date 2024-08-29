package database

import (
	"context"
	"fmt"
	"sync"

	"geico.visualstudio.com/Billing/plutus/logging"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbContext struct {
	Database *pgxpool.Pool
}

var connectionString string

var log = logging.GetLogger("DbContext")
var once sync.Once
var dbContext *DbContext

func Init(configHandler config.AppConfiguration) {
	host := configHandler.GetString("PaymentPlatform.Db.Host", "")
	port := configHandler.GetInt("PaymentPlatform.Db.Port", 0)
	user := configHandler.GetString("PaymentPlatform.Db.UserName", "")
	password := configHandler.GetString("PaymentPlatform.Db.Password", "")
	dbname := configHandler.GetString("PaymentPlatform.Db.Dbname", "")

	// Patch for non migrated code
	//TODO: remove the patch after all the code has been migrated
	if host == "" {
		host = configHandler.GetString("db.host", "")
	}
	if port == 0 {
		port = configHandler.GetInt("db.port", 0)
	}
	if user == "" {
		user = configHandler.GetString("db.user", "")
	}
	if password == "" {
		password = configHandler.GetString("db.password", "")
	}
	if dbname == "" {
		dbname = configHandler.GetString("db.dbname", "")
	}

	connectionString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

func NewDbContext() *DbContext {
	once.Do(func() {
		db, err := pgxpool.New(context.Background(), connectionString)
		if err != nil {
			log.Error(context.Background(), err, "Unable to create connection pool.")
			panic(err)
		}
		dbContext = &DbContext{Database: db}
	})

	return dbContext
}

// Used by the integration tests to switch database contexts.
func NewPgxPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Error(context.Background(), err, "Unable to create connection pool.")
		panic(err)
	}
	return pool
}

func GetDbContext() *DbContext {
	return dbContext
}

func SetDbContext(dbc DbContext) {
	dbContext = &dbc
}
