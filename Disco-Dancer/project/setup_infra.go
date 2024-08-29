package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/geico-private/pv-bil-frameworks/config"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v3"
)

var (
	administratorConfig config.AppConfiguration
	executorConfig      config.AppConfiguration
	taskmgrconfig       config.AppConfiguration
	vaultconfig         config.AppConfiguration
)

func main() {
	administratorConfig = config.NewConfigBuilder().
		AddJsonFile("./payment-administrator/worker/config/appsettings.json").
		AddJsonFile("./payment-administrator/worker/config/secrets.json").Build()

	executorConfig = config.NewConfigBuilder().
		AddJsonFile("./payment-executor/worker/config/appsettings.json").
		AddJsonFile("./payment-executor/worker/config/secrets.json").Build()

	taskmgrconfig = config.NewConfigBuilder().
		AddJsonFile("./task-manager/worker/config/appsettings.json").
		AddJsonFile("./task-manager/worker/config/secrets.json").Build()

	vaultconfig = config.NewConfigBuilder().
		AddJsonFile("./payment-vault/worker/config/appsettings.json").
		AddJsonFile("./payment-vault/worker/config/secrets.json").Build()

	err := setupPostgresTables()
	if err != nil {
		panic(err)
	}
	err = setupKafkaTopics(administratorConfig.GetString("PaymentPlatform.Kafka.Brokers", ""), "./project/kafkaconfiguration/component.dv.yaml")
	if err != nil {
		panic(err)
	}
}

func setupKafkaTopics(brokers, topicFilePath string) error {
	conn, err := kafka.Dial("tcp", brokers)
	if err != nil {
		return fmt.Errorf("failed to dial broker: %w", err)
	}

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get the controller: %w", err)
	}
	conn.Close()

	fmt.Printf("Controller: %+v\n", controller)
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to dial the controller: %w", err)
	}
	defer controllerConn.Close()

	yamlConfigBytes, err := os.ReadFile(topicFilePath)
	if err != nil {
		return err
	}
	var configFile map[string]any
	err = yaml.Unmarshal([]byte(yamlConfigBytes), &configFile)
	if err != nil {
		return err
	}
	department, ok := configFile["department"].(string)
	if !ok {
		return errors.New("invalid yaml config")
	}

	topicConfigs, ok := configFile["topics"].([]any)
	if !ok {
		return errors.New("invalid yaml config")
	}
	createConfigs := make([]kafka.TopicConfig, 0)
	topicNames := []string{}

	for _, t := range topicConfigs {
		topic, ok := t.(map[string]any)
		if !ok {
			return errors.New("invalid kafka topic config")
		}
		eventType, ok := topic["eventType"].(string)
		if !ok {
			return errors.New("invalid kafka event type")
		}
		topicName, ok := topic["topicName"].(string)
		if !ok {
			return errors.New("invalid kafka topic name")
		}
		fullName := fmt.Sprintf("%s.%s.%s", department, eventType, topicName)
		fmt.Println("parsed topic: ", topicName)
		err = controllerConn.DeleteTopics(fullName)
		if err != nil && !strings.Contains(err.Error(), "Unknown Topic") {
			fmt.Println("failed to detele topics ", fullName, err)
		}
		topicNames = append(topicNames, fullName)
		createConfigs = append(createConfigs, kafka.TopicConfig{
			Topic:             fullName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}
	time.Sleep(3 * time.Second)
	return controllerConn.CreateTopics(createConfigs...)
}

func setupPostgresTables() error {
	testDB, err := createDefaultDB(administratorConfig, "postgresdb")
	if err != nil {
		return err
	}
	defer testDB.Close()

	dbNames := []string{"payment_administrator", "payment_executor", "task_manager", "payment_vault"}
	for _, db := range dbNames {
		stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", db)
		_, err = testDB.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to drop database %s: %w", db, err)
		}
		_, err = testDB.Exec(fmt.Sprintf("CREATE DATABASE %s", db))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", db, err)
		}
	}

	err = initTables(administratorConfig, "./payment-administrator/common/database-schema/")
	if err != nil {
		return fmt.Errorf("failed to initialize tables for payment_administrator: %w", err)
	}

	err = initTables(executorConfig, "./payment-executor/common/database-schema")
	if err != nil {
		return fmt.Errorf("failed to initialize tables for payment_executor: %w", err)
	}

	err = initTables(taskmgrconfig, "./task-manager/common/database-schema")
	if err != nil {
		return fmt.Errorf("failed to initialize tables for task_manager: %w", err)
	}

	err = initTables(vaultconfig, "./payment-vault/common/database-schema")
	if err != nil {
		return fmt.Errorf("failed to initialize tables for payment-vault: %w", err)
	}
	return nil
}

func initTables(conf config.AppConfiguration, sqlFilePath string) error {
	// switch to  database and initialize tables.
	dB, err := createDefaultDB(conf, conf.GetString("PaymentPlatform.Db.Dbname", ""))

	if err != nil {
		return err
	}

	files, err := os.ReadDir(sqlFilePath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			b, err := os.ReadFile(filepath.Join(sqlFilePath, f.Name()))
			if err != nil {
				return err
			}
			stmt := string(b)
			_, err = dB.Exec(stmt)
			if err != nil {
				return err
			}
			fmt.Println("SQL file Executed: ", f.Name())
		}
	}
	return nil
}

func createDefaultDB(conf config.AppConfiguration, dbName string) (*sql.DB, error) {
	// Create databases from defaultDB connection, in this defaultDB is postgresdb
	host := conf.GetString("PaymentPlatform.Db.Host", "")
	port := conf.GetInt("PaymentPlatform.Db.Port", 0)
	user := conf.GetString("PaymentPlatform.Db.UserName", "")
	password := conf.GetString("PaymentPlatform.Db.Password", "")
	connectionString := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`, host, port, user, password, dbName)
	testDB, err := sql.Open("postgres", connectionString)
	return testDB, err
}
