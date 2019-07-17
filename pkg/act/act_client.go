package act

import (
	"fmt"
	"io"
)

// curl -H "Metadata-Flavor: Google" 'http://metadata/computeMetadata/v1/instance/service-accounts/default/identity?audience=https://download-auctions-server-7ynizip42q-uc.a.run.app'

func NewClient(serviceURL string) (Client, error) {
	token, err := GetToken(serviceURL)
	if err != nil {
		return Client{}, err
	}

	return Client{ServiceURL: serviceURL, Token: token}, nil
}

type Client struct {
	ServiceURL string
	Token      string
}

func (c Client) Call(routeEndpoint string, method string, body io.Reader) (ResponseMeta, error) {
	if routeEndpoint == "/" {
		routeEndpoint = ""
	}

	return Call(RequestMeta{
		ServiceURL: fmt.Sprintf("%s/%s", c.ServiceURL, routeEndpoint),
		Token:      c.Token,
		Method:     method,
		Body:       body,
	})
}
