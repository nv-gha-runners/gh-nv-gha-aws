package cmd

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Receive AWS Credentials by providing both an Organization Name and a Team Name",
	Args:  cobra.ExactArgs(2),
	Run: func(command *cobra.Command, args []string) {
		orgName := args[0]
		teamName := args[1]

		client, err := api.DefaultRESTClient()
		if err != nil {
			fmt.Printf("Error while initializing REST client %v", err)
			os.Exit(1)
		}

		var orgResponse GHResponse
		var teamResponse GHResponse

		err = client.Get(fmt.Sprintf("orgs/%s", orgName), &orgResponse)
		if err != nil {
			fmt.Printf("Error while accessing org %v", err)
			os.Exit(1)
		}

		err = client.Get(fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName), &teamResponse)
		if err != nil {
			fmt.Printf("Error while accessing Team %v", err)
			os.Exit(1)
		}

		authToken := getGHToken()

		jwtQuery := fmt.Sprintf("%s/gh/team/%d/%d?audience=%s", getFlag(command, "idp-url"), orgResponse.Id, teamResponse.Id, getFlag(command, "aud"))
		jwtToken := getJWTToken(jwtQuery, authToken)

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
	rootCmd.AddCommand(teamCmd)
}
