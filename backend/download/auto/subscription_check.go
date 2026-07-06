package autodownload

import (
	"aura/cache"
	"aura/database"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils"
	"context"
	"fmt"
	"time"
)

// CheckSubscriptions checks all enabled subscriptions for new sets from creators
// and auto-downloads matching items to the user's media server.
func CheckSubscriptions(ctx context.Context) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Checking Subscriptions for New Sets", logging.LevelInfo)
	defer logAction.Complete()

	// Get all enabled subscriptions
	subs, Err := database.GetEnabledSubscriptions(ctx)
	if Err.Message != "" {
		return Err
	}

	if len(subs) == 0 {
		logging.LOGGER.Info().Timestamp().Msg("No active subscriptions to check")
		return logging.LogErrorInfo{}
	}

	logging.LOGGER.Info().Timestamp().Int("count", len(subs)).Msg("Checking subscriptions for new sets")

	newItemsCount := 0
	errorCount := 0
	skippedCount := 0

	for _, sub := range subs {
		subCtx, subLog := logging.CreateLoggingContext(context.Background(), fmt.Sprintf("Subscription Check: %s", sub.Username))
		subAction := subLog.AddAction(fmt.Sprintf("Checking subscription for creator '%s'", sub.Username), logging.LevelInfo)
		subCtx = logging.WithCurrentAction(subCtx, subAction)
		defer subAction.Complete()

		// Fetch the creator's latest sets from MediUX
		creatorSets, Err := mediux.GetAllUserSets(subCtx, sub.Username)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("username", sub.Username).Str("error", Err.Message).
				Msg("Failed to fetch creator sets for subscription check")
			errorCount++
			subLog.Log()
			continue
		}

		// Process each type of set based on media_scope
		processedItems := map[string]bool{} // track unique items

		// Process show sets
		if sub.MediaScope == "all" || sub.MediaScope == "shows" {
			for _, set := range creatorSets.ShowSets {
				for _, itemID := range set.ItemIDs {
					if processedItems[itemID] {
						continue
					}
					processedItems[itemID] = true
					if err := processSubscriptionSet(subCtx, sub, set.PosterSet, itemID, "show"); err.Message != "" {
						errorCount++
					} else {
						newItemsCount++
					}
				}
			}
		}

		// Process movie sets
		if sub.MediaScope == "all" || sub.MediaScope == "movies" {
			for _, set := range creatorSets.MovieSets {
				for _, itemID := range set.ItemIDs {
					if processedItems[itemID] {
						continue
					}
					processedItems[itemID] = true
					if err := processSubscriptionSet(subCtx, sub, set.PosterSet, itemID, "movie"); err.Message != "" {
						errorCount++
					} else {
						newItemsCount++
					}
				}
			}
		}

		// Process collection sets
		if sub.MediaScope == "all" || sub.MediaScope == "collections" {
			for _, set := range creatorSets.CollectionSets {
				for _, itemID := range set.ItemIDs {
					if processedItems[itemID] {
						continue
					}
					processedItems[itemID] = true
					if err := processSubscriptionSet(subCtx, sub, set.PosterSet, itemID, "movie"); err.Message != "" {
						errorCount++
					} else {
						newItemsCount++
					}
				}
			}
		}

		logging.LOGGER.Info().Timestamp().Str("username", sub.Username).
			Int("items_processed", len(processedItems)).
			Msg("Completed subscription check for creator")
		subLog.Log()
	}

	logging.LOGGER.Info().Timestamp().
		Int("new_items", newItemsCount).
		Int("errors", errorCount).
		Int("skipped", skippedCount).
		Msg("Completed subscription check for all creators")

	return logging.LogErrorInfo{}
}

// processSubscriptionSet checks if a specific set's item exists in the user's library
// and if it's already saved. If not saved, it creates a new SavedItem entry.
func processSubscriptionSet(ctx context.Context, sub models.UserSubscription, set models.PosterSet, itemTMDBID string, itemType string) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Processing Subscription Set for item %s", itemTMDBID), logging.LevelDebug)
	defer logAction.Complete()

	// Check if this item already exists in the database
	ignored, _, existingSets, _ := database.CheckIfMediaItemExists(ctx, itemTMDBID, "")
	if ignored {
		logAction.AppendResult("status", "ignored")
		return logging.LogErrorInfo{}
	}

	// Check if we already have this set saved
	for _, existingSet := range existingSets {
		if existingSet.ID == set.ID {
			logAction.AppendResult("status", "already_saved")
			return logging.LogErrorInfo{}
		}
	}

	// Find the item in the user's library cache
	var foundItem *models.MediaItem
	var libraryTitle string

	// Search through all library sections
	sections := cache.LibraryStore.GetAllSectionsSortedByTitle()
	for _, section := range sections {
		item, found := cache.LibraryStore.GetMediaItemFromSectionByTMDBID(section.Title, itemTMDBID)
		if found && item != nil && item.RatingKey != "" {
			foundItem = item
			libraryTitle = section.Title
			break
		}
	}

	if foundItem == nil {
		logAction.AppendResult("status", "not_in_library")
		return logging.LogErrorInfo{}
	}

	// Check if library section matches subscription filter
	if sub.LibrarySection != nil && *sub.LibrarySection != "" && *sub.LibrarySection != libraryTitle {
		logAction.AppendResult("status", "library_mismatch")
		return logging.LogErrorInfo{}
	}

	// Build the DBSavedItem with subscription's image types
	dbItem := models.DBSavedItem{
		MediaItem: *foundItem,
		PosterSets: []models.DBPosterSetDetail{
			{
				PosterSet: set,
				SelectedTypes: models.SelectedTypes{
					Poster:              sub.ImageTypes.Poster,
					Backdrop:            sub.ImageTypes.Backdrop,
					SeasonPoster:        sub.ImageTypes.SeasonPoster,
					SpecialSeasonPoster: sub.ImageTypes.SpecialSeasonPoster,
					Titlecard:           sub.ImageTypes.Titlecard,
				},
				AutoDownload:              true,
				AutoAddNewCollectionItems: false,
				Priority:                  sub.Priority,
				LastDownloaded:            time.Time{}, // Zero time - will trigger download
			},
		},
	}

	// Upsert to database
	Err := database.UpsertSavedItem(ctx, dbItem)
	if Err.Message != "" {
		logging.LOGGER.Error().Timestamp().Str("item", utils.MediaItemInfo(*foundItem)).Str("error", Err.Message).
			Msg("Failed to upsert subscription set to database")
		return Err
	}

	logging.LOGGER.Info().Timestamp().Str("item", utils.MediaItemInfo(*foundItem)).
		Str("set_id", set.ID).Str("set_title", set.Title).
		Str("username", sub.Username).
		Msg("Auto-added new set from subscription")

	logAction.AppendResult("status", "added")
	logAction.AppendResult("set_id", set.ID)
	logAction.AppendResult("item", utils.MediaItemInfo(*foundItem))

	return logging.LogErrorInfo{}
}

// CheckSubscriptionForNewSet is called when a WebSocket update arrives for a set.
// It checks if any subscription matches the creator and auto-adds the set if needed.
func CheckSubscriptionForNewSet(ctx context.Context, setID int, creatorUsername string, setTMDBID string, set models.PosterSet) {
	// Check if any subscription exists for this creator
	sub, Err := database.GetSubscriptionByUsername(ctx, creatorUsername)
	if Err.Message != "" || sub == nil || !sub.Enabled {
		return // No subscription for this creator
	}

	logging.LOGGER.Info().Timestamp().Str("username", creatorUsername).Int("set_id", setID).
		Str("tmdb_id", setTMDBID).
		Msg("Subscription match found for updated set")

	// Process the set
	_ = processSubscriptionSet(ctx, *sub, set, setTMDBID, set.Type)
}
