package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gookit/ini"
)

type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Profile         string
}

type GetCredsInput struct {
	Duration int32
	JWT      string
	Profile  string
	RoleArn  string
	Username string
}

func GetCreds(ctx context.Context, inputs *GetCredsInput) (*Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	output, err := stsClient.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(inputs.RoleArn),
		RoleSessionName:  aws.String(fmt.Sprintf("nv-gha-aws-%s", inputs.Username)),
		WebIdentityToken: aws.String(inputs.JWT),
		DurationSeconds:  aws.Int32(inputs.Duration),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	return &Credentials{
		AccessKeyId:     *output.Credentials.AccessKeyId,
		SecretAccessKey: *output.Credentials.SecretAccessKey,
		SessionToken:    *output.Credentials.SessionToken,
		Profile:         inputs.Profile,
	}, nil
}

func (creds *Credentials) Print(outputType string) error {
	switch outputType {
	case "shell":
		return creds.printShell()
	case "json":
		return creds.printJson()
	case "creds-file":
		return creds.printFile()
	default:
		return fmt.Errorf("invalid output type %s", outputType)
	}
}

func (creds *Credentials) printShell() error {
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKeyId)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
	return nil
}

func (creds *Credentials) printJson() error {
	output, err := json.MarshalIndent(*creds, "", "")
	if err != nil {
		return fmt.Errorf("failed to marshall credentials: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func (creds *Credentials) printFile() error {
	fmt.Printf("[%s]\n", creds.Profile)
	fmt.Printf("aws_access_key_id=%s\n", creds.AccessKeyId)
	fmt.Printf("aws_secret_access_key=%s\n", creds.SecretAccessKey)
	fmt.Printf("aws_session_token=%s\n", creds.SessionToken)
	return nil
}

func (creds *Credentials) Write(path string) error {
	credsFile, err := ini.LoadExists(os.ExpandEnv(path))
	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}

	values := map[string]string{
		"aws_access_key_id":     creds.AccessKeyId,
		"aws_secret_access_key": creds.SecretAccessKey,
		"aws_session_token":     creds.SessionToken,
	}
	err = credsFile.SetSection(creds.Profile, values)
	if err != nil {
		return fmt.Errorf("failed to set section: %w", err)
	}

	_, err = credsFile.WriteToFile(os.ExpandEnv(path))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
