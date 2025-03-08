package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type PGSQLStorage struct {
	DbConnection *sql.DB
	Schema       string
	Tables       map[string]string
	Logger       *zap.Logger
}

func NewPGSQLStorage(cfg *viper.Viper, logger *zap.Logger) (PGSQLStorage, error) {
	var (
		host     = cfg.GetString("storage.db.postgresql.host")
		port     = cfg.GetString("storage.db.postgresql.port")
		user     = cfg.GetString("storage.db.postgresql.user")
		password = cfg.GetString("storage.db.postgresql.password")
		database = cfg.GetString("storage.db.postgresql.database")
		db       *sql.DB
		err      error
	)

	logger.Info("pgsql storage instanced")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable dbname=%s",
		host,
		port,
		user,
		password,
		database,
	)

	if db, err = sql.Open("postgres", dsn); err != nil {
		logger.Error("pgsql storage instance error", zap.Any("error =>", err))
		return PGSQLStorage{}, fmt.Errorf("connection to PostgreSQL failed")
	}

	if db.Ping() != nil {
		logger.Error("pgsql storage instance error", zap.Any("error =>", err))
		return PGSQLStorage{}, fmt.Errorf("connection to PostgreSQL database failed")
	}

	return PGSQLStorage{
		DbConnection: db,
		Logger:       logger,
		Tables:       map[string]string{
			//Example: "accounts":          cfg.GetString("storage.db.postgresql.tables.accounts"),
		},
	}, nil
}
