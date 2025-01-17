package slack

import (
	"context"
	"fmt"
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

// ShowStickerPreview sends a sticker preview to the specified user as an ephemeral message in the Slack channel
func (p *Provider) ShowStickerPreview(_ context.Context, userID, channelID, tag, imageURL string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShowStickerPreview").Logger()
	sticker := model.StickerBlockMetadata{
		Tag:    tag,
		Index:  0,
		ImgURL: imageURL,
	}

	blocks := createStickerPreviewBlock(sticker, true)

	_, _, err := p.client.PostMessage(
		channelID,
		slack.MsgOptionPostEphemeral(userID),
		slack.MsgOptionBlocks(blocks.BlockSet...),
	)
	if err != nil {
		log.Err(err).Msg("PostMessage failed")
		return err
	}

	return nil
}

// ShuffleStickerPreview updates the sticker preview by replacing the original ephemeral message with a new one
// containing a shuffled sticker, updating the displayed image based on the tag and index.
func (p *Provider) ShuffleStickerPreview(_ context.Context, userID, channelID, responseURL string, sticker model.StickerBlockMetadata) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "ShuffleStickerPreview").Logger()

	block := createStickerPreviewBlock(sticker, true)

	_, _, err := p.client.PostMessage(
		channelID,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionPostEphemeral(userID),
		slack.MsgOptionReplaceOriginal(responseURL),
		slack.MsgOptionBlocks(block.BlockSet...),
	)
	if err != nil {
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
func (p *Provider) SendStickerToChannel(_ context.Context, userID, channelID, responseURL string, sticker model.StickerBlockMetadata) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "SendStickerToChannel").Logger()

	contextElements := []slack.MixedElement{
		slack.NewTextBlockObject(slack.MarkdownType, FooterText, false, false),
		slack.NewImageBlockElement("https://auth.slackstickers.com/img/favicon.81c8692c.svg", "slack stickers logo"),
	}

	blocks := []slack.Block{
		slack.NewImageBlock(
			sticker.ImgURL,
			sticker.Tag,
			model.StickerImageBlockID,
			slack.NewTextBlockObject(slack.PlainTextType, sticker.Tag, false, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_sent by_ <@%s>.", userID), false, false),
			nil, nil,
		),
		slack.NewContextBlock(
			model.StickerContextBlockID,
			contextElements...,
		),
	}

	_, timestamp, err := p.client.PostMessage(
		channelID,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		log.Err(err).Msg("PostMessage failed to send sticker")
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
