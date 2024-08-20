package cmd

import (
	"fmt"
	"log"
	"strconv"

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
			log.Fatalf("Error while initializing REST client %v", err)
		}

		var orgResponse GHResponse
		var teamResponse GHResponse

		err = client.Get(fmt.Sprintf("orgs/%s", orgName), &orgResponse)
		if err != nil {
			log.Fatalf("Error while accessing org %v", err)
		}

		err = client.Get(fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName), &teamResponse)
		if err != nil {
			log.Fatalf("Error while accessing Team %v", err)
		}

		authToken := getAuthToken()

		orgId := strconv.Itoa(orgResponse.Id)
		teamId := strconv.Itoa(teamResponse.Id)
		jwtQuery := fmt.Sprintf("%s/gh/team/%s/%s?audience=%s", getFlag(command, "idp-url"), orgId, teamId, getFlag(command, "aud"))
		jwtToken := getJWTToken(jwtQuery, authToken)

		awsCredOutput := assumeRole(jwtToken, getFlag(command, "role-arn"))
		creds := getCreds(awsCredOutput)

		writeAWSCredentials(creds, command)
		printOutput(creds, command)
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
}
