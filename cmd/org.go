package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nv-gha-runners/gh-nv-gha-aws/pkg/aws"
	"github.com/nv-gha-runners/gh-nv-gha-aws/pkg/gh"
	"github.com/nv-gha-runners/gh-nv-gha-aws/pkg/jwt"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Receive AWS Credentials by providing an organization name",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		orgName := args[0]

		idpUrl, err := command.Flags().GetString("idp-url")
		if err != nil {
			return fmt.Errorf("failed to get --idp-url flag: %w", err)
		}

		aud, err := command.Flags().GetString("aud")
		if err != nil {
			return fmt.Errorf("failed to get --aud flag: %w", err)
		}

		roleArn, err := command.Flags().GetString("role-arn")
		if err != nil {
			return fmt.Errorf("failed to get --role-arn flag: %w", err)
		}

		duration, err := command.Flags().GetInt32("duration")
		if err != nil {
			return fmt.Errorf("failed to get --duration flag: %w", err)
		}

		profile, err := command.Flags().GetString("profile")
		if err != nil {
			return fmt.Errorf("failed to get --profile flag: %w", err)
		}

		output, err := command.Flags().GetString("output")
		if err != nil {
			return fmt.Errorf("failed to get --output flag: %w", err)
		}

		write, err := command.Flags().GetBool("write")
		if err != nil {
			return fmt.Errorf("failed to get --write flag: %w", err)
		}

		file, err := command.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to get --file flag: %w", err)
		}

		ghToken, err := gh.GetGHToken()
		if err != nil {
			return err
		}

		ghClient, err := gh.NewClient(ghToken)
		if err != nil {
			return fmt.Errorf("failed to create GH client: %w", err)
		}

		username, err := ghClient.GetUsername()
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}

		orgID, err := ghClient.GetOrgID(orgName)
		if err != nil {
			return fmt.Errorf("failed to get org ID: %w", err)
		}

		jwt, err := jwt.GetOrgJWT(&jwt.JWTInputs{
			Audience: aud,
			GHToken:  ghToken,
			IDPUrl:   idpUrl,
		}, orgID)
		if err != nil {
			return fmt.Errorf("failed to get org JWT: %w", err)
		}

		creds, err := aws.GetCreds(ctx, &aws.GetCredsInput{
			Duration: duration,
			JWT:      jwt,
			Profile:  profile,
			RoleArn:  roleArn,
			Username: username,
		})
		if err != nil {
			return fmt.Errorf("failed to get AWS credentials: %w", err)
		}

		if err = creds.Print(output); err != nil {
			return fmt.Errorf("failed to print credentials: %w", err)
		}

		if write {
			if err = creds.Write(file); err != nil {
				return fmt.Errorf("failed to write credentials file: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
}
