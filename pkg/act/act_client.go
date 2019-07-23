package act

import (
	"errors"
	"fmt"
)

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

func (c Client) Call(routeEndpoint string, method string, body []byte) (ResponseMeta, error) {
	if routeEndpoint == "" {
		return ResponseMeta{}, errors.New("route-endpoint cannot be blank")
	}

	return Call(RequestMeta{
		ServiceURL: fmt.Sprintf("%s%s", c.ServiceURL, routeEndpoint),
		Token:      c.Token,
		Method:     method,
		Body:       body,
	})
}
