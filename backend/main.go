// @title Aura API
// @version 1.0
// @BasePath /
// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name aura_session
// @description Browser session cookie, set by POST /api/login or the OIDC callback. Not usable directly via Swagger "Authorize" - log in through the app UI in the same browser tab.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key
// @description API key for programmatic/integration access. Generate one under Settings > Authentication, then send it as this header on every request. This is the intended auth method for scripts and integrations - the session cookie above is for browser use only.
// @securityDefinitions.basic BasicAuth
// @description Used only by /api/sonarr/webhook, since Sonarr/Radarr's built-in Webhook connection type has no custom-header support. Any username works; the password must be the API key.
package main

import (
	"aura/config"
	"aura/logging"
	"aura/notification"
	"aura/routing"
	"context"
	"os"
	"strings"
	"sync/atomic"
)

var (
	APP_NAME    = "aura"
	APP_VERSION = "dev"
	AUTHOR      = "xmoosex"
	LICENSE     = "MIT"
	APP_PORT    = 8888
)

var activeHandler atomic.Value

func init() {
	if strings.HasSuffix(APP_VERSION, "dev") {
		logging.SetDevMode(true)
	}
}

func main() {
	// Serve immediately with onboarding/public routes first.
	config.AppFullyLoaded = false
	config.AppVersion = APP_VERSION
	config.AppLoadingStep = "Initializing Application"
	activeHandler.Store(routing.NewRouter())

	// Start API now (non-blocking for init pipeline).
	go startAPI()

	// Run startup pipeline in background.
	go func() {
		bootStrapSuccess := runBootstrap()

		// Keep callback for onboarding finalization path.
		routing.OnboardingComplete = func() {
			preflightSuccess := runPreFlight()
			if !preflightSuccess {
				logging.LOGGER.Error().Timestamp().Msg("Preflight failed during OnboardingComplete, not swapping routers")
				return
			}
			// Init the database synchronously (fast, required by full routes),
			// then serve full routes immediately while the slow library scan
			// refreshes in the background.
			if !runWarmupCritical(context.Background()) {
				logging.LOGGER.Fatal().Timestamp().Msg("Database init failed during OnboardingComplete. Exiting application.")
				return
			}
			config.AppFullyLoaded = false
			config.AppLoadingStep = "Serving API - background library refresh in progress"
			activeHandler.Store(routing.NewRouter())
			go runWarmupBackground()
			logging.LOGGER.Info().Timestamp().Msg("Onboarding complete. Main routes active (library refresh in background).")
		}

		if bootStrapSuccess {
			preflightSuccess := runPreFlight()
			if !preflightSuccess {
				config.Valid = false
				activeHandler.Store(routing.NewRouter()) // stays onboarding
				return
			}

			// Init the database synchronously (fast, required by full routes),
			// then serve full routes immediately while the slow library scan
			// (74k+ items can take 15+ minutes) refreshes in the background.
			// Search/detail/image endpoints work on whatever is cached so far,
			// with live-fetch fallbacks for items not yet scanned.
			if !runWarmupCritical(context.Background()) {
				logging.LOGGER.Fatal().Timestamp().Msg("Database init failed. Exiting application.")
				os.Exit(1)
			}

			config.AppFullyLoaded = false
			config.AppLoadingStep = "Serving API - background library refresh in progress"
			// Send App Start Notification
			// Send notification (only if not dev & notifications enabled)
			if !strings.Contains(APP_VERSION, "dev") &&
				config.Current.Notifications.Enabled {
				notification.SendAppStartNotification(APP_PORT, APP_NAME, APP_VERSION)
			} else {
				logging.LOGGER.Warn().Timestamp().Bool("notifications_enabled", config.Current.Notifications.Enabled).Bool("dev_version", strings.Contains(APP_VERSION, "dev")).Msg("App start notification not sent")
			}
			activeHandler.Store(routing.NewRouter()) // swap to full routes now
			go runWarmupBackground()
			logging.LOGGER.Info().Timestamp().Msg("Main routes active (library refresh in background).")
			return
		}

		// Config not loaded/valid: onboarding mode remains active.
		activeHandler.Store(routing.NewRouter())
	}()

	// Keep process alive while startAPI runs.
	select {}
}
