package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	MigrationDir string
	DBConfigFile string
	EnvFile      string
}

type DBConfig struct {
	Development struct {
		Dialect    string `yaml:"dialect"`
		DataSource string `yaml:"datasource"`
	} `yaml:"development"`
}

func NewConfig() *AppConfig {
	err := godotenv.Load(os.Getenv("ENV_FILE"))
	if err != nil {
		errEnv := godotenv.Load()
		if errEnv != nil {
			return nil
		}
	}

	return &AppConfig{
		MigrationDir: getEnv("MIGRATION_DIR", "migrations"),
		DBConfigFile: getEnv("DB_CONFIG_FILE", "dbconfig.yml"),
		EnvFile:      getEnv("ENV_FILE", ".env"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func LoadDBConfig(filename string) (*DBConfig, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", filename)
	}

	var config DBConfig

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func(file *os.File) {
		errFile := file.Close()
		if errFile != nil {
			return
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	return &config, nil
}

func ConnectDB(configFile string) (*sql.DB, error) {
	dbConfig, err := LoadDBConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading database config: %w", err)
	}

	connStr := os.ExpandEnv(dbConfig.Development.DataSource)

	db, err := sql.Open(dbConfig.Development.Dialect, connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60)

	log.Printf("Connected to database using config: %s", configFile)
	return db, nil
}

func ApplyMigration(db *sql.DB, migrationDir string) error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationDir,
	}

	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		return fmt.Errorf("migration directory not found: %s", migrationDir)
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("error applying migrations: %w", err)
	}

	log.Printf("Successfully applied %d migration(s) from %s", n, migrationDir)
	return nil
}

func ApplyMigrationWithConfig(db *sql.DB, config *AppConfig) error {
	return ApplyMigration(db, config.MigrationDir)
}
