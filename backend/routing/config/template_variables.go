package routes_config

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

type templateVariablesResponse struct {
	Variables config.NotificationTemplateVariableCatalog `json:"variables"`
}

// GetNotificationTemplateVariables godoc
// @Summary      Get Notification Template Variables
// @Description  Get all available notification template variables grouped by category and by template type.
// @Tags         Config
// @Produce      json
// @Security     SessionCookie
// @Security     ApiKeyAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success      200  {object}  httpx.JSONResponse{data=routes_config.templateVariablesResponse}
// @Router       /api/config/template-variables [get]
func GetNotificationTemplateVariables(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Notification Template Variables", logging.LevelTrace)
	ctx = logging.WithCurrentAction(ctx, logAction)

	response := templateVariablesResponse{
		Variables: config.GetNotificationTemplateVariableCatalog(),
	}

	httpx.SendResponse(w, ld, response)
}
