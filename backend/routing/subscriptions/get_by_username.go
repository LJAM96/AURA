package routes_subscriptions

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

type getSubscriptionByUsernameResponse struct {
	Subscription *models.UserSubscription `json:"subscription"`
	Subscribed   bool                     `json:"subscribed"`
}

// GetSubscriptionByUsername godoc
// @Summary      Get User Subscription by Username
// @Description  Check if a subscription exists for a specific MediUX creator username. Returns the subscription details if found, or an empty response with subscribed=false.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        username query string true "MediUX creator username"
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=getSubscriptionByUsernameResponse}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/subscriptions/username [get]
func GetSubscriptionByUsername(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get User Subscription by Username", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var response getSubscriptionByUsernameResponse

	username := r.URL.Query().Get("username")
	if username == "" {
		logAction.SetError("Username query parameter is required", "", nil)
		httpx.SendResponse(w, ld, response)
		return
	}

	sub, Err := database.GetSubscriptionByUsername(ctx, username)
	if Err.Message != "" {
		response.Subscribed = false
		httpx.SendResponse(w, ld, response)
		return
	}

	response.Subscription = sub
	response.Subscribed = true
	httpx.SendResponse(w, ld, response)
}
