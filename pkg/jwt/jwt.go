package jwt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JWTInputs struct {
	Audience string
	GHToken  string
	IDPUrl   string
}

func GetOrgJWT(inputs *JWTInputs, orgID int) (string, error) {
	queryUrl := fmt.Sprintf("%s/gh/org/%d?audience=%s", inputs.IDPUrl, orgID,
		inputs.Audience)
	return getJWT(queryUrl, inputs.GHToken)
}

func GetTeamJWT(inputs *JWTInputs, orgID int, teamID int) (string, error) {
	queryUrl := fmt.Sprintf("%s/gh/team/%d/%d?audience=%s", inputs.IDPUrl, orgID,
		teamID, inputs.Audience)
	return getJWT(queryUrl, inputs.GHToken)
}

func getJWT(queryUrl string, ghToken string) (string, error) {
	req, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ghToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query endpoint: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	token := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return token.Token, nil
}
