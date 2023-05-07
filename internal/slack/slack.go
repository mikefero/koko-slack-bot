/*
Copyright © 2023 Kong, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package slack

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kong/koko-slack-bot/internal/github"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"
)

// Options contain the parameters to create a new Slack instance.
type Options struct {
	// AppToken represents the SLACK_APP_TOKEN for the client
	AppToken string
	// BotToken represents the SLACK_BOT_TOKEN for the client
	BotToken string
	// Debug represents a toggling flag to enable/disable debug output
	Debug bool
	// GitHubClient represents the client for accessing GitHub's API
	GitHubClient *github.Client
	// Logger represents the base logger to use for the Slack package
	Logger *zap.Logger
}

// Slack represents a slack instance.
type Slack struct {
	// client represents the API connection to the Slack service
	client *slack.Client
	// botID represents the bot identifier for this Slack application
	botID string
	// gitHubClient represents the client for accessing GitHub's API
	gitHubClient *github.Client
	// handler represents the registration mechanism for Slack events
	handler *socketmode.SocketmodeHandler
	// logger represents the logger to use for the Slack package
	logger *zap.Logger
}

// gatewaySchemaChange represents the response from the processing of the
// gateway schema change event.
type gatewaySchemaChange struct {
	// organization represents the GitHub organization/owner
	organization string
	// pullRequest represents the GitHub pull request number associated with the
	// schema change event
	pullRequest int
	// repository represents the GitHub repository
	repository string
}

// wrapped logger represents the logger for go-slack logger interface.
type wrappedLogger struct {
	// Logger represents the logger to use for the wrapped message
	logger *zap.Logger
}

// Output is the implementation of the go-slack logger interface for use with
// Zap logger.
func (w *wrappedLogger) Output(_ int, message string) error {
	w.logger.Debug(message)
	return nil
}

// NewSlack will validate options and instantiate a new Slack instance.
func NewSlack(opts Options) (*Slack, error) {
	// Validate required options
	if len(strings.TrimSpace(opts.AppToken)) == 0 {
		return nil, errors.New("slack application token is not set")
	}
	if !strings.HasPrefix(opts.AppToken, "xapp-") {
		return nil, errors.New("slack application token must have the prefix 'xapp-'")
	}
	if len(strings.TrimSpace(opts.BotToken)) == 0 {
		return nil, errors.New("slack bot token is not set")
	}
	if !strings.HasPrefix(opts.BotToken, "xoxb-") {
		return nil, errors.New("slack bot token must have the prefix 'xoxb-'")
	}
	if opts.GitHubClient == nil {
		return nil, errors.New("client for GitHub is not set")
	}
	if opts.Logger == nil {
		return nil, errors.New("logger is not set")
	}

	// Create the Slack logger
	logger := opts.Logger.With(zap.String("component", "slack"))

	// Initialize the API Slack client using the wrapped zap logger
	client := slack.New(
		opts.BotToken,
		slack.OptionDebug(opts.Debug),
		slack.OptionLog(&wrappedLogger{logger: logger.With(zap.String("component", "api"))}),
		slack.OptionAppLevelToken(opts.AppToken),
	)

	// Initialize the socketmode socketClient using the wrapped zap logger
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(opts.Debug),
		socketmode.OptionLog(&wrappedLogger{logger: logger.With(zap.String("component", "scocketmode"))}),
	)

	return &Slack{
		client:       client,
		gitHubClient: opts.GitHubClient,
		handler:      socketmode.NewSocketmodeHandler(socketClient),
		logger:       logger,
	}, nil
}

// Run will register all the handlers and begin processing Slack events.
func (s *Slack) Run() error {
	// Determine the bot ID for the Slack application
	res, err := s.client.AuthTest()
	if err != nil {
		return fmt.Errorf("unable to determine bot ID: %w", err)
	}
	s.botID = res.BotID

	// Handle basic socketmode events
	s.handler.Handle(socketmode.EventTypeConnecting, func(e *socketmode.Event, c *socketmode.Client) {
		s.logger.Info("Connecting to Slack with Socket Mode")
	})
	s.handler.Handle(socketmode.EventTypeConnected, func(e *socketmode.Event, c *socketmode.Client) {
		s.logger.Info("Connected to Slack with Socket Mode")
	})
	s.handler.Handle(socketmode.EventTypeConnectionError, func(e *socketmode.Event, c *socketmode.Client) {
		s.logger.Info("Connection failed; retrying connection")
	})
	s.handler.Handle(socketmode.EventTypeHello, func(e *socketmode.Event, c *socketmode.Client) {
		s.logger.Info("Received hello event")
	})

	// Handle message events from channels, DMs, and groups
	s.handler.HandleEvents(slackevents.Message, s.handleMessageEvent)

	// Start handling Slack events
	err = s.handler.RunEventLoop()
	if err != nil {
		return fmt.Errorf("error running handler event loop: %w", err)
	}
	return nil
}

// handleMessageEvents will process only messages while routing and handling
// them appropriately.
func (s *Slack) handleMessageEvent(e *socketmode.Event, c *socketmode.Client) {
	// Ensure this is the correct event
	eventsAPIEvent, ok := e.Data.(slackevents.EventsAPIEvent)
	if !ok {
		s.logger.Error("event ignored as it is not an API event", zap.Any("event", e))
		return
	}
	c.Ack(*e.Request)

	// Only handle message events that are not from our own bot application
	messageEvent, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.MessageEvent)
	if !ok {
		s.logger.Debug("event ignored as it is not a message event", zap.Any("event", e))
		return
	}
	// Ensure user messages and bot messages are allowed to be processed
	if len(messageEvent.SubType) > 0 && messageEvent.SubType != "bot_message" {
		s.logger.Debug("event ignored as the sub-type is not a new message", zap.String("sub-type", messageEvent.SubType))
		return
	}
	// Do not handle messages from Koko's Slack bot
	if messageEvent.BotID == s.botID {
		s.logger.Debug("event ignored as the message originated from our application",
			zap.String("bot-id", messageEvent.BotID))
		return
	}

	// Get the channel where the message event occurred
	conversationInfo, err := c.GetConversationInfo(&slack.GetConversationInfoInput{ChannelID: messageEvent.Channel})
	if err != nil {
		s.logger.Error("unable to get channel information", zap.Error(err))
		return
	}
	channelName := conversationInfo.Name

	switch messageEvent.SubType {
	case "bot_message":
		if err := s.handleBotMessage(messageEvent, channelName); err != nil {
			s.logger.Error("unable to handle bot message", zap.Error(err))
		}
	default:
		// Get the username of the person who sent the message
		userInfo, err := c.GetUserInfo(messageEvent.User)
		if err != nil {
			s.logger.Error("unable to get user information", zap.Error(err))
			return
		}
		username := userInfo.Name

		s.logger.Debug("message received", zap.String("channel", channelName),
			zap.String("username", username),
			zap.String("message", messageEvent.Text))
	}
}

// handleBotMessage will process all bot messages that occur.
func (s *Slack) handleBotMessage(messageEvent *slackevents.MessageEvent, channelName string) error {
	switch channelName {
	case "gateway-schema-change-feed":
		gsc, err := s.handleGatewaySchemaChangeEvent(messageEvent)
		if err != nil {
			return fmt.Errorf("unable to handle gateway schema change event: %w", err)
		}

		// Get the pull request description for the change
		// TODO(fero): add description variable to use with next steps; Jira integration
		_, err = s.gitHubClient.PRDescription(gsc.organization, gsc.repository, gsc.pullRequest)
		if err != nil {
			return fmt.Errorf("unable to get PR description for gateway schema change event: %w", err)
		}
	default:
		s.logger.Debug("bot message received", zap.String("bot-id", messageEvent.BotID), zap.String("channel", channelName))
	}
	return nil
}

// handleGatewaySchemaChangeEvent will process all Kong Gateway schema change
// events that occur from #gateway-schema-change-feed on Slack.
func (s *Slack) handleGatewaySchemaChangeEvent(messageEvent *slackevents.MessageEvent) (gatewaySchemaChange, error) {
	s.logger.Debug("gateway change event received", zap.Any("message-event", messageEvent))
	if len(messageEvent.BotID) == 0 {
		return gatewaySchemaChange{}, errors.New("bot ID is missing from message event")
	}
	numOfAttachments := len(messageEvent.Attachments)
	if numOfAttachments == 0 {
		return gatewaySchemaChange{}, errors.New("attachments are missing from message event")
	}
	if numOfAttachments > 1 {
		return gatewaySchemaChange{}, fmt.Errorf("too many attachments from message event: %d > 1", numOfAttachments)
	}

	// Get author of the commit from the attachment
	author := messageEvent.Attachments[0].AuthorName
	if len(author) == 0 {
		return gatewaySchemaChange{}, errors.New("gateway change event is missing author")
	}

	// Get the organization, repository, and pull request from the attachment
	var pullRequest int64 = math.MinInt64
	var organization string
	var repository string
	for _, field := range messageEvent.Attachments[0].Fields {
		switch strings.ToLower(field.Title) {
		case "ref":
			tokens := strings.Split(field.Value, "/")
			numOfTokens := len(tokens)
			if numOfTokens < 3 {
				return gatewaySchemaChange{},
					fmt.Errorf("not enough tokens in ref value to parse pull request: %d < 3", numOfTokens)
			}
			var err error
			pullRequest, err = strconv.ParseInt(tokens[2], 10, 0)
			if err != nil {
				return gatewaySchemaChange{}, fmt.Errorf("unable to convert pull request to number: %s", tokens[2])
			}
		case "commit":
			tokens := strings.Split(field.Value, "|")
			numOfTokens := len(tokens)
			if numOfTokens != 2 {
				return gatewaySchemaChange{}, fmt.Errorf("invalid commit value format: %s", field.Value)
			}
			tokens = strings.Split(tokens[0], "/")
			numOfTokens = len(tokens)
			if numOfTokens < 5 {
				return gatewaySchemaChange{},
					fmt.Errorf("not enough tokens in commit value to parse repository: %d < 5", numOfTokens)
			}
			organization = tokens[3]
			repository = tokens[4]
		}
	}

	// Ensure the attachments contains the appropriate information
	if pullRequest == math.MinInt64 {
		return gatewaySchemaChange{}, errors.New("pull request number was not present in message event")
	}
	if len(organization) == 0 {
		return gatewaySchemaChange{}, errors.New("organization was not present in message event")
	}
	if len(repository) == 0 {
		return gatewaySchemaChange{}, errors.New("repository was not present in message event")
	}

	return gatewaySchemaChange{
		organization: organization,
		pullRequest:  int(pullRequest),
		repository:   repository,
	}, nil
}
