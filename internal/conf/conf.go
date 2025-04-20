package conf

import (
	"net/url"
	"time"
)

// GlobalConfiguration holds all the configuration that applies to all instances.
type GlobalConfiguration struct {
	API APIConfiguration
}

type APIConfiguration struct {
	Host               string
	Port               string `envconfig:"PORT" default:"8081"`
	Endpoint           string
	RequestIDHeader    string        `envconfig:"REQUEST_ID_HEADER"`
	ExternalURL        string        `json:"external_url" envconfig:"API_EXTERNAL_URL" required:"true"`
	MaxRequestDuration time.Duration `json:"max_request_duration" split_words:"true" default:"10s"`
}

func (c *APIConfiguration) Validate() error {
	_, err := url.ParseRequestURI(c.ExternalURL)
	if err != nil {
		return err
	}

	return nil
}
