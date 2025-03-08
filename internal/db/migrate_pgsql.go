package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// atlas command should be installed => curl -sSf https://atlasgo.sh | sh

func MigratePGSQL(cfg *viper.Viper, enable bool, logger *zap.Logger) error {

	if cfg == nil || logger == nil {
		return fmt.Errorf("config or logger instance is nil")
	}

	if !enable {
		log.Println("migration disabled")
		return nil
	}

	var (
		host     = cfg.GetString("storage.db.postgresql.host")
		port     = cfg.GetString("storage.db.postgresql.port")
		user     = cfg.GetString("storage.db.postgresql.user")
		password = cfg.GetString("storage.db.postgresql.password")
		pgSchema = cfg.GetString("storage.db.postgresql.schema")
		database = cfg.GetString("storage.db.postgresql.database")
		db       *sql.DB
		driver   migrate.Driver
		existing *schema.Realm
		desired  schema.Realm
		diff     []schema.Change
		hcl      []byte
		err      error
	)

	logger.Info("pgsql schema migration started")

	ctx := context.Background()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		database,
	)

	if db, err = sql.Open("postgres", dsn); err != nil {
		logger.Error("opening connection to postgresql failed: ", zap.Any("error =>", err))
		return fmt.Errorf("opening connection to PostgreSQL failed: %v", err)
	}

	driver, err = postgres.Open(db)

	if err != nil {
		logger.Error("failed to connect to postgresql: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	existing, err = driver.InspectRealm(
		ctx,
		&schema.InspectRealmOption{Schemas: []string{pgSchema}},
	)

	if err != nil {
		logger.Error("failed to inspect existing schema: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to inspect existing schema: %w", err)
	}

	//Read and parse the target schema from HCL file
	_, currentFile, _, _ := runtime.Caller(0)

	migPath := filepath.Join(filepath.Dir(currentFile), "pgsql_migration/schema.hcl")

	hcl, err = os.ReadFile(migPath)

	if err != nil {
		logger.Error("failed to read HCL file: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to read HCL file: %w", err)
	}

	if err = postgres.EvalHCLBytes(hcl, &desired, nil); err != nil {
		logger.Error("failed to evaluate target schema: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to evaluate target schema: %w", err)
	}

	// Step 4: Compare the existing and desired schemas
	diff, err = driver.RealmDiff(existing, &desired)

	if err != nil {
		logger.Error("failed to calculate schema diff: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to calculate schema diff: %w", err)
	}

	if len(diff) == 0 {
		logger.Info("no schema changes detected; schemas are identical.")
		return nil
	}

	//Apply the new schema from the HCL file
	logger.Info("Applying new schema...")

	if err = driver.ApplyChanges(ctx, diff); err != nil {
		logger.Error("failed to apply new schema: ", zap.Any("error =>", err))
		return fmt.Errorf("failed to apply new schema: %w", err)
	}

	logger.Info("Migration completed successfully.")

	return nil
}
