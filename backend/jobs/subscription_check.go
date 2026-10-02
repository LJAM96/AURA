package jobs

import (
	"aura/config"
	autodownload "aura/download/auto"
	"aura/logging"
	"context"
)

func runSubscriptionCheckJobBody(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in SubscriptionCheckJob")
		}
	}()
	ctx, ld := logging.CreateLoggingContext(ctx, "Cron Job")
	action := ld.AddAction("Subscription Check", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	Err := autodownload.CheckSubscriptions(ctx)
	if Err.Message != "" {
		logging.LOGGER.Error().Timestamp().Str("error", Err.Message).
			Msg("Error running Subscription Check Job")
	}
	ld.Log()
}

func StartSubscriptionCheckJob() error {
	mu.Lock()
	defer mu.Unlock()

	if c == nil {
		logging.LOGGER.Error().Timestamp().Msg("Cron Jobs Scheduler is not initialized")
		return nil
	}

	removeScheduledJob(JobKeySubscriptionCheck)

	enabled, spec := config.Current.Jobs.SubscriptionCheck.Resolve(config.JobDefaults.SubscriptionCheck)
	if !enabled {
		logging.LOGGER.Info().Timestamp().Msg("Subscription Check Job Stopped")
		return nil
	}

	if err := addScheduledJob(JobKeySubscriptionCheck, spec, func() {
		runSubscriptionCheckJobBody(context.Background())
	}); err != nil {
		return err
	}

	logging.LOGGER.Info().Timestamp().
		Str("cron", spec).
		Msg("Subscription Check Job Started")
	return nil
}

func RunSubscriptionCheckJobNow() {
	go func() {
		recordManualRun(JobKeySubscriptionCheck)
		runSubscriptionCheckJobBody(context.Background())
	}()
}
