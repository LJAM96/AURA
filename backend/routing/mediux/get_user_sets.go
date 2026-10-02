package routes_mediux

import (
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

type GetAllUserSets_Response struct {
	Sets models.CreatorSetsResponse `json:"sets"`
}

// GetAllUserSets godoc
// @Summary      Get Mediux User Sets
// @Description  Retrieve all item sets created by the specified user in Mediux. This endpoint returns an array of item sets, including details such as the set name, type (show, movie, or collection), and the items contained within each set. This allows clients to display the user's custom collections and preferences within the Mediux ecosystem in the UI.
// @Tags         Mediux
// @Accept       json
// @Produce      json
// @Param        username query string true "Username of the user whose sets are being retrieved"
// @Security     SessionCookie
// @Security     ApiKeyAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=GetAllUserSets_Response}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/mediux/sets/user [get]
func GetAllUserSets(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Mediux User Sets", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	var response GetAllUserSets_Response

	username := r.URL.Query().Get("username")
	if username == "" {
		logAction.SetError("Username query parameter is required", "", nil)
		httpx.SendResponse(w, ld, response)
		return
	}

	userSets, Err := mediux.GetAllUserSets(ctx, username)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.Sets = userSets
	httpx.SendResponse(w, ld, response)
}
