package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	programName = "nv-gha-aws"
)

var rootCmd = &cobra.Command{
	Use:   programName,
	Short: "A GitHub CLI Extension Tool to receive AWS Credentials",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("role-arn", "", "Role ARN ")
	_ = rootCmd.MarkFlagRequired("role-arn")

	rootCmd.PersistentFlags().String("idp-url", "https://token.gha-runners.nvidia.com", "Identity Provider URL")
	rootCmd.PersistentFlags().String("aud", "sts.amazonaws.com", "Audience of Web Identity Token")
	rootCmd.PersistentFlags().StringP("output", "o", "pretty", "Output format of credentials in either pretty or json format")
	rootCmd.PersistentFlags().BoolP("write-to-file", "w", false, "Specifies if Credentials should be written to AWS Credentials file")
	rootCmd.PersistentFlags().StringP("profile", "p", "", "Profile where credentials should be written in AWS Credentials File")
}
