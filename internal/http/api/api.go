package api

import (
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"wasselli-backend/emailing"

	"wasselli-backend/internal/db"
	"wasselli-backend/internal/http/api/handlers"
)

func NewAPIHandler(
	cfg *viper.Viper,
	stg db.Storage,
	logger *zap.Logger,
) (*handlers.Handler, error) {
	var (
		emailSvc *emailing.EmailService
		minio    db.Minio
		err      error
	)

	if logger == nil || cfg == nil || stg == nil {
		return nil, errors.New("api handler instances arguments are nil")
	}

	emailSvc, err = emailing.New(cfg, logger)

	if err != nil || emailSvc == nil {
		return nil, fmt.Errorf("email svc error %v", err)
	}

	minio, err = db.NewMinioClient(cfg, logger)

	if err != nil || minio == nil {
		return nil, fmt.Errorf("minio svc error %v", err)
	}

	return &handlers.Handler{
		Mux:       chi.NewMux(),
		Emailing:  emailSvc,
		Config:    cfg,
		Validator: validator.New(),
		Minio:     minio,
		Logger:    logger,
		Storage:   stg,
	}, nil

}
