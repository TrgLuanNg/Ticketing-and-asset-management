package database

import (
	"fmt"
	"strings"
	"time"

	"go-backend/configs"
	"go-backend/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *configs.Config) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if cfg.Server.Mode == "debug" {
		logLevel = gormlogger.Info
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	}

	var dialector gorm.Dialector

	driver := strings.ToLower(cfg.Database.Driver)
	switch driver {
	case "sqlserver", "mssql":
		dsn := cfg.Database.SQLServer.DSN()
		logger.Infof("Connecting to SQL Server database: %s:%d/%s",
			cfg.Database.SQLServer.Host, cfg.Database.SQLServer.Port, cfg.Database.SQLServer.DBName)
		dialector = sqlserver.Open(dsn)
	case "postgres", "postgresql":
		fallthrough
	default:
		dsn := cfg.Database.Postgres.DSN()
		logger.Infof("Connecting to PostgreSQL database: %s:%d/%s",
			cfg.Database.Postgres.Host, cfg.Database.Postgres.Port, cfg.Database.Postgres.DBName)
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		logger.Errorf("Warning: Database ping failed: %v", err)
	} else {
		logger.Info("Database connection established successfully!")
	}

	DB = db
	return db, nil
}
