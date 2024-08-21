package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cli/go-gh/pkg/auth"
	"github.com/gookit/ini"
	"github.com/spf13/cobra"
)

type JWTToken struct {
	Token string `json:"token"`
}

type GHResponse struct {
	Id int `json:"id"`
}

type GHUser struct {
	Username string `json:"login"`
}

func assumeRole(token string, arn string, user string) *sts.AssumeRoleWithWebIdentityOutput {
	// assumeRole will assume the AWS role from the provided ARN using the
	// AWS AssumeRoleWithWebIdentity functionality
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-2"),
	)

	if err != nil {
		fmt.Printf("Error while loading config %v", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	sessionName := fmt.Sprintf("nv-gha-aws-%s", user)
	var durationSessions int32 = 43200 // 12 Hours in Seconds

	input := sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &arn,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &token,
		DurationSeconds:  &durationSessions,
	}

	output, err := stsClient.AssumeRoleWithWebIdentity(context.TODO(), &input)
	if err != nil {
		fmt.Printf("Error while assuming AWS Role %v", err)
		os.Exit(1)
	}
	return output
}

func getGHToken() string {
	// getGHToken will get the user's PAT using the GitHub CLI
	authToken, _ := auth.TokenForHost("github.com")
	if authToken == "" {
		fmt.Printf("Error: Unable to getGH Token")
		os.Exit(1)
	}
	return authToken
}

func getFlag(command *cobra.Command, flag string) string {
	// getFlag is a helper function that will get the command's flag
	res, err := command.Flags().GetString(flag)
	if err != nil {
		fmt.Printf("Error retrieving %s flag %v", flag, err)
		os.Exit(1)
	}
	return res
}

func getJWTToken(url string, ghToken string) string {
	// getJWTToken will get the user's JWT token from the OIDC provider
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error while creating request %v", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ghToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %v", err)
		os.Exit(1)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to get response body: %v", err)
		os.Exit(1)
	}
	var jwtToken JWTToken
	if err := json.Unmarshal(body, &jwtToken); err != nil {
		fmt.Printf("Error Unmarshalling JSON %v", err)
		os.Exit(1)
	}
	return jwtToken.Token
}

func prettyPrint(awsCreds *sts.AssumeRoleWithWebIdentityOutput, profile string) {
	// prettyPrint will print the output in the same format as the AWS Credentials file
	if len(profile) > 0 {
		fmt.Printf("[%s]\n", profile)
	}

	fmt.Printf("EXPORT %s=%s\n", "AWS_ACCESS_KEY_ID", *awsCreds.Credentials.AccessKeyId)
	fmt.Printf("EXPORT %s=%s\n", "AWS_SECRET_ACCESS_KEY", *awsCreds.Credentials.SecretAccessKey)
	fmt.Printf("EXPORT %s=%s\n", "AWS_SESSION_TOKEN", *awsCreds.Credentials.SessionToken)
}

func printOutput(creds *sts.AssumeRoleWithWebIdentityOutput, command *cobra.Command) {
	// printOutput will print the AWS Credentials to the Command line based on the
	// format that they provided
	if getFlag(command, "output") == "pretty" {
		prettyPrint(creds, getFlag(command, "profile"))
	} else if getFlag(command, "output") == "json" {
		jsonOutput, err := json.MarshalIndent(creds.Credentials, "", "  ")
		if err != nil {
			fmt.Printf("Error Marshalling JSON for output %v", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", jsonOutput)
	} else {
		fmt.Printf("Invalid Output Format")
		os.Exit(1)
	}
}

func writeToAWSCredsFile(awsCredOutput *sts.AssumeRoleWithWebIdentityOutput, profile string) {
	// writeToAWSCredsFile writes to the user's local ~/.aws/credentials file under the
	// profile that they specify
	if profile == "" {
		profile = "default"
	}

	homeDir, _ := os.UserHomeDir()
	awsCredentialsPath := filepath.Join(homeDir, ".aws", "credentials")

	config, err := ini.LoadFiles(awsCredentialsPath)
	if err != nil {
		fmt.Printf("AWS Credentials file does not exist %v", err)
		os.Exit(1)
	}

	awsCredsMap := map[string]string{
		"aws_access_key_id":     *awsCredOutput.Credentials.AccessKeyId,
		"aws_secret_access_key": *awsCredOutput.Credentials.SessionToken,
		"aws_session_token":     *awsCredOutput.Credentials.SecretAccessKey,
	}

	err = config.SetSection(profile, awsCredsMap)
	if err != nil {
		fmt.Printf("Error setting AWS Credentials file values %v", err)
		os.Exit(1)
	}

	_, err = config.WriteToFile(awsCredentialsPath)
	if err != nil {
		fmt.Printf("Error writing to AWS Credentials file %v", err)
		os.Exit(1)
	}
}

func writeAWSCredentials(creds *sts.AssumeRoleWithWebIdentityOutput, command *cobra.Command) {
	// writeAWSCredentials gets the AWS Credentials and writes them to the user's AWS Credentials File if specified

	writeToAWS, err := command.Flags().GetBool("write-to-file")
	if err != nil {
		fmt.Printf("Error getting write flag %v", err)
		os.Exit(1)
	}

	profile := getFlag(command, "profile")

	if writeToAWS {
		writeToAWSCredsFile(creds, profile)
	} else if !writeToAWS && profile != "" {
		fmt.Print("Must set write-to-file flag if specifying a profile")
		os.Exit(1)
	}
}
