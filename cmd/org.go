package cmd

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Receive AWS Credentials by providing an organization name",
	Args:  cobra.ExactArgs(1),
	Run: func(command *cobra.Command, args []string) {
		orgName := args[0]

		client, err := api.DefaultRESTClient()
		if err != nil {
			fmt.Printf("Error while initializing REST client %v", err)
			os.Exit(1)
		}

		var response GHResponse
		err = client.Get(fmt.Sprintf("orgs/%s", orgName), &response)
		if err != nil {
			fmt.Printf("Error while accessing org %v", err)
			os.Exit(1)
		}

		ghToken := getGHToken()

		jwtQuery := fmt.Sprintf("%s/gh/org/%d?audience=%s", getFlag(command, "idp-url"), response.Id, getFlag(command, "aud"))
		jwtToken := getJWTToken(jwtQuery, ghToken)

		var user GHUser
		err = client.Get("user", &user)
		if err != nil {
			fmt.Printf("Error while accessing user information %v", err)
			os.Exit(1)
		}

		awsCredOutput := assumeRole(jwtToken, getFlag(command, "role-arn"), user.Username)

		writeAWSCredentials(awsCredOutput, command)
		printOutput(awsCredOutput, command)
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
}
