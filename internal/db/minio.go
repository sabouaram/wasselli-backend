package db

import (
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Minio interface {
	//TODO
}

type MinioClient struct {
	client *minio.Client
	logger *zap.Logger
}

func NewMinioClient(cfg *viper.Viper, logger *zap.Logger) (*MinioClient, error) {

	if cfg == nil || logger == nil {
		return nil, errors.New("minio client instances arguments are nil")
	}

	var (
		endpoint  = cfg.GetString("s3.minio.endpoint")
		accessKey = cfg.GetString("s3.minio.access-key")
		secretKey = cfg.GetString("s3.minio.secret-key")
		useSSL    = cfg.GetBool("s3.minio.ssl")
	)

	logger.Info("minio client instanced")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		logger.Error("minio client instance error", zap.Any("error =>", err))
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	logger.Info("minio client instanced created")

	return &MinioClient{client: client, logger: logger}, nil
}
