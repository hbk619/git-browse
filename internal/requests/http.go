package requests

import "net/http"

type (
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	AuthorisedHTTPClientOptions struct {
		AuthToken string
	}

	AuthorisedHTTPClient struct {
		opts *AuthorisedHTTPClientOptions
	}
)

func NewAuthorisedHTTPClient(opts *AuthorisedHTTPClientOptions) *AuthorisedHTTPClient {
	return &AuthorisedHTTPClient{
		opts: opts,
	}
}

func (c *AuthorisedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+c.opts.AuthToken)
	return client.Do(req)
}
