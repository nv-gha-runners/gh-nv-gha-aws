package gh

import (
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

type client struct {
	restClient *api.RESTClient
	gqlClient  *api.GraphQLClient
}

func GetGHToken() (string, error) {
	ghToken, _ := auth.TokenForHost("github.com")
	if ghToken == "" {
		return "", errors.New("failed to get GH token")
	}
	return ghToken, nil
}

func NewClient(ghToken string) (*client, error) {
	gqlClient, err := api.NewGraphQLClient(api.ClientOptions{
		AuthToken: ghToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}
	restClient, err := api.NewRESTClient(api.ClientOptions{
		AuthToken: ghToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}
	return &client{
		gqlClient:  gqlClient,
		restClient: restClient,
	}, nil
}

func (c *client) GetOrgID(orgName string) (int, error) {
	org := struct {
		ID int `json:"id"`
	}{}

	url := fmt.Sprintf("orgs/%s", orgName)
	err := c.restClient.Get(url, &org)
	if err != nil {
		return 0, fmt.Errorf("failed to query /%s endpoint: %w", url, err)
	}

	return org.ID, nil
}

func (c *client) GetTeamID(orgName string, teamName string) (int, error) {
	team := struct {
		ID int `json:"id"`
	}{}

	url := fmt.Sprintf("orgs/%s/teams/%s", orgName, teamName)
	err := c.restClient.Get(url, &team)
	if err != nil {
		return 0, fmt.Errorf("failed to query /%s endpoint: %w", url, err)
	}

	return team.ID, nil
}

func (c *client) GetEnterpriseID(enterpriseName string) (string, error) {
	vars := map[string]interface{}{}
	res := struct {
		Enterprise struct {
			Id string
		}
	}{}

	query := fmt.Sprintf("query { enterprise(slug: \"%s\") { id } }", enterpriseName)
	err := c.gqlClient.Do(query, vars, &res)
	if err != nil {
		return "", fmt.Errorf("failed to %s endpoint: %w", query, err)
	}

	return res.Enterprise.Id, nil
}

func (c *client) GetUsername() (string, error) {
	user := struct {
		Username string `json:"login"`
	}{}

	url := "user"
	err := c.restClient.Get(url, &user)
	if err != nil {
		return "", fmt.Errorf("failed to query /%s endpoint: %w", url, err)
	}

	return user.Username, nil
}
