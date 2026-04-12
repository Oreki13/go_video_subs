package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "go_video_subs",
	Short: "go_video_subs is a tool to generate subtitles for videos",
	Long:  `go_video_subs is a tool to generate subtitles for videos`,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}