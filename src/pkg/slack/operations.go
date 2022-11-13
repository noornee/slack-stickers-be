package slack

import (
	"fmt"

	"github.com/slack-go/slack"

	"github.com/odetolakehinde/slack-stickers-be/src/pkg/helper"
)

// Push sends the message to the specified Slack channel
func (p *Provider) Push(title, msg, slackChannelID string, data map[string]string) error {
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
		p.logger.Err(err).Str(helper.LogStrKeyMethod, "push").Msg("slack push message failed")
		return err
	}

	p.logger.Info().Msgf("message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// SendSticker sends the sticker to the conversation.
func (p *Provider) SendSticker(slackChannelID, imageURL string) error {
	log := p.logger.With().Str(helper.LogStrKeyMethod, "SendSticker").Logger()
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
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		log.Err(err).Msg("slack send sticker failed")
		return err
	}

	p.logger.Info().Msgf("sticker successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// ShowSearchModal triggers the modal to show the user to put in the tag they want to use.
func (p *Provider) ShowSearchModal(channelID, triggerID string) error {
	modalRequest := generateSearchModalRequest()
	_, err := p.client.OpenView(triggerID, modalRequest)
	if err != nil {
		fmt.Printf("Error opening view: %s", err)
		return err
	}

	if err != nil {
		p.logger.Err(err).Msg("slack send sticker failed")
		return err
	}

	return nil
}
