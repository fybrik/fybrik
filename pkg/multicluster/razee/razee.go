package razee

import (
	"net/http"
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
