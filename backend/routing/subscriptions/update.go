package routes_subscriptions

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
	"strconv"
)

type updateSubscriptionRequest struct {
	CreatorID      string               `json:"creator_id"`
	ImageTypes     models.SelectedTypes `json:"image_types"`
	MediaScope     string               `json:"media_scope"`
	LibrarySection *string              `json:"library_section"`
	Enabled        bool                 `json:"enabled"`
}

type updateSubscriptionResponse struct {
	Success bool `json:"success"`
}

// UpdateSubscription godoc
// @Summary      Update User Subscription
// @Description  Update an existing user subscription configuration. This allows changing image types, media scope, library section, or enabled status.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      int                      true  "Subscription ID"
// @Param        req  body      updateSubscriptionRequest true  "Updated Subscription Configuration"
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=updateSubscriptionResponse}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/subscriptions/{id} [put]
func UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Update User Subscription", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var req updateSubscriptionRequest
	var response updateSubscriptionResponse

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logAction.SetError("Invalid subscription ID", "", map[string]any{"id": idStr})
		httpx.SendResponse(w, ld, response)
		return
	}

	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Update Subscription - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	// Validate media_scope
	switch req.MediaScope {
	case "all", "movies", "shows", "collections":
		// valid
	case "":
		req.MediaScope = "all"
	default:
		logAction.SetError("Invalid media_scope", "Must be one of: all, movies, shows, collections", map[string]any{
			"media_scope": req.MediaScope,
		})
		httpx.SendResponse(w, ld, response)
		return
	}

	sub := models.UserSubscription{
		ID:             id,
		CreatorID:      req.CreatorID,
		ImageTypes:     req.ImageTypes,
		MediaScope:     req.MediaScope,
		LibrarySection: req.LibrarySection,
		Enabled:        req.Enabled,
	}

	Err = database.UpdateSubscription(ctx, sub)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.Success = true
	httpx.SendResponse(w, ld, response)
}
