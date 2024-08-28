package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	programName = "nv-gha-aws"
)

var rootCmd = &cobra.Command{
	Use:          programName,
	Short:        "A GitHub CLI Extension Tool to receive AWS Credentials",
	SilenceUsage: true,
}

func Execute(version string) {
	rootCmd.Version = version
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
	rootCmd.PersistentFlags().Int32P("duration", "d", 43200, "The maximum session duration with the temporary AWS Credentials in seconds")
	rootCmd.PersistentFlags().StringP("output", "o", "shell", "Output format of credentials in one of: shell, json, or creds-file format")
	rootCmd.PersistentFlags().BoolP("write", "w", false, "Specifies if Credentials should be written to AWS Credentials file")
	rootCmd.PersistentFlags().StringP("file", "f", "$HOME/.aws/credentials", "File path to write AWS Credentials")
	rootCmd.PersistentFlags().StringP("profile", "p", "default", "Profile where credentials should be written to in AWS Credentials File")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		writeFlag, err := cmd.Flags().GetBool("write")
		if err != nil {
			return fmt.Errorf("failed to get --write flag: %w", err)
		}

		printFormatFlag, err := cmd.Flags().GetString("output")
		if err != nil {
			return fmt.Errorf("failed to get --output flag: %w", err)
		}

		// ensures that the writeFlag must be set in order to set the file flag
		if cmd.Flags().Changed("file") && !writeFlag {
			return errors.New("the write flag must be set if specifying a file path")
		}

		if printFormatFlag != "creds-file" && cmd.Flags().Changed("profile") {
			return errors.New("the profile can only be set if the output flag is set to creds-file")
		}

		if printFormatFlag != "creds-file" && writeFlag {
			return errors.New("the write flag can only be set if the output flag is set to creds-file")
		}

		return nil
	}
}
