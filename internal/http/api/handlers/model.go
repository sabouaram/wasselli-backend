package handlers

import (
	"wasselli-backend/emailing"
	"wasselli-backend/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Handler struct {
	Mux       *chi.Mux
	Config    *viper.Viper
	Storage   db.Storage
	Minio     db.Minio
	Validator *validator.Validate
	Emailing  *emailing.EmailService
	Logger    *zap.Logger
}
