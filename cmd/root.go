package cmd

import (
	"github.com/spf13/cobra"
)

func SetupCli() {
	rootCmd.AddCommand(DownloadCmd)
}

var rootCmd = &cobra.Command{
	Use:   "audiotheker",
	Short: "`audiotheker allows downloading all episodes of a program in the ARD Audiothek.",
	Long: `audiotheker allows downloading all episodes of a program in the ARD Audiothek.
It queries the official GraphQL API to gather the download URLs.`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
