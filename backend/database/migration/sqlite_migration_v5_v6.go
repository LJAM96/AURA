package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

// migrate_5_to_6 adds an "edition" column to MediaItems, SavedItems, and
// IgnoredItems, and folds it into each table's uniqueness constraint so that
// multiple editions of the same TMDB item (e.g. Theatrical vs Director's Cut)
// no longer collide with each other. Existing rows get edition = ”.
func migrate_5_to_6(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v5 to v6", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 5).Int("To Version", 6).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	backupErr := database.Backup(ctx, 5, 6)
	if backupErr.Message != "" {
		return backupErr
	}

	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	editionColumnExists, checkColumnErr := checkColumnExists(ctx, "MediaItems", "edition")
	if checkColumnErr.Message != "" {
		return checkColumnErr
	}

	if !editionColumnExists {
		// A v5 database cannot have a SavedItems.priority column, so there is
		// nothing to carry across the rebuild
		editionErr := applyEditionSchema(ctx, conn, logAction, false)
		if editionErr.Message != "" {
			return editionErr
		}
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v5.0 to v6.0 completed successfully")
	return Err
}
