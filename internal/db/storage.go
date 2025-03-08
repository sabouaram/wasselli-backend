package db

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Storage interface {
	//TODO
}

func NewStorage(cfg *viper.Viper, logger *zap.Logger) (Storage, error) {

	if cfg == nil || logger == nil {
		return nil, errors.New("storage instances arguments are nil")
	}

	var storageType = cfg.GetString("storage.db.type")

	switch storageType {
	case "postgresql":
		return NewPGSQLStorage(cfg, logger)
	default:
		return nil, fmt.Errorf("storage type %v is not supported", storageType)
	}
}
