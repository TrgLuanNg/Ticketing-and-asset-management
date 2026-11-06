package database

import (
	"fmt"
	"strings"
	"time"

	"go-backend/configs"
	"go-backend/pkg/logger"
	"go-backend/pkg/utils"

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
		ensurePasswordSchemaAndHashes(db, driver)
	}

	DB = db
	return db, nil
}

// ensurePasswordSchemaAndHashes ensures the user_info.pass column is wide enough for 60-char bcrypt hashes
// and migrates any legacy non-bcrypt password records in-place to bcrypt hashes.
func ensurePasswordSchemaAndHashes(db *gorm.DB, driver string) {
	if driver == "" || driver == "postgres" || driver == "postgresql" {
		var maxLen int
		err := db.Raw(`
			SELECT COALESCE(character_maximum_length, 0)
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'user_info' AND column_name = 'pass'
		`).Scan(&maxLen).Error
		if err == nil && maxLen > 0 && maxLen < 255 {
			logger.Infof("Upgrading user_info.pass column from VARCHAR(%d) to VARCHAR(255) for bcrypt hashes...", maxLen)
			_ = db.Exec(`DROP VIEW IF EXISTS "USERINFO"`).Error
			_ = db.Exec(`ALTER TABLE user_info ALTER COLUMN pass TYPE VARCHAR(255)`).Error
			_ = db.Exec(`
				CREATE OR REPLACE VIEW "USERINFO" AS
				SELECT
					user_id AS "UserID",
					pass AS "Pass",
					user_name AS "UserName",
					description AS "Description",
					area AS "Area",
					role AS "Role",
					is_active AS "IsActive",
					failed_login_attempts AS "FailedLoginAttempts",
					lockout_end AS "LockoutEnd"
				FROM user_info
			`).Error
		}
	}

	type userPassRow struct {
		UserID string `gorm:"column:UserID"`
		Pass   string `gorm:"column:Pass"`
	}
	var rows []userPassRow
	if err := db.Table("USERINFO").Select(`"UserID", "Pass"`).Find(&rows).Error; err != nil {
		return
	}

	hasher := utils.NewPasswordHasher()
	for _, r := range rows {
		if r.Pass != "" && !utils.IsBcryptHash(r.Pass) {
			if hashed, err := hasher.HashPassword(r.Pass); err == nil {
				_ = db.Table("USERINFO").Where(`"UserID" = ?`, r.UserID).Update("Pass", hashed).Error
				logger.Infof("Migrated user %s password to salted bcrypt hash", r.UserID)
			}
		}
	}
}
