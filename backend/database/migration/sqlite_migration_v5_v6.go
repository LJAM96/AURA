package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

func migrate_5_to_6(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v5 to v6", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 5).Int("To Version", 6).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 5, 6)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// Create UserSubscriptions table
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS UserSubscriptions (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			username        TEXT NOT NULL,
			creator_id      TEXT NOT NULL DEFAULT '',
			image_types     TEXT NOT NULL DEFAULT '{"poster":false,"backdrop":false,"season_poster":false,"special_season_poster":false,"titlecard":false}',
			media_scope     TEXT NOT NULL DEFAULT 'all' CHECK (media_scope IN ('all','movies','shows','collections')),
			library_section TEXT,
			enabled         INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
			date_created    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			date_updated    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(username)
		);
	`
	_, err := conn.ExecContext(ctx, createTableQuery)
	if err != nil {
		logAction.SetError("Failed to create UserSubscriptions table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": createTableQuery,
		})
		return *logAction.Error
	}

	// Add index on username for fast lookups
	createIndexQuery := `CREATE INDEX IF NOT EXISTS idx_usersubscriptions_username ON UserSubscriptions(username);`
	_, err = conn.ExecContext(ctx, createIndexQuery)
	if err != nil {
		logAction.SetError("Failed to create index on UserSubscriptions", err.Error(), map[string]any{
			"error": err.Error(),
			"query": createIndexQuery,
		})
		return *logAction.Error
	}

	// Add index on enabled for subscription checks
	createEnabledIndexQuery := `CREATE INDEX IF NOT EXISTS idx_usersubscriptions_enabled ON UserSubscriptions(enabled);`
	_, err = conn.ExecContext(ctx, createEnabledIndexQuery)
	if err != nil {
		logAction.SetError("Failed to create enabled index on UserSubscriptions", err.Error(), map[string]any{
			"error": err.Error(),
			"query": createEnabledIndexQuery,
		})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v5.0 to v6.0 completed successfully")
	return Err
}
