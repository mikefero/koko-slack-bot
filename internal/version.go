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
package internal

var (
	version   string
	commit    string
	osArch    string
	goVersion string
	buildDate string
)

// Version represents the version of the application.
func Version() string {
	return version
}

// Commit represents the SHA of the application when it was built.
func Commit() string {
	return commit
}

// OSArch represents the architecture of the golang compiler used when the
// application was built.
func OSArch() string {
	return osArch
}

// GoVersion represents the version of the golang compiler used when the
// application was built.
func GoVersion() string {
	return goVersion
}

// BuildDate represents the date the application was built.
func BuildDate() string {
	return buildDate
}
