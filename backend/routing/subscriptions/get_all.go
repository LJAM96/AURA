package routes_subscriptions

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

type getAllSubscriptionsResponse struct {
	Subscriptions []models.UserSubscription `json:"subscriptions"`
}

// GetAllSubscriptions godoc
// @Summary      Get All User Subscriptions
// @Description  Retrieve all user subscriptions. Returns an array of subscriptions with their configuration including image types, media scope, and enabled status.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=getAllSubscriptionsResponse}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/subscriptions [get]
func GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All User Subscriptions", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var response getAllSubscriptionsResponse

	subs, Err := database.GetAllSubscriptions(ctx)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.Subscriptions = subs
	httpx.SendResponse(w, ld, response)
}
