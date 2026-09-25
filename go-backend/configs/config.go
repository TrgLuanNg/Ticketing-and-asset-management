package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SSO      SSOConfig      `mapstructure:"sso"`
	AutoSync AutoSyncConfig `mapstructure:"auto_sync"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	Mode     string `mapstructure:"mode"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type DatabaseConfig struct {
	Driver    string          `mapstructure:"driver"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	SQLServer SQLServerConfig `mapstructure:"sqlserver"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		p.Host, p.User, p.Password, p.DBName, p.Port, p.SSLMode)
}

type SQLServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

func (s SQLServerConfig) DSN() string {
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustServerCertificate=true",
		s.User, s.Password, s.Host, s.Port, s.DBName)
}

type JWTConfig struct {
	SecretKey         string `mapstructure:"secret_key"`
	ExpirationMinutes int    `mapstructure:"expiration_minutes"`
}

type SSOConfig struct {
	Host       string `mapstructure:"host"`
	IssuerHost string `mapstructure:"issuer_host"`
	ClientID   string `mapstructure:"client_id"`
	SecretKey  string `mapstructure:"secret_key"`
}

type AutoSyncConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	CronSchedule string `mapstructure:"cron_schedule"`
}

var AppConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = &cfg
	return &cfg, nil
}
