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
package github

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("GitHub", func() {
	var logger *zap.Logger

	BeforeEach(func() {
		var err error
		logger, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("creating a new GitHub instance", Label("github-options"), func() {
		When("the options are valid", func() {
			It("a new GitHub client instance will be instantiated", func() {
				c, err := NewClient(Options{
					Token:  "token",
					Logger: logger,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())
			})
		})

		When("the options are invalid", func() {
			Context("and the GitHub token is missing", func() {
				It("a new GitHub client instance will not be instantiated", func() {
					c, err := NewClient(Options{})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("github token is not set"))
					Expect(c).To(BeNil())
				})
			})

			Context("and the logger is missing", func() {
				It("a new GitHub client instance will not be instantiated", func() {
					c, err := NewClient(Options{
						Token: "token",
					})
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("logger is not set"))
					Expect(c).To(BeNil())
				})
			})
		})
	})
})
