package database

import (
	"aura/logging"
	"aura/models"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (s *SQliteDB) CreateSubscription(ctx context.Context, sub models.UserSubscription) (int, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating User Subscription", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return 0, *logAction.Error
	}

	imageTypesJSON, err := json.Marshal(sub.ImageTypes)
	if err != nil {
		logAction.SetError("DB: failed to marshal image types", err.Error(), map[string]any{"error": err.Error()})
		return 0, *logAction.Error
	}

	q := `
INSERT INTO UserSubscriptions (username, creator_id, image_types, priority, media_scope, library_section, enabled, date_created, date_updated)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(username) DO UPDATE SET
  creator_id      = excluded.creator_id,
  image_types     = excluded.image_types,
  priority        = excluded.priority,
  media_scope     = excluded.media_scope,
  library_section = excluded.library_section,
  enabled         = excluded.enabled,
  date_updated    = excluded.date_updated
RETURNING id;
`
	now := time.Now().UTC()
	var id int
	err = s.conn.QueryRowContext(ctx, q,
		sub.Username,
		sub.CreatorID,
		string(imageTypesJSON),
		sub.Priority,
		sub.MediaScope,
		sub.LibrarySection,
		boolToInt(sub.Enabled),
		now,
		now,
	).Scan(&id)
	if err != nil {
		logAction.SetError("DB: UPSERT UserSubscriptions failed", err.Error(), map[string]any{"error": err.Error()})
		return 0, *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Str("username", sub.Username).Msg("Created/updated UserSubscription")
	return id, logging.LogErrorInfo{}
}

func (s *SQliteDB) UpdateSubscription(ctx context.Context, sub models.UserSubscription) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Updating User Subscription", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return *logAction.Error
	}

	imageTypesJSON, err := json.Marshal(sub.ImageTypes)
	if err != nil {
		logAction.SetError("DB: failed to marshal image types", err.Error(), map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	q := `
UPDATE UserSubscriptions
SET creator_id = ?, image_types = ?, priority = ?, media_scope = ?, library_section = ?, enabled = ?, date_updated = ?
WHERE id = ?;
`
	now := time.Now().UTC()
	_, err = s.conn.ExecContext(ctx, q,
		sub.CreatorID,
		string(imageTypesJSON),
		sub.Priority,
		sub.MediaScope,
		sub.LibrarySection,
		boolToInt(sub.Enabled),
		now,
		sub.ID,
	)
	if err != nil {
		logAction.SetError("DB: UPDATE UserSubscriptions failed", err.Error(), map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Int("id", sub.ID).Msg("Updated UserSubscription")
	return logging.LogErrorInfo{}
}

func (s *SQliteDB) DeleteSubscription(ctx context.Context, id int) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Deleting User Subscription", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return *logAction.Error
	}

	q := `DELETE FROM UserSubscriptions WHERE id = ?;`
	_, err := s.conn.ExecContext(ctx, q, id)
	if err != nil {
		logAction.SetError("DB: DELETE UserSubscriptions failed", err.Error(), map[string]any{"error": err.Error()})
		return *logAction.Error
	}

	logging.LOGGER.Info().Timestamp().Int("id", id).Msg("Deleted UserSubscription")
	return logging.LogErrorInfo{}
}

func (s *SQliteDB) GetAllSubscriptions(ctx context.Context) ([]models.UserSubscription, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting All User Subscriptions", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return nil, *logAction.Error
	}

	q := `
SELECT id, username, creator_id, image_types, priority, media_scope, library_section, enabled, date_created, date_updated
FROM UserSubscriptions
ORDER BY priority ASC, date_created DESC;
`
	rows, err := s.conn.QueryContext(ctx, q)
	if err != nil {
		logAction.SetError("DB: SELECT UserSubscriptions failed", err.Error(), map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}
	defer rows.Close()

	var subs []models.UserSubscription
	for rows.Next() {
		var sub models.UserSubscription
		var imageTypesJSON string
		var librarySection *string
		var enabledInt int
		err := rows.Scan(
			&sub.ID,
			&sub.Username,
			&sub.CreatorID,
			&imageTypesJSON,
			&sub.Priority,
			&sub.MediaScope,
			&librarySection,
			&enabledInt,
			&sub.DateCreated,
			&sub.DateUpdated,
		)
		if err != nil {
			logAction.SetError("DB: scan UserSubscription failed", err.Error(), map[string]any{"error": err.Error()})
			return nil, *logAction.Error
		}
		sub.LibrarySection = librarySection
		sub.Enabled = enabledInt == 1
		if err := json.Unmarshal([]byte(imageTypesJSON), &sub.ImageTypes); err != nil {
			logAction.SetError("DB: unmarshal image types failed", err.Error(), map[string]any{"error": err.Error()})
			return nil, *logAction.Error
		}
		subs = append(subs, sub)
	}

	return subs, logging.LogErrorInfo{}
}

func (s *SQliteDB) GetSubscriptionByUsername(ctx context.Context, username string) (*models.UserSubscription, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting User Subscription by Username", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return nil, *logAction.Error
	}

	q := `
SELECT id, username, creator_id, image_types, priority, media_scope, library_section, enabled, date_created, date_updated
FROM UserSubscriptions
WHERE username = ?;
`
	var sub models.UserSubscription
	var imageTypesJSON string
	var librarySection *string
	var enabledInt int
	err := s.conn.QueryRowContext(ctx, q, username).Scan(
		&sub.ID,
		&sub.Username,
		&sub.CreatorID,
		&imageTypesJSON,
		&sub.Priority,
		&sub.MediaScope,
		&librarySection,
		&enabledInt,
		&sub.DateCreated,
		&sub.DateUpdated,
	)
	if err != nil {
		logAction.SetError(fmt.Sprintf("DB: no subscription found for username '%s'", username), err.Error(), map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}
	sub.LibrarySection = librarySection
	sub.Enabled = enabledInt == 1
	if err := json.Unmarshal([]byte(imageTypesJSON), &sub.ImageTypes); err != nil {
		logAction.SetError("DB: unmarshal image types failed", err.Error(), map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}

	return &sub, logging.LogErrorInfo{}
}

func (s *SQliteDB) GetEnabledSubscriptions(ctx context.Context) ([]models.UserSubscription, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting Enabled User Subscriptions", logging.LevelInfo)
	defer logAction.Complete()

	if s == nil || s.conn == nil {
		logAction.SetError("DB: connection is nil", "", map[string]any{})
		return nil, *logAction.Error
	}

	q := `
SELECT id, username, creator_id, image_types, priority, media_scope, library_section, enabled, date_created, date_updated
FROM UserSubscriptions
WHERE enabled = 1
ORDER BY priority ASC, date_created DESC;
`
	rows, err := s.conn.QueryContext(ctx, q)
	if err != nil {
		logAction.SetError("DB: SELECT enabled UserSubscriptions failed", err.Error(), map[string]any{"error": err.Error()})
		return nil, *logAction.Error
	}
	defer rows.Close()

	var subs []models.UserSubscription
	for rows.Next() {
		var sub models.UserSubscription
		var imageTypesJSON string
		var librarySection *string
		var enabledInt int
		err := rows.Scan(
			&sub.ID,
			&sub.Username,
			&sub.CreatorID,
			&imageTypesJSON,
			&sub.Priority,
			&sub.MediaScope,
			&librarySection,
			&enabledInt,
			&sub.DateCreated,
			&sub.DateUpdated,
		)
		if err != nil {
			logAction.SetError("DB: scan UserSubscription failed", err.Error(), map[string]any{"error": err.Error()})
			return nil, *logAction.Error
		}
		sub.LibrarySection = librarySection
		sub.Enabled = enabledInt == 1
		if err := json.Unmarshal([]byte(imageTypesJSON), &sub.ImageTypes); err != nil {
			logAction.SetError("DB: unmarshal image types failed", err.Error(), map[string]any{"error": err.Error()})
			return nil, *logAction.Error
		}
		subs = append(subs, sub)
	}

	return subs, logging.LogErrorInfo{}
}
