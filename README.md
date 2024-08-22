# gh-nv-gha-aws

`gh-nv-gha-aws` is a Golang-based GitHub CLI Extension used to provide temporary AWS Credentials for developers using a GitHub PAT.

## Steps to install

1. Please ensure that you have the `gh` CLI tool [installed](https://docs.github.com/en/github-cli/github-cli/quickstart).

2. Login and authenticate with `gh auth login`. This is required to ensure you have correct credentials to receive the AWS Credentials.

3. Run `gh extension install nv-gha-aws`

## Usage Details
There are two main subcommands that can be run in order to receive temporary AWS Credentials.

1. `gh nv-gha-aws org <ORG_NAME>`
    - This commands requires you to provide an organization name and will provide temporary AWS credentials if you are a member of a valid organization. 

2. `gh nv-gha-aws team <ORG_NAME> <TEAM_NAME>`
    - This commands requires you to provide an organization name and team name and will provide temporary AWS Credentials if you are a member of a valid organization and team.

Both subcommands will require you to pass in a `role-arn` using the `--role-arn` flag. 

There are three possible output formats: `shell` (default), `json` and `creds-file`. The `shell` output will provide AWS Credentials that can be exported as [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html#envvars-set). The `creds-file` output will provide AWS Credentials in the same format as an [AWS Credentials File](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-files.html#cli-configure-files-format)

If you would like this command to automatically write credentials to a file, then please pass the `-w` flag, the file you would like the credentials to be written to (defaults to `~/.aws/credentials`), and the profile you would like to write to using the `-p` flag (defaults to the `default` profile). **Please note** that the `output` flag must be set to `creds-file` in order to write credentials to a file.

More information about all of the available flags is available when running `gh nv-gha-aws --help`

