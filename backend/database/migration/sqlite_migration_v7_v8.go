package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_7_to_8(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v7 to v8", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 7).Int("To Version", 8).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 7, 8)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Add priority column to SavedItems table (0 = not from subscription, 1+ = subscription priority)
	addPriorityQuery := `ALTER TABLE SavedItems ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;`
	_, err := conn.ExecContext(ctx, addPriorityQuery)
	if err != nil {
		logAction.SetError("Failed to add priority column to SavedItems table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": addPriorityQuery,
		})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v7.0 to v8.0 completed successfully")
	return Err
}
