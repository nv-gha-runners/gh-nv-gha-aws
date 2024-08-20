# gh-nv-gha-aws

`gh-nv-gha-aws` is a Golang-based GitHub CLI Extension used to provide temporary AWS Credentials for developers using a GitHub PAT.

## Steps to install

1. Please ensure that you have the `gh` CLI tool [installed](https://docs.github.com/en/github-cli/github-cli/quickstart).

2. Login and authenticate with `gh auth login`. This is required to ensure you have correct credentials to receive the AWS Credentials.

3. Run `gh extension install nv-gha-aws`

## Usage Details
There are two main subcommands that can be run in order to receive temporary AWS Credentials.

1. `gh nv-gha-aws org <ORG_NAME>`
    - This commands requires you to provide an organization name and will check if you are a member of this organization and will authenticate you if you are.

2. `gh nv-gha-aws org <ORG_NAME> team <TEAM_NAME>`
    - This commands requires you to provide an organization name and team name and will check if you are a member of this organization and this team and will authenticate you if you are.

- Both subcommands will require you to pass in a `role-arn` using the `--role-arn` flag and if you would like this command to automatically write credentials to your `~/.aws/credentials` file, then please pass the `-w` flag and the profile you would like to specify it to using the `-p` flag (it uses the `default` profile otherwise). More information about all of the available flags is available when running `gh nv-gha-aws --help`

