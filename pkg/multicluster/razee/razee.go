package razee

import (
	"context"
	"github.com/machinebox/graphql"
	"net/http"
	"time"
)

//nolint:golint,unused
type RazeeGithubApiRoundTripper struct {
	apiKey string
}

//nolint:golint,unused
func (t *RazeeGithubApiRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Add("x-api-key", t.apiKey)
	return http.DefaultTransport.RoundTrip(request)
}

const TokenValidityDuration = 40 * time.Minute

//nolint:golint,unused
type LocalAuthClient struct {
	url            string
	login          string
	password       string
	token          string
	tokenTimestamp time.Time
}

// Authentication method used by sat-con-client library
//nolint:golint,unused
func (t *LocalAuthClient) Authenticate(request *http.Request) error {
	if t.token == "" || time.Since(t.tokenTimestamp) >= TokenValidityDuration {
		req := graphql.NewRequest(`
		mutation ($login: String!, $password: String!) {
		  signIn(
			login: $login
			password: $password
		  ) {
			token
		  }
		}
	`)

		req.Var("login", t.login)
		req.Var("password", t.password)
		var result struct {
			Signin struct {
				Token string `json:"token"`
			} `json:"signIn"`
		}

		client := graphql.NewClient(t.url)

		err := client.Run(context.Background(), req, &result)

		if err != nil {
			return err
		}
		t.token = result.Signin.Token
	}

	request.Header.Add("Authorization", "Bearer "+t.token)
	return nil
}
