package api

import (
	"net/http"

	"github.com/sebest/xff"
	"github.com/trranminhquang/go-boilerplate/internal/conf"
)

const (
	defaultVersion = "unknown version"
)

type API struct {
	config  *conf.GlobalConfiguration
	handler http.Handler
	version string
}

func (a *API) Config() *conf.GlobalConfiguration {
	return a.config
}

func (a *API) Version() string {
	return a.version
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.ServeHTTP(w, r)
}

// NewAPI instantiates a new REST API
func NewAPI(config *conf.GlobalConfiguration, opts ...Option) *API {
	return NewApiWithVersion(defaultVersion, config, opts...)
}

// NewAPIWithVersion creates a new REST API using the specified version
func NewApiWithVersion(version string, config *conf.GlobalConfiguration, opts ...Option) *API {
	api := &API{
		config:  config,
		version: version,
	}

	for _, o := range opts {
		o.apply(api)
	}

	xffmw, _ := xff.Default()

	r := newRouter()
	r.UseBypass(xffmw.Handler)
	r.UseBypass(recoverer)

	r.Get("/health", api.HealthCheck)

	api.handler = r

	return api
}

type HealthCheckResponse struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) error {
	return sendJSON(w, http.StatusOK, HealthCheckResponse{
		Version:     a.version,
		Name:        "Go Boilerplate",
		Description: "A simple boilerplate for building REST APIs in Go",
	})
}
