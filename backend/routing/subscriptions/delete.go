package routes_subscriptions

import (
	"aura/database"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
	"strconv"
)

type deleteSubscriptionResponse struct {
	Success bool `json:"success"`
}

// DeleteSubscription godoc
// @Summary      Delete User Subscription
// @Description  Delete a user subscription. This stops automatic downloading of new sets from the creator. Existing saved sets are not affected.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Subscription ID"
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=deleteSubscriptionResponse}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/subscriptions/{id} [delete]
func DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Delete User Subscription", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var response deleteSubscriptionResponse

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logAction.SetError("Invalid subscription ID", "", map[string]any{"id": idStr})
		httpx.SendResponse(w, ld, response)
		return
	}

	Err := database.DeleteSubscription(ctx, id)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.Success = true
	httpx.SendResponse(w, ld, response)
}
