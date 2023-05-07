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
package main

import (
	"fmt"
	"os"

	"github.com/kong/koko-slack-bot/internal/github"
	"github.com/kong/koko-slack-bot/internal/slack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var BotID string

func main() {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	gitHubToken := os.Getenv("GITHUB_TOKEN")

	logConfig := zap.NewProductionConfig()
	logConfig.Encoding = "console"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logConfig.Level.SetLevel(zap.DebugLevel)
	logger, err := logConfig.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger: %v", err)
		os.Exit(1)
	}

	githubClient, err := github.NewClient(github.Options{
		Token:  gitHubToken,
		Logger: logger,
	})
	if err != nil {
		logger.Error("unable to create GitHub client", zap.Error(err))
		os.Exit(1)
	}

	s, err := slack.NewSlack(slack.Options{
		AppToken:     appToken,
		BotToken:     botToken,
		Debug:        true,
		GitHubClient: githubClient,
		Logger:       logger,
	})
	if err != nil {
		logger.Error("unable to create Slack instance", zap.Error(err))
		os.Exit(1)
	}
	err = s.Run()
	if err != nil {
		logger.Error("issue running Slack instance", zap.Error(err))
		os.Exit(1)
	}
}
