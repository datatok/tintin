package main

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func newRootCmd(out io.Writer, args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tintin",
		Short:        "Generate Djobi reports.",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			l, err := logrus.ParseLevel(settings.LogLevel)

			if err != nil {
				fmt.Printf("log level \"%s\" is invalid!", settings.LogLevel)
				l = logrus.InfoLevel
			}

			logrus.SetLevel(l)
		},
	}

	flags := cmd.PersistentFlags()

	cmd.AddCommand(
		newReportBuildCmd(out),
		newWebServerCmd(out),
	)

	settings.AddFlags(flags)

	// We can safely ignore any errors that flags.Parse encounters since
	// those errors will be caught later dxuring the call to cmd.Execution.
	// This call is required to gather configuration information prior to
	// execution.
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.Parse(args)

	return cmd
}
