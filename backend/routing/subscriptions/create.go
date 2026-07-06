package routes_subscriptions

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
)

type createSubscriptionRequest struct {
	Username       string              `json:"username"`
	CreatorID      string              `json:"creator_id"`
	ImageTypes     models.SelectedTypes `json:"image_types"`
	MediaScope     string              `json:"media_scope"`
	LibrarySection *string             `json:"library_section"`
	Enabled        bool                `json:"enabled"`
}

type createSubscriptionResponse struct {
	ID int `json:"id"`
}

// CreateSubscription godoc
// @Summary      Create User Subscription
// @Description  Create or update a subscription for a MediUX creator. When a subscription is active, new sets from this creator matching the specified image types and media scope will be automatically downloaded.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        req  body      createSubscriptionRequest  true  "Subscription Configuration"
// @Security 	 BearerAuth
// @Failure      401  {object}  httpx.UnauthorizedResponse "Unauthorized (only when Auth.Enabled=true)"
// @Success	  200  {object}  httpx.JSONResponse{data=createSubscriptionResponse}
// @Failure	  500  {object}  httpx.JSONResponse "Internal Server Error"
// @Router       /api/subscriptions [post]
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Create User Subscription", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var req createSubscriptionRequest
	var response createSubscriptionResponse

	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Create Subscription - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	// Validate required fields
	if req.Username == "" {
		logAction.SetError("Username is required", "", nil)
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
		Username:       req.Username,
		CreatorID:      req.CreatorID,
		ImageTypes:     req.ImageTypes,
		MediaScope:     req.MediaScope,
		LibrarySection: req.LibrarySection,
		Enabled:        req.Enabled,
	}

	id, Err := database.CreateSubscription(ctx, sub)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, response)
		return
	}

	response.ID = id
	httpx.SendResponse(w, ld, response)
}
