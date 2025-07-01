package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/slack-go/slack"

	"github.com/odetolakehinde/slack-stickers-be/src/model"
	"github.com/odetolakehinde/slack-stickers-be/src/pkg/helper"
)

// Push sends the message to the specified Slack channel
func (p *Provider) Push(title, msg, slackChannelID string, data map[string]string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "Push").Logger()
	footer := "sandbox mode"

	var fields []slack.AttachmentField
	if len(data) > 0 {
		for k, v := range data {
			fields = append(fields, slack.AttachmentField{
				Title: k,
				Value: v,
			})
		}
	}

	// build a slack attachment
	payload := slack.Attachment{
		Color:  "#F26722",
		Title:  fmt.Sprintf("[%s] - %s", title, msg),
		Fields: fields,
		Footer: footer,
	}
	channelID, timestamp, err := p.client.PostMessage(
		slackChannelID,
		slack.MsgOptionAttachments(payload),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		log.Err(err).Msg("slack push message failed")
		return err
	}

	log.Info().Msgf("message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// SendMessage sends the sticker to the conversation.
func (p *Provider) SendMessage(_ context.Context, slackChannelID, imageURL string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "SendMessage").Logger()
	// build a slack attachment
	payload := slack.Attachment{
		Title:    HeaderText,
		Color:    ColorText,
		Footer:   FooterText,
		ImageURL: imageURL,
	}

	channelID, timestamp, err := p.client.PostMessage(
		slackChannelID,
		slack.MsgOptionAttachments(payload),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Err(err).Msg("slack send sticker failed")
		return err
	}

	log.Info().Msgf("sticker successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// ShowSearchModal triggers the modal to show the user to put in the tag they want to use.
func (p *Provider) ShowSearchModal(_ context.Context, triggerID, channelID string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShowSearchModal").Logger()
	modalRequest := generateSearchModalRequest(channelID)
	_, err := p.client.OpenView(triggerID, modalRequest)
	if err != nil {
		log.Err(err).Msg("OpenView failed")
		return err
	}

	return nil
}

// ShowSearchResultModal triggers the modal to show the user the search result
func (p *Provider) ShowSearchResultModal(_ context.Context, triggerID, channelID string, sticker model.StickerBlockMetadata, externalViewID *string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShowSearchResultModal").Logger()
	var err error

	if externalViewID == nil {
		modalRequest := generateSearchResultModal(channelID, sticker, false)
		if _, err = p.client.OpenView(triggerID, modalRequest); err != nil {
			log.Err(err).Msg("OpenView failed")
			return err
		}
	} else {
		// let us replace what the guy sees on the screen
		modalRequest := generateSearchResultModal(channelID, sticker, true)
		if _, err = p.client.UpdateView(modalRequest, *externalViewID, "", ""); err != nil {
			log.Err(err).Msg("UpdateView failed")
			return err
		}
	}

	return nil
}

// ShowStickerPreview sends a sticker preview message to the specified channel.
// It optionally deletes a previously mentioned message if the bot was mentioned.
//
//   - threadTS: The timestamp of a message in the thread where the preview is posted, if any.
//   - mentionTS: The timestamp of the message to be deleted. This is only valid if the bot was mentioned.
func (p *Provider) ShowStickerPreview(_ context.Context, userID, channelID, tag, imageURL string, threadTS, _ *string, isDM bool, responseURL string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShowStickerPreview").Logger()

	sticker := model.StickerBlockMetadata{
		Tag:      tag,
		Index:    0,
		ImgURL:   imageURL,
		ThreadTS: threadTS,
	}

	blocks := createStickerPreviewBlock(sticker, true)

	// If in DM, use response_url instead of PostMessage
	if isDM && responseURL != "" {
		return p.sendMessageViaResponseURL(responseURL, slack.ResponseTypeEphemeral, blocks.BlockSet)
	}

	msgOptions := []slack.MsgOption{
		slack.MsgOptionPostEphemeral(userID),
		slack.MsgOptionBlocks(blocks.BlockSet...),
	}

	// If threadTs is not nil, include it in the message options
	if sticker.ThreadTS != nil {
		msgOptions = append(msgOptions, slack.MsgOptionTS(*sticker.ThreadTS))
	}

	if _, _, err := p.client.PostMessage(channelID, msgOptions...); err != nil {
		log.Err(err).Msg("PostMessage failed")
		if err.Error() == helper.ErrChannelNotFound.Error() {
			return helper.ErrChannelNotFound
		}
		return err
	}

	return nil
}

// ShuffleStickerPreview updates the sticker preview by replacing the original ephemeral message with a new one
// containing a shuffled sticker, updating the displayed image based on the tag and index.
func (p *Provider) ShuffleStickerPreview(_ context.Context, userID, channelID, responseURL string, sticker model.StickerBlockMetadata) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShuffleStickerPreview").Logger()

	// Create new blocks for the sticker preview
	block := createStickerPreviewBlock(sticker, true)

	// Set up the options for the message
	msgOptions := []slack.MsgOption{
		slack.MsgOptionPostEphemeral(userID),
		slack.MsgOptionBlocks(block.BlockSet...),
		slack.MsgOptionReplaceOriginal(responseURL), // Ensure it replaces the original ephemeral message
	}

	// If threadTs is provided, post to thread
	if sticker.ThreadTS != nil {
		msgOptions = append(msgOptions, slack.MsgOptionTS(*sticker.ThreadTS))
	}

	if _, _, err := p.client.PostMessage(channelID, msgOptions...); err != nil {
		log.Err(err).Msg("PostMessage failed")
		return err
	}

	return nil
}

// CancelStickerPreview removes the sticker preview message from Slack by deleting the original ephemeral message.
func (p *Provider) CancelStickerPreview(_ context.Context, channelID, responseURL string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "CancelStickerPreview").Logger()
	_, _, err := p.client.PostMessage(
		channelID,
		slack.MsgOptionDeleteOriginal(responseURL),
	)
	if err != nil {
		log.Err(err).Msg("failed to cancel sticker preview")
		return err
	}

	return nil
}

// SendStickerToChannel sends the specified sticker to the Slack channel as a permanent message,
func (p *Provider) SendStickerToChannel(_ context.Context, userID, channelID, responseURL string, isDM bool, sticker model.StickerBlockMetadata) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "SendStickerToChannel").Logger()

	contextElements := []slack.MixedElement{
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_sent by_ <@%s> • <%s|Stickers For Slack>", userID, DocsURL), false, false),
	}

	blocks := []slack.Block{
		slack.NewImageBlock(
			sticker.ImgURL,
			sticker.Tag,
			model.StickerImageBlockID,
			slack.NewTextBlockObject(slack.PlainTextType, sticker.Tag, false, false),
		),
		slack.NewContextBlock(
			model.StickerContextBlockID,
			contextElements...,
		),
	}

	// If in DM, use response_url instead of PostMessage
	if isDM && responseURL != "" {
		return p.sendMessageViaResponseURL(responseURL, slack.ResponseTypeInChannel, blocks)
	}

	msgOptions := []slack.MsgOption{
		slack.MsgOptionAsUser(true),
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionDisableMediaUnfurl(),
	}

	// If threadTs is not nil, include it in the message options
	if sticker.ThreadTS != nil {
		msgOptions = append(msgOptions, slack.MsgOptionTS(*sticker.ThreadTS))
	}

	// send sticker
	_, timestamp, err := p.client.PostMessage(
		channelID,
		msgOptions...,
	)
	if err != nil {
		log.Err(err).Msg("PostMessage failed to send sticker")
		if err.Error() == helper.ErrChannelNotFound.Error() {
			return helper.ErrChannelNotFound
		}
		return err
	}

	if !strings.EqualFold(responseURL, "") {
		// responseURL wont be blank if its an ephemeral message
		_, _, err = p.client.PostMessage(
			channelID,
			slack.MsgOptionDeleteOriginal(responseURL), // delete ephemeral message
		)
		if err != nil {
			log.Err(err).Msg("failed to delete original ephemeral message")
			return err
		}
	}

	p.logger.Info().Msgf("sticker successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// ShowHelp to show help
func (p *Provider) ShowHelp(_ context.Context, userID, channelID, responseURL string, isDM bool) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShowHelp").Logger()

	blocks := generateHelpBlocks()

	// If in DM, use response_url instead of PostMessage
	if isDM && responseURL != "" {
		return p.sendMessageViaResponseURL(responseURL, slack.ResponseTypeEphemeral, blocks.BlockSet)
	}

	msgOptions := []slack.MsgOption{
		slack.MsgOptionPostEphemeral(userID),
		slack.MsgOptionBlocks(blocks.BlockSet...),
	}

	if _, _, err := p.client.PostMessage(channelID, msgOptions...); err != nil {
		log.Err(err).Msg("PostMessage failed")
		return err
	}

	return nil
}

func (p *Provider) sendMessageViaResponseURL(responseURL, responseType string, blocks []slack.Block) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "sendMessageViaResponseURL").Logger()

	payload := map[string]any{
		"response_type":   responseType, // "in_channel" or "ephemeral"
		"blocks":          blocks,
		"delete_original": responseType == slack.ResponseTypeInChannel, // this deletes the original ephemeral message when user sends the sticker in chat
	}

	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(responseURL)
	if err != nil {
		log.Err(err).Msg("failed to send responseURL request")
		return fmt.Errorf("failed to send responseURL request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error().
			Int("status", resp.StatusCode()).
			Str("body", resp.String()).
			Msg("non-200 response from response_url")
		return fmt.Errorf("response_url returned non-200: %s", resp.String())
	}

	log.Info().Msg("response sent successfully via response_url")
	return nil
}

// JoinChannel if not present
func (p *Provider) JoinChannel(_ context.Context, channelID string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "JoinChannel").Logger()

	info, err := p.client.GetConversationInfo(channelID, false)
	if err == nil && !info.IsMember {
		_, _, _, err := p.client.JoinConversation(channelID)
		if err != nil {
			log.Err(err).Msg("Failed to join channel")
		}

	}

	return nil
}
