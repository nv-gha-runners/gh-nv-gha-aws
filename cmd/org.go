package cmd

import (
	"fmt"
	"log"
	"strconv"

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
			log.Fatalf("Error while initializing REST client %v", err)
		}

		var response GHResponse
		err = client.Get(fmt.Sprintf("orgs/%s", orgName), &response)
		if err != nil {
			log.Fatalf("Error while accessing org %v", err)
		}

		ghToken := getAuthToken()

		orgId := strconv.Itoa(response.Id)
		jwtQuery := fmt.Sprintf("%s/gh/org/%s?audience=%s", getFlag(command, "idp-url"), orgId, getFlag(command, "aud"))
		jwtToken := getJWTToken(jwtQuery, ghToken)

		awsCredOutput := assumeRole(jwtToken, getFlag(command, "role-arn"))
		creds := getCreds(awsCredOutput)

		writeAWSCredentials(creds, command)
		printOutput(creds, command)
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
}
