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
	"encoding/json"

	"github.com/kong/koko-slack-bot/internal/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/zap"
)

var _ = Describe("slack", func() {
	var logger *zap.Logger

	BeforeEach(func() {
		var err error
		logger, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("creating a new slack instance", Label("slack-options"), func() {
		When("the options are valid", func() {
			It("a new slack instance will be instantiated", func() {
				s, err := NewSlack(Options{
					AppToken:     "xapp-",
					BotToken:     "xoxb-",
					Logger:       logger,
					GitHubClient: &github.Client{},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(s).NotTo(BeNil())
			})
		})

		When("the options are invalid", func() {
			Context("and the application token is missing", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("slack application token is not set"))
					Expect(s).To(BeNil())
				})
			})

			Context("and the application token is not valid", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{
						AppToken: "invalid",
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("slack application token must have the prefix 'xapp-'"))
					Expect(s).To(BeNil())
				})
			})

			Context("and the bot token is missing", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{
						AppToken: "xapp-",
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("slack bot token is not set"))
					Expect(s).To(BeNil())
				})
			})

			Context("and the bot token is not valid", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{
						AppToken: "xapp-",
						BotToken: "invalid",
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("slack bot token must have the prefix 'xoxb-'"))
					Expect(s).To(BeNil())
				})
			})

			Context("and the GitHub client is missing", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{
						AppToken: "xapp-",
						BotToken: "xoxb-",
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("client for GitHub is not set"))
					Expect(s).To(BeNil())
				})
			})

			Context("and the logger is missing", func() {
				It("a new slack instance will not be instantiated", func() {
					s, err := NewSlack(Options{
						AppToken:     "xapp-",
						BotToken:     "xoxb-",
						GitHubClient: &github.Client{},
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("logger is not set"))
					Expect(s).To(BeNil())
				})
			})
		})
	})

	Describe("processing slack events", Label("slack-events"), func() {
		var s *Slack

		BeforeEach(func() {
			var err error
			s, err = NewSlack(Options{
				AppToken:     "xapp-",
				BotToken:     "xoxb-",
				Logger:       logger,
				GitHubClient: &github.Client{},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
		})

		When("the event is a bot message", Label("bot-message"), func() {
			Context("message is a Kong Gateway schema change event", Label("kong-gateway"), func() {
				var messageEvent slackevents.MessageEvent

				When("message contains valid attachments", func() {
					BeforeEach(func() {
						//nolint:lll
						messageJSON := `{
							"bot_id": "B056LB75L5R",
							"attachments": [
								{
									"color": "2eb886",
									"fallback": "null",
									"id": 1,
									"author_name": "team-koko-bot",
									"author_link": "https://github.com/team-koko-bot",
									"author_icon": "https://github.com/team-koko-bot.png?size=32",
									"fields": [
										{
											"title": "Ref",
											"value": "refs/pull/5291/merge",
											"short": true
										},
										{
											"title": "Event",
											"value": "pull_request_target",
											"short": true
										},
										{
											"title": "Actions URL",
											"value": "<https://github.com/kong/team-koko-bot/commit/180edcde77e59309c21acad4837b53d0c85fc908/checks|Labeler>",
											"short": true
										},
										{
											"title": "Commit",
											"value": "<https://github.com/kong/team-koko-bot/commit/180edcde77e59309c21acad4837b53d0c85fc908|180edc>",
											"short": true
										},
										{
											"title": "Message",
											"value": "null",
											"short": false
										}
									],
									"blocks": null,
									"footer": "<https://github.com/rtCamp/github-actions-library|Powered By rtCamp's GitHub Actions Library>"
								}
							]
						}`
						err := json.Unmarshal([]byte(messageJSON), &messageEvent)
						Expect(err).NotTo(HaveOccurred())
					})

					It("the gateway schema change event will be handled without error", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&messageEvent)
						Expect(err).NotTo(HaveOccurred())
						Expect(gsc).Should(BeEquivalentTo(gatewaySchemaChange{
							organization: "kong",
							repository:   "team-koko-bot",
							pullRequest:  5291,
						}))
					})
				})

				When("message event does not contain the bot id", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("bot ID is missing from message event"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event does not contain attachments", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("attachments are missing from message event"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains too many attachments", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{},
								{},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("too many attachments from message event: 2 > 1"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event does not contain author", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("gateway change event is missing author"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments with invalid ref value", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("not enough tokens in ref value to parse pull request: 2 < 3"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments with invalid pull request number in ref value", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull/not-a-number",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("unable to convert pull request to number: not-a-number"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments with invalid format for commit value", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull/1",
										},
										{
											Title: "Commit",
											Value: "https://github.com/kong",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("invalid commit value format: https://github.com/kong"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments with invalid number of tokens for commit value", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull/1",
										},
										{
											Title: "Commit",
											Value: "https://github.com/kong|sha",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("not enough tokens in commit value to parse repository: 4 < 5"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments however is missing pull request ref", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Commit",
											Value: "https://github.com/kong/team-koko-bot|sha",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("pull request number was not present in message event"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments however is missing organization and repository in commit field", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull/1",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("organization was not present in message event"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})

				When("message event contains attachments however is missing repository in commit field", func() {
					It("the gateway schema change event will not be handled and fail", func() {
						gsc, err := s.handleGatewaySchemaChangeEvent(&slackevents.MessageEvent{
							BotID: "id",
							Attachments: []slack.Attachment{
								{
									AuthorName: "team-koko-bot",
									Fields: []slack.AttachmentField{
										{
											Title: "Ref",
											Value: "refs/pull/1",
										},
										{
											Title: "Commit",
											Value: "https://github.com/kong/|sha",
										},
									},
								},
							},
						})
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError("repository was not present in message event"))
						Expect(gsc).Should(Equal(gatewaySchemaChange{}))
					})
				})
			})
		})
	})
})
