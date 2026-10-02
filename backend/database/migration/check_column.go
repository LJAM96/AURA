package migration

import (
	"aura/database"
	"aura/logging"
	"context"
	"database/sql"
)

func checkColumnExists(ctx context.Context, tableName string, columnName string) (exists bool, Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	// Get DB connection
	conn, _, getDBConnErr := database.GetDBConnection(ctx)
	if getDBConnErr.Message != "" {
		return false, getDBConnErr
	}

	return tableColumnExists(ctx, conn, tableName, columnName)
}

// tableColumnExists is checkColumnExists for callers that already hold a
// connection, such as migrations that need to inspect a table before deciding
// how to rebuild it.
func tableColumnExists(ctx context.Context, conn *sql.DB, tableName string, columnName string) (exists bool, Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	// Check if the column already exists to avoid duplicate column error
	checkColumnQuery := `PRAGMA table_info(` + tableName + `);`
	rows, err := conn.QueryContext(ctx, checkColumnQuery)
	if err != nil {
		Err = logging.LogErrorInfo{
			Message: "Failed to query " + tableName + " table info",
			Detail:  map[string]any{"error": err.Error()},
		}
		return false, Err
	}
	defer rows.Close()

	exists = false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int
		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			Err = logging.LogErrorInfo{
				Message: "Failed to scan " + tableName + " table info",
				Detail:  map[string]any{"error": err.Error()},
			}
			return false, Err
		}
		if name == columnName {
			exists = true
			break
		}
	}
	if err = rows.Err(); err != nil {
		Err = logging.LogErrorInfo{
			Message: "Failed to read " + tableName + " table info",
			Detail:  map[string]any{"error": err.Error()},
		}
		return false, Err
	}

	return exists, Err
}

// tableExists reports whether a table is already present in the database.
func tableExists(ctx context.Context, conn *sql.DB, tableName string) (exists bool, Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	row := conn.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?;`, tableName)

	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, Err
		}
		Err = logging.LogErrorInfo{
			Message: "Failed to check if " + tableName + " table exists",
			Detail:  map[string]any{"error": err.Error()},
		}
		return false, Err
	}

	return true, Err
}
