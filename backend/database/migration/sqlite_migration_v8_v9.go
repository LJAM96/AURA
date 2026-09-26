package migration

import (
	"aura/database"
	"aura/logging"
	"context"
)

// migrate_8_to_9 is a catch-up migration.
//
// Upstream used database versions 6 and 7 for the "edition" column and the
// DownloadQueueJobs/DownloadHistory tables, while this fork had already used
// those same versions for the UserSubscriptions table and its priority column.
// A database created by this fork is therefore already stamped as version 8
// while missing the upstream schema, and would never pick it up.
//
// Every step below is guarded so this migration is safe to run against a
// database coming from either lineage, and safe to re-run.
func migrate_8_to_9(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Migrating Database from v8 to v9", logging.LevelInfo)
	defer logAction.Complete()
	logging.LOGGER.Info().Timestamp().Int("From Version", 8).Int("To Version", 9).Msg("Starting database migration")

	Err = logging.LogErrorInfo{}

	// Create a backup of the current database
	backupErr := database.Backup(ctx, 8, 9)
	if backupErr.Message != "" {
		return backupErr
	}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return getDBConnErr
	}

	// --- Upstream's edition schema, if this database predates it ---
	editionColumnExists, checkColumnErr := checkColumnExists(ctx, "MediaItems", "edition")
	if checkColumnErr.Message != "" {
		return checkColumnErr
	}

	if !editionColumnExists {
		// This fork added SavedItems.priority in v7 -> v8, so it has to survive
		// the rebuild that introduces the edition column
		editionErr := applyEditionSchema(ctx, conn, logAction, true)
		if editionErr.Message != "" {
			return editionErr
		}
	}

	// --- Upstream's download queue and history tables, if this database predates them ---
	queueJobsExists, tableCheckErr := tableExists(ctx, conn, "DownloadQueueJobs")
	if tableCheckErr.Message != "" {
		return tableCheckErr
	}

	if !queueJobsExists {
		queueErr := createDownloadQueueAndHistoryTables(ctx, conn, logAction)
		if queueErr.Message != "" {
			return queueErr
		}
	}

	// --- This fork's UserSubscriptions table, if this database predates it ---
	subscriptionsExist, tableCheckErr := tableExists(ctx, conn, "UserSubscriptions")
	if tableCheckErr.Message != "" {
		return tableCheckErr
	}

	if !subscriptionsExist {
		createQuery := `
CREATE TABLE IF NOT EXISTS UserSubscriptions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT NOT NULL,
    creator_id      TEXT NOT NULL DEFAULT '',
    image_types     TEXT NOT NULL DEFAULT '{"poster":false,"backdrop":false,"season_poster":false,"special_season_poster":false,"titlecard":false}',
    priority        INTEGER NOT NULL DEFAULT 1,
    media_scope     TEXT NOT NULL DEFAULT 'all' CHECK (media_scope IN ('all','movies','shows','collections')),
    library_section TEXT,
    enabled         INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    date_created    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_updated    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(username)
);
		`
		if _, err := conn.ExecContext(ctx, createQuery); err != nil {
			logAction.SetError("Failed to create UserSubscriptions table", err.Error(), map[string]any{
				"error": err.Error(),
				"query": createQuery,
			})
			return *logAction.Error
		}
	}

	// --- Priority columns, if they are missing ---
	// A database from this fork's original v6 -> v7 has UserSubscriptions but no
	// priority column, and a database from upstream's v6 -> v7 has neither table
	// nor column, so both are checked independently.
	subscriptionPriorityExists, checkColumnErr := checkColumnExists(ctx, "UserSubscriptions", "priority")
	if checkColumnErr.Message != "" {
		return checkColumnErr
	}

	if !subscriptionPriorityExists {
		addPriorityQuery := `ALTER TABLE UserSubscriptions ADD COLUMN priority INTEGER NOT NULL DEFAULT 1;`
		if _, err := conn.ExecContext(ctx, addPriorityQuery); err != nil {
			logAction.SetError("Failed to add priority column to UserSubscriptions table", err.Error(), map[string]any{
				"error": err.Error(),
				"query": addPriorityQuery,
			})
			return *logAction.Error
		}

		createPriorityIndexQuery := `CREATE INDEX IF NOT EXISTS idx_usersubscriptions_priority ON UserSubscriptions(priority);`
		if _, err := conn.ExecContext(ctx, createPriorityIndexQuery); err != nil {
			logAction.SetError("Failed to create priority index on UserSubscriptions", err.Error(), map[string]any{
				"error": err.Error(),
				"query": createPriorityIndexQuery,
			})
			return *logAction.Error
		}
	}

	savedItemsPriorityExists, checkColumnErr := checkColumnExists(ctx, "SavedItems", "priority")
	if checkColumnErr.Message != "" {
		return checkColumnErr
	}

	if !savedItemsPriorityExists {
		addPriorityQuery := `ALTER TABLE SavedItems ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;`
		if _, err := conn.ExecContext(ctx, addPriorityQuery); err != nil {
			logAction.SetError("Failed to add priority column to SavedItems table", err.Error(), map[string]any{
				"error": err.Error(),
				"query": addPriorityQuery,
			})
			return *logAction.Error
		}
	}

	logging.LOGGER.Info().Timestamp().Msg("Database migration v8.0 to v9.0 completed successfully")
	return Err
}
