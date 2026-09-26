package routes_config

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

type reloadConfigResponse struct {
	Status AppConfigStatus `json:"status"`
}

// ConfigReload godoc
// @Summary      Reload Config
// @Description  Reload the configuration file and return the current config status
// @Tags         Config
// @Produce      json
// @Security     SessionCookie
// @Security     ApiKeyAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success      200  {object}  httpx.JSONResponse{data=reloadConfigResponse}
// @Failure      500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/config [patch]
func ReloadAppConfig(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Reload Config", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var response reloadConfigResponse

	// Reload the config file
	config.LoadYAML(ctx)

	// Print the config details (sanitized)
	config.Current.PrintDetails()

	// Sanitize the config before sending it back
	sanitizedConfig := config.Current.SanitizeConfig(ctx)

	response.Status = AppConfigStatus{
		ConfigLoaded:     config.Loaded,
		ConfigValid:      (config.Valid && config.MediuxValid && config.MediaServerValid),
		NeedsSetup:       !(config.Loaded && config.Valid && config.MediuxValid && config.MediaServerValid),
		CurrentSetup:     *sanitizedConfig,
		MediaServerName:  config.MediaServerName,
		APIKeyConfigured: config.Current.Auth.APIKeyHash != "",
	}

	httpx.SendResponse(w, ld, response)
}
