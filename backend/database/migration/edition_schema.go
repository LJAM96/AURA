package migration

import (
	"aura/logging"
	"context"
	"database/sql"
)

// applyEditionSchema rebuilds MediaItems, SavedItems and IgnoredItems with an
// "edition" column folded into their uniqueness constraints, so that multiple
// editions of the same TMDB item (e.g. Theatrical vs Director's Cut) no longer
// collide with each other. Existing rows are given edition = ”.
//
// preserveSavedItemsPriority carries an existing SavedItems.priority column
// across the rebuild. It is required by the v8 -> v9 catch-up migration, which
// can run against databases that gained a priority column before this edition
// schema existed. The v5 -> v6 migration never needs it, since a v5 database
// cannot have a priority column.
func applyEditionSchema(ctx context.Context, conn *sql.DB, logAction *logging.LogAction, preserveSavedItemsPriority bool) (Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	priorityExists := false
	if preserveSavedItemsPriority {
		priorityExists, Err = tableColumnExists(ctx, conn, "SavedItems", "priority")
		if Err.Message != "" {
			return Err
		}
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logAction.SetError("Failed to begin transaction for adding edition column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	// --- MediaItems ---
	if _, err = tx.ExecContext(ctx, `ALTER TABLE MediaItems RENAME TO MediaItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to rename MediaItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `
		CREATE TABLE MediaItems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tmdb_id TEXT NOT NULL,
			library_title TEXT NOT NULL,
			edition TEXT NOT NULL DEFAULT '',
			rating_key TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('movie','show')),
			title TEXT NOT NULL,
			year INTEGER NOT NULL,
			on_server INTEGER NOT NULL DEFAULT 0 CHECK (on_server IN (0,1)),
			UNIQUE (tmdb_id, library_title, edition)
		);
	`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to create new MediaItems table with edition column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO MediaItems (id, tmdb_id, library_title, edition, rating_key, type, title, year, on_server)
		SELECT id, tmdb_id, library_title, '', rating_key, type, title, year, on_server
		FROM MediaItems_old;
	`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to copy data into new MediaItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `DROP TABLE MediaItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to drop old MediaItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	// --- SavedItems ---
	if _, err = tx.ExecContext(ctx, `ALTER TABLE SavedItems RENAME TO SavedItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to rename SavedItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	priorityColumn := ""
	prioritySelect := ""
	if priorityExists {
		priorityColumn = "\n\t\t\t\tpriority INTEGER NOT NULL DEFAULT 0,"
		prioritySelect = "\n\t\t\t\tautodownload, priority, auto_add_new_collection_items, last_downloaded"
		prioritySelect += "\n\t\t\tSELECT\n\t\t\t\tautodownload, priority, auto_add_new_collection_items, last_downloaded"
	} else {
		prioritySelect = "\n\t\t\t\tautodownload, auto_add_new_collection_items, last_downloaded"
		prioritySelect += "\n\t\t\tSELECT\n\t\t\t\tautodownload, auto_add_new_collection_items, last_downloaded"
	}

	createSavedItemsQuery := `
		CREATE TABLE SavedItems (
			tmdb_id TEXT NOT NULL,
			library_title TEXT NOT NULL,
			edition TEXT NOT NULL DEFAULT '',
			poster_set_id INTEGER NOT NULL,

			poster_selected INTEGER NOT NULL DEFAULT 0 CHECK (poster_selected IN (0,1)),
			backdrop_selected INTEGER NOT NULL DEFAULT 0 CHECK (backdrop_selected IN (0,1)),
			season_poster_selected INTEGER NOT NULL DEFAULT 0 CHECK (season_poster_selected IN (0,1)),
			special_season_poster_selected INTEGER NOT NULL DEFAULT 0 CHECK (special_season_poster_selected IN (0,1)),
			titlecard_selected INTEGER NOT NULL DEFAULT 0 CHECK (titlecard_selected IN (0,1)),

			autodownload INTEGER NOT NULL DEFAULT 0 CHECK (autodownload IN (0,1)),
			auto_add_new_collection_items INTEGER NOT NULL DEFAULT 0 CHECK (auto_add_new_collection_items IN (0,1)),` + priorityColumn + `
			last_downloaded DATETIME NOT NULL,

			PRIMARY KEY (tmdb_id, library_title, edition, poster_set_id),

			FOREIGN KEY (poster_set_id) REFERENCES PosterSets(id)
				ON DELETE CASCADE
				ON UPDATE CASCADE,

			FOREIGN KEY (tmdb_id, library_title, edition) REFERENCES MediaItems(tmdb_id, library_title, edition)
				ON DELETE CASCADE
				ON UPDATE CASCADE
		) WITHOUT ROWID;
	`

	if _, err = tx.ExecContext(ctx, createSavedItemsQuery); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to create new SavedItems table with edition column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	copySavedItemsQuery := `
		INSERT INTO SavedItems (
			tmdb_id, library_title, edition, poster_set_id,
			poster_selected, backdrop_selected, season_poster_selected, special_season_poster_selected, titlecard_selected,` + prioritySelect + `
		)
		FROM SavedItems_old;
	`

	if _, err = tx.ExecContext(ctx, copySavedItemsQuery); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to copy data into new SavedItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `DROP TABLE SavedItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to drop old SavedItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	// --- IgnoredItems ---
	if _, err = tx.ExecContext(ctx, `ALTER TABLE IgnoredItems RENAME TO IgnoredItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to rename IgnoredItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `
		CREATE TABLE IgnoredItems (
			tmdb_id TEXT NOT NULL,
			library_title TEXT NOT NULL,
			edition TEXT NOT NULL DEFAULT '',
			mode TEXT NOT NULL CHECK (mode IN ('always','until-set-available','until-new-set-available')),
			current_sets TEXT NOT NULL DEFAULT '[]',
			PRIMARY KEY (tmdb_id, library_title, edition)
		) WITHOUT ROWID;
	`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to create new IgnoredItems table with edition column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO IgnoredItems (tmdb_id, library_title, edition, mode, current_sets)
		SELECT tmdb_id, library_title, '', mode, current_sets
		FROM IgnoredItems_old;
	`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to copy data into new IgnoredItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}
	if _, err = tx.ExecContext(ctx, `DROP TABLE IgnoredItems_old;`); err != nil {
		tx.Rollback()
		logAction.SetError("Failed to drop old IgnoredItems table", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	if err = tx.Commit(); err != nil {
		logAction.SetError("Failed to commit transaction for adding edition column", "", map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	return Err
}
