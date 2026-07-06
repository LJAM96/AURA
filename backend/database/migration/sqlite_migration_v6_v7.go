package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_6_to_7(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v6 to v7", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 6).Int("To Version", 7).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 6, 7)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Add priority column to UserSubscriptions table (1 = highest priority, higher number = lower priority)
	addPriorityQuery := `ALTER TABLE UserSubscriptions ADD COLUMN priority INTEGER NOT NULL DEFAULT 1;`
	_, err := conn.ExecContext(ctx, addPriorityQuery)
	if err != nil {
		logAction.SetError("Failed to add priority column to UserSubscriptions table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": addPriorityQuery,
		})
		return *logAction.Error
	}

	// Add index on priority for sorting
	createPriorityIndexQuery := `CREATE INDEX IF NOT EXISTS idx_usersubscriptions_priority ON UserSubscriptions(priority);`
	_, err = conn.ExecContext(ctx, createPriorityIndexQuery)
	if err != nil {
		logAction.SetError("Failed to create priority index on UserSubscriptions", err.Error(), map[string]any{
			"error": err.Error(),
			"query": createPriorityIndexQuery,
		})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v6.0 to v7.0 completed successfully")
	return Err
}
