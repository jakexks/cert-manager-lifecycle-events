/*
Copyright 2021 Jetstack Ltd.
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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "cert-manager-lifecycle-events",
		Short: "A controller that reports all lifecycle events for Certificates and CertificateRequests",
		Long:  "A controller that reports all lifecycle events for Certificates and CertificateRequests.",
		Run:   root,
	}
)

func init() {
	rootCmd.PersistentFlags().Bool("use-structured-logging", false, "Use structured logging - for production environments")
	rootCmd.PersistentFlags().String("event-url", "", "URL to POST certificate lifecycle events to")
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		// Flags collide or other viper weirdness
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	cobra.OnInitialize(flagsFromEnv)
}

// flagsFromEnv allows users to set flags from environment variables - e.g.
// --use-structured-logging = $CM_LIFECYCLE_EVENTS_USE_STRUCTURED_LOGGING
func flagsFromEnv() {
	viper.SetEnvPrefix("cm_lifecycle_events")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error in main: %s", err.Error())
	}
}
