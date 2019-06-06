## Ona SRE Tooling

A set of useful SRE tools. This project is written Golang (v1.12 and above recommended).

### Building sre-tooling

Before you install sre-tooling, make sure your environment is setup and ready for Golang packages. Install the Golang compiler using your package manager. On Ubuntu, run:

```sh
sudo apt install golang
```

If you haven't already, set the `GOPATH` environment variable (preferably in any of your local shell environment files e.g ~/.bashrc):

```sh
export GOPATH=$HOME/go
```

You will also want to add the bin directory in your GOPATH to your PATH:

```sh
export PATH=$PATH:$GOPATH/bin
```

Instructions assume you have only one directory in your GOPATH.

Now get the latest version of sre-tooling by running:

```sh
go get github.com/onaio/sre-tooling
```

You can check whether the binary is installed by running:

```sh
sre-tooling help
```

### Environment Variables

The following environment variables need to be set for the sre-tooling command to work as expected

- `AWS_ACCESS_KEY_ID`: Required if AWS credentials not configured in ~/.aws/credentials. The AWS access key ID to use to authenticate against the API.
- `AWS_SECRET_ACCESS_KEY`: Required if AWS credentials not configured in ~/.aws/credentials. The AWS access key to use to authenticate against the API.
- `SRE_BILLING_REQUIRED_TAGS`: Required. Comma-separated list of keys that are required for billing e.g `"OwnerList,EnvironmentList,EndDate"`.
- `SRE_SLACK_WEBHOOK_URL`: Not required. Slack Webhook URL to use to send notifications to Slack. If not set, tool will not try to send notifications to Slack.
