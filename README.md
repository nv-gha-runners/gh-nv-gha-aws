# gh-nv-gha-aws

`gh-nv-gha-aws` is a `gh` extension that allows users to obtain temporary AWS credentials for preconfigured IAM roles based on GitHub organization or team membership.

## Steps to install

1. Please ensure that you have the `gh` CLI tool [installed](https://docs.github.com/en/github-cli/github-cli/quickstart).

2. Login and authenticate with `gh auth login`. This is required to ensure you have correct credentials to receive the AWS Credentials.

3. Run `gh extension install nv-gha-aws`

More information about all of the available flags and their associated usage is available when running `gh nv-gha-aws --help`

