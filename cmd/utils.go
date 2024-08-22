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

func assumeRole(cmd *cobra.Command, token string, arn string, user string) *sts.AssumeRoleWithWebIdentityOutput {
	// assumeRole will assume the AWS role from the provided ARN using the
	// AWS AssumeRoleWithWebIdentity functionality
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-2"),
	)

	if err != nil {
		fmt.Printf("Error while loading config %v", err)
		os.Exit(1)
	}

	stsClient := sts.NewFromConfig(cfg)
	sessionName := fmt.Sprintf("nv-gha-aws-%s", user)
	sessionDuration, err := cmd.Flags().GetInt32("duration")

	if err != nil {
		fmt.Printf("Error while getting duration flag %v", err)
		os.Exit(1)
	}

	input := sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &arn,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &token,
		DurationSeconds:  &sessionDuration,
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

func printCredsFormat(awsCreds *sts.AssumeRoleWithWebIdentityOutput, profile string) {
	// printCredsFormat will print in the format of the AWS Credentials File
	fmt.Printf("[%s]\n", profile)
	fmt.Printf("%s=%s\n", "aws_access_key_id", *awsCreds.Credentials.AccessKeyId)
	fmt.Printf("%s=%s\n", "aws_secret_access_key", *awsCreds.Credentials.SecretAccessKey)
	fmt.Printf("%s=%s\n", "aws_session_token", *awsCreds.Credentials.SessionToken)
}

func printShellFormat(awsCreds *sts.AssumeRoleWithWebIdentityOutput) {
	// printShellFormat will print in format of exporting AWS Credentials Environment Variables
	fmt.Printf("EXPORT %s=%s\n", "AWS_ACCESS_KEY_ID", *awsCreds.Credentials.AccessKeyId)
	fmt.Printf("EXPORT %s=%s\n", "AWS_SECRET_ACCESS_KEY", *awsCreds.Credentials.SecretAccessKey)
	fmt.Printf("EXPORT %s=%s\n", "AWS_SESSION_TOKEN", *awsCreds.Credentials.SessionToken)
}

func printOutput(creds *sts.AssumeRoleWithWebIdentityOutput, command *cobra.Command) {
	// printOutput will print the AWS Credentials to the Command line based on the
	// format that they provided
	outputFlag := getFlag(command, "output")
	if outputFlag == "shell" {
		printShellFormat(creds)
	} else if outputFlag == "json" {
		jsonOutput, err := json.MarshalIndent(creds.Credentials, "", "  ")
		if err != nil {
			fmt.Printf("Error Marshalling JSON for output %v", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", jsonOutput)
	} else if outputFlag == "creds-file" {
		profile := getFlag(command, "profile")
		printCredsFormat(creds, profile)
	} else {
		fmt.Printf("Invalid Output Format")
		os.Exit(1)
	}
}

func writeCredsToFile(awsCredOutput *sts.AssumeRoleWithWebIdentityOutput, path string, profile string) {
	// writeCredsToFile writes to the user's specified filepath with the profile that they specify

	homeDir, _ := os.UserHomeDir()
	credFilePath := filepath.Join(homeDir, path)

	config, err := ini.LoadExists(credFilePath)
	if err != nil {
		fmt.Printf("Error loading %s. %v", credFilePath, err)
		os.Exit(1)
	}

	awsCredsMap := map[string]string{
		"aws_access_key_id":     *awsCredOutput.Credentials.AccessKeyId,
		"aws_secret_access_key": *awsCredOutput.Credentials.SessionToken,
		"aws_session_token":     *awsCredOutput.Credentials.SecretAccessKey,
	}

	err = config.SetSection(profile, awsCredsMap)
	if err != nil {
		fmt.Printf("Error setting %s values %v", credFilePath, err)
		os.Exit(1)
	}

	// Will create file if file does not exist
	_, err = config.WriteToFile(credFilePath)
	if err != nil {
		fmt.Printf("Error writing to %s %v", credFilePath, err)
		os.Exit(1)
	}
}

func writeAWSCredentials(creds *sts.AssumeRoleWithWebIdentityOutput, command *cobra.Command) {
	// writeAWSCredentials gets the AWS Credentials and writes them to the user's AWS Credentials File if specified

	writeToAWS, err := command.Flags().GetBool("write")
	if err != nil {
		fmt.Printf("Error getting write flag %v", err)
		os.Exit(1)
	}

	if writeToAWS {
		filepath := getFlag(command, "file")
		profile := getFlag(command, "profile")
		writeCredsToFile(creds, filepath, profile)
	}
}
