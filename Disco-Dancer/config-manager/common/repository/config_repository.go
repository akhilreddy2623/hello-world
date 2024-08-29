package repository

import (
	"context"
	"errors"
	"fmt"

	"geico.visualstudio.com/Billing/plutus/config-manager-common/models/db"
	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/geico-private/pv-bil-frameworks/logging"
)

const (
	unhandledExceptionOccurred = "unhandled exception occurred"
)

var repositoryLog = logging.GetLogger("config-manager-repository")

//go:generate mockery --name ConfigRepositoryInterface --output ./mocks/ --filename mock_config_repository.go
type ConfigRepositoryInterface interface {
	GetConfig(applicationName string, environment string) (*[]db.ConfigResponse, error)
}

type ConfigRepository struct {
}

func (c *ConfigRepository) GetConfig(applicationName string, environment string) (*[]db.ConfigResponse, error) {
	var configResponses []db.ConfigResponse

	query := c.buildConfigQuery(applicationName, environment)
	rows, err := database.NewDbContext().Database.Query(context.Background(), query)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in getting Config Details from Configuration table")
		return nil, errors.New(unhandledExceptionOccurred)
	}
	defer rows.Close()

	for rows.Next() {
		var config db.ConfigResponse
		if err := rows.Scan(&config.ConfigId, &config.Key, &config.Value, &config.Description, &config.Scope, &config.Environment, &config.Application, &config.Tenant, &config.Product, &config.Vendor); err != nil {
			repositoryLog.Error(context.Background(), err, "error in scanning Config Details from Configuration table")
			return nil, errors.New(unhandledExceptionOccurred)
		}
		configResponses = append(configResponses, config)
	}

	if err := rows.Err(); err != nil {
		repositoryLog.Error(context.Background(), err, "error in iterating over rows from Configuration table")
		return nil, errors.New(unhandledExceptionOccurred)
	}

	return &configResponses, nil
}

func (c *ConfigRepository) buildConfigQuery(appName string, env string) string {
	baseQuery := `
        SELECT 
            config."ConfigId", 
            config."Key", 
            config."Value", 
            config."Description", 
            scopeLevel."Name" as "Scope", 
            config."Environment", 
            config."Filters"->>'Application' as "Application",
            config."Filters"->>'Tenant' as "Tenant",
            config."Filters"->>'Product' as "Product",
            config."Filters"->>'Vendor' as "Vendor"           
        FROM 
            configuration as config
        INNER JOIN 
            configuration_scope_level as scopeLevel
        ON 
            config."ScopeLevelId" = scopeLevel."ScopeLevelId"
        WHERE
            scopeLevel."Name" = 'Global'
    `

	if appName != "" {
		appNameFilter := fmt.Sprintf(`OR config."Filters"->>'Application' = '%s'`, appName)
		baseQuery += appNameFilter
	}

	if env != "" {
		envFilter := fmt.Sprintf(`AND config."Environment" = '%s'`, env)
		baseQuery += envFilter
	}

	return baseQuery
}
