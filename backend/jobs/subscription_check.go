package jobs

import (
	"aura/config"
	autodownload "aura/download/auto"
	"aura/logging"
	"context"
	"runtime/debug"
)

func StartSubscriptionCheckJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	if subscriptionCheckJobID != 0 {
		c.Remove(subscriptionCheckJobID)
		subscriptionCheckJobID = 0
	}

	enabled := config.Current.AutoDownload.Enabled
	if !enabled {
		logging.LOGGER.Info().Timestamp().Msg("Subscription Check Job Stopped (AutoDownload disabled)")
		return nil
	}

	// Run every 6 hours by default
	spec := "0 */6 * * *"

	var err error

	subscriptionCheckJobID, err = c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().
					Timestamp().
					Interface("recover", r).
					Str("stack", string(debug.Stack())).
					Msg("PANIC: in scheduled Subscription Check Job")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Cron Job")
		action := ld.AddAction("Subscription Check", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := autodownload.CheckSubscriptions(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(subscriptionCheckJobID).Next.String()).
				Msg("Error running Subscription Check Job")
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(subscriptionCheckJobID).Next.String()).
				Msg("Subscription Check Job Completed")
		}
	})
	if err != nil {
		return err
	}
	jobSpecs[subscriptionCheckJobID] = spec

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Msg("Subscription Check Job Started")
	return nil
}

func RunSubscriptionCheckJobNow() {
	mu.Lock()
	defer mu.Unlock()

	if config.Current.AutoDownload.Enabled == false {
		logging.LOGGER.Warn().Timestamp().Msg("AutoDownload is disabled, cannot run subscription check job")
		return
	}

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return
	}

	if subscriptionCheckJobID == 0 {
		logging.LOGGER.Error().Timestamp().Msg("Subscription Check Job is not scheduled")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("Panic in Subscription Check Job")
			}
		}()
		ctx, ld := logging.CreateLoggingContext(context.Background(), "Manual Job Run")
		action := ld.AddAction("Subscription Check", logging.LevelInfo)
		ctx = logging.WithCurrentAction(ctx, action)
		Err := autodownload.CheckSubscriptions(ctx)
		if Err.Message != "" {
			logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
				Str("next_run", c.Entry(subscriptionCheckJobID).Next.String()).
				Msg("Error running Manual Subscription Check Job")
			return
		} else {
			logging.LOGGER.Info().Timestamp().
				Str("next_run", c.Entry(subscriptionCheckJobID).Next.String()).
				Msg("Manual Subscription Check Job Completed")
		}
	}()
}
