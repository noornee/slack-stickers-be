// Package media houses all media related APIs
package media

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"

	restModel "github.com/odetolakehinde/slack-stickers-be/src/api/model"
	"github.com/odetolakehinde/slack-stickers-be/src/controller"
	"github.com/odetolakehinde/slack-stickers-be/src/model"
	"github.com/odetolakehinde/slack-stickers-be/src/pkg/environment"
)

type slackHandler struct {
	logger      zerolog.Logger
	controller  controller.Operations
	environment *environment.Env
}

// New creates a new instance of the auth rest handler
func New(r *gin.RouterGroup, l zerolog.Logger, c controller.Operations, env *environment.Env) {
	slack := slackHandler{
		logger:      l,
		controller:  c,
		environment: env,
	}

	slackGroup := r.Group("/slack") // ,slack.controller.Middleware().AuthMiddleware(),

	// Endpoints exposed under Media API Handler
	slackGroup.POST("/send-message", slack.sendMessage())
	slackGroup.POST("/interactivity", slack.interactivityUsed())
	slackGroup.POST("/slash-command", slack.slashCommandUsed())
}

// sendMessage handles authentication for users
// @Summary This uploads stickers into the database. Keep in mind that it checks to ensure that it dfors
// @Description Takes the user email and password and returns user and token details
// @Tags Auth
// @Accept json
// @Param uploadRequest body uploadRequest true "Upload Request"
// @Success 201 {object} model.GenericResponse{data=uploadRequest}
// @Failure 400,401,502 {object} model.GenericResponse{error=model.GenericResponse}
// @Router /api/v1/slack/send-message [post]
func (s slackHandler) sendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var req uploadRequest
		//
		//// run the validation first
		//if err := c.ShouldBindJSON(&req); err != nil {
		//	s.logger.Error().Msgf("%v", err)
		//	restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
		//	return
		//}
		//
		//err := restModel.ValidateRequest(req)
		//if err != nil {
		//	s.logger.Error().Msgf("%v", err)
		//	restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
		//	return
		//}

		err := s.controller.SendSticker(context.Background(), "", "")
		if err != nil {
			s.logger.Error().Msgf("%v", err)
			restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		restModel.OkResponse(c, http.StatusOK, "Message sent successfully", "response")
	}
}

func (s slackHandler) interactivityUsed() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			s.logger.Err(err).Msgf("error parsing response :: %v", err.Error())
		}

		parsedBody, err := url.ParseQuery(string(requestBody))
		if err != nil {
			restModel.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		var (
			i             slack.InteractionCallback
			tag           string
			indexToReturn = "0"
		)
		err = json.Unmarshal([]byte(parsedBody["payload"][0]), &i)
		if err != nil {
			restModel.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		switch i.Type {
		case model.SubmissionViewType:
			if len(i.View.Blocks.BlockSet) > 1 && i.View.Blocks.BlockSet[1].BlockType() == model.BlockTypeImage {
				// they actually wanna send the message. Let us proceed
				var details Block

				err = mapstructure.Decode(i.View.Blocks.BlockSet[1], &details)
				if err != nil {
					s.logger.Error().Msgf("%v", err)
					restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
					return
				}

				channelToSendSticker := i.View.CallbackID
				err = s.controller.SendSticker(context.Background(), channelToSendSticker, details.ImageURL)
				if err != nil {
					s.logger.Error().Msgf("%v", err)
					restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
					return
				}

				c.String(http.StatusOK, "Hurray! You've sent your sticker")
				return
			}

			// this is the initial search
			if i.View.CallbackID == model.InitialDataSearchID {
				// Note there might be a better way to get this info, but I figured this structure out from looking at the json response
				tag = i.View.State.Values["Tag"]["tag"].Value
			}
		case model.BlockActionsViewType:
			if len(i.ActionCallback.BlockActions) > 0 {
				if i.ActionCallback.BlockActions[0].ActionID == model.ActionIDShuffle {
					indexToReturn = i.ActionCallback.BlockActions[0].Value
					tag = i.View.PrivateMetadata
				}
			}
		}

		externalViewID := i.View.ExternalID

		err = s.controller.SearchByTag(context.Background(), i.TriggerID, tag, indexToReturn, i.View.CallbackID, &externalViewID)
		if err != nil {
			s.logger.Error().Msgf("%v", err)
			restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		//restModel.OkResponse(c, http.StatusOK, "Shortcut initiated", "response")
		return
	}
}

func (s slackHandler) slashCommandUsed() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req restModel.ShortcutPayload

		requestBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			s.logger.Err(err).Msgf("error parsing response :: %v", err.Error())
		}

		parsedBody, err := url.ParseQuery(string(requestBody))
		if err != nil {
			restModel.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		var decoder = schema.NewDecoder()
		err = decoder.Decode(&req, parsedBody)
		if err != nil {
			// Handle error;
			s.logger.Err(err).Msg("e don happen")
		}

		if len(req.Text) < 1 {
			// they did not pass anything else asides the slash command
			err = s.controller.ShowSearchModal(context.Background(), req.TriggerID, req.ChannelID)
		} else {
			// something else was passed asides the slash command
			tag := req.Text
			err = s.controller.SearchByTag(context.Background(), req.TriggerID, tag, "0", req.ChannelID, nil)
		}
		if err != nil {
			s.logger.Error().Msgf("%v", err)
			restModel.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		//restModel.OkResponse(c, http.StatusOK, "Slash command initiated", "response")
		return
	}
}
