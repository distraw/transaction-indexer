package cli

import (
	"github.com/distraw/transaction-indexer/assets"
	"github.com/distraw/transaction-indexer/internal/config"
	migrate "github.com/rubenv/sql-migrate"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: assets.Migrations,
	Root:       "migrations",
}

func MigrateUp(cfg config.Config) error {
	applied, err := migrate.Exec(cfg.DB().RawDB(), "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to migrate DB up")
	}
	cfg.Log().WithField("applied", applied).Info("migrated DB up successfully")
	return nil
}

func MigrateDown(cfg config.Config) error {
	applied, err := migrate.Exec(cfg.DB().RawDB(), "postgres", migrations, migrate.Down)
	if err != nil {
		return errors.Wrap(err, "failed to migrate DB down")
	}
	cfg.Log().WithField("applied", applied).Info("migrated DB down successfully")
	return nil
}
