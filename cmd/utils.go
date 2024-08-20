package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cli/go-gh"
	"github.com/gookit/ini"
	"github.com/spf13/cobra"
)

type JWTToken struct {
	Token string `json:"token"`
}

type GHResponse struct {
	Id int `json:"id"`
}

type AWSCreds struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
}

func assumeRole(token string, arn string) *sts.AssumeRoleWithWebIdentityOutput {
	// assumeRole will assume the AWS role from the provided ARN using the
	// AWS AssumeRoleWithWebIdentity functionality
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-2"),
	)

	if err != nil {
		log.Fatalf("Error while loading config %v", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	sessionName := "nv-gha-aws"

	input := sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &arn,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &token,
	}

	output, err := stsClient.AssumeRoleWithWebIdentity(context.TODO(), &input)
	if err != nil {
		log.Fatalf("Error while assuming AWS Role %v", err)
	}
	return output
}

func getAuthToken() string {
	// getAuthToken will get the user's PAT using the GitHub CLI
	authToken, _, err := gh.Exec("auth", "token")
	if err != nil {
		log.Fatalf("Error while getting GH Token %v", err)
	}
	return strings.TrimSpace(authToken.String())
}

func getCreds(awsCredOutput *sts.AssumeRoleWithWebIdentityOutput) *AWSCreds {
	return &AWSCreds{
		AccessKeyId:     *awsCredOutput.Credentials.AccessKeyId,
		SecretAccessKey: *awsCredOutput.Credentials.SecretAccessKey,
		SessionToken:    *awsCredOutput.Credentials.SessionToken,
	}
}

func getFlag(command *cobra.Command, flag string) string {
	// getFlag is a helper function that will get the command's flag
	res, err := command.Flags().GetString(flag)
	if err != nil {
		log.Fatalf("Error retrieving flag %v", err)
	}
	return res
}

func getJWTToken(url string, ghToken string) string {
	// getJWTToken will get the user's JWT token from the OIDC provider
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error while creating request %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ghToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to get response body: %v", err)
	}
	var jwtToken JWTToken
	if err := json.Unmarshal(body, &jwtToken); err != nil {
		log.Fatalf("Error Unmarshalling JSON %v", err)
	}
	return jwtToken.Token
}

func prettyPrint(awsCreds *AWSCreds, profile string) {
	// prettyPrint will print the output in the same format as the AWS Credentials file
	if len(profile) > 0 {
		fmt.Printf("[%s]\n", profile)
	}

	fmt.Printf("%s=%s\n", "aws_access_key_id", awsCreds.AccessKeyId)
	fmt.Printf("%s=%s\n", "aws_secret_access_key", awsCreds.SecretAccessKey)
	fmt.Printf("%s=%s\n", "aws_session_token", awsCreds.SessionToken)
}

func printOutput(creds *AWSCreds, command *cobra.Command) {
	// printOutput will print the AWS Credentials to the Command line based on the
	// format that they provided
	if getFlag(command, "output") == "pretty" {
		prettyPrint(creds, getFlag(command, "profile"))
	} else if getFlag(command, "output") == "json" {
		jsonOutput, err := json.MarshalIndent(creds, "", "  ")
		if err != nil {
			log.Fatalf("Error Marshalling JSON for output %v", err)
		}
		log.Println(jsonOutput)
	} else {
		log.Fatalf("Invalid Output Format")
	}
}

func writeToAWSCredsFile(awsCreds *AWSCreds, profile string) {
	// writeToAWSCredsFile writes to the user's local ~/.aws/credentials file under the
	// profile that they specify
	if profile == "" {
		profile = "default"
	}

	homeDir, _ := os.UserHomeDir()
	awsCredentialsPath := filepath.Join(homeDir, ".aws", "credentials")

	config, err := ini.LoadFiles(awsCredentialsPath)
	if err != nil {
		log.Fatalf("AWS Credentials file does not exist %v", err)
	}

	awsCredsMap := map[string]string{
		"aws_access_key_id":     awsCreds.AccessKeyId,
		"aws_session_token":     awsCreds.SessionToken,
		"aws_secret_access_key": awsCreds.SecretAccessKey,
	}

	err = config.SetSection(profile, awsCredsMap)
	if err != nil {
		log.Fatalf("Error setting AWS Credentials file values %v", err)
	}

	config.WriteToFile(awsCredentialsPath)
}

func writeAWSCredentials(creds *AWSCreds, command *cobra.Command) {
	// writeAWSCredentials gets the AWS Credentials and writes them to the user's AWS Credentials File if specified

	writeToAWS, err := command.Flags().GetBool("write-to-file")
	if err != nil {
		log.Fatalf("Error getting write flag %v", err)
	}

	profile := getFlag(command, "profile")

	if writeToAWS {
		writeToAWSCredsFile(creds, profile)
	} else if !writeToAWS && profile != "" {
		log.Fatal("Must set write-to-file flag if specifying a profile")
	}
}
