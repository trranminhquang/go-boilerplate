package conf

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// GlobalConfiguration holds all the configuration that applies to all instances.
type GlobalConfiguration struct {
	API APIConfiguration

	SiteURL         string `json:"site_url" split_words:"true" default:"http://localhost:8080"`
	URIAllowListMap map[string]glob.Glob
}

type APIConfiguration struct {
	Host            string
	Port            string `envconfig:"PORT" default:"8080"`
	Endpoint        string
	RequestIDHeader string `envconfig:"REQUEST_ID_HEADER"`
	// ExternalURL        string        `json:"external_url" envconfig:"API_EXTERNAL_URL" required:"true"`
	MaxRequestDuration time.Duration `json:"max_request_duration" split_words:"true" default:"10s"`
}

func (c *APIConfiguration) Validate() error {
	// _, err := url.ParseRequestURI(c.ExternalURL)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Validate validates all of configuration.
func (c *GlobalConfiguration) Validate() error {
	validatables := []interface {
		Validate() error
	}{
		&c.API,
		// &c.DB,
		// &c.Tracing,
		// &c.Metrics,
		// &c.SMTP,
		// &c.Mailer,
		// &c.SAML,
		// &c.Security,
		// &c.Sessions,
		// &c.Hook,
		// &c.JWT.Keys,
	}

	for _, validatable := range validatables {
		if err := validatable.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// LoadFile calls godotenv.Load() when the given filename is empty ignoring any
// errors loading, otherwise it calls godotenv.Overload(filename).
//
// godotenv.Load: preserves env, ".env" path is optional
// godotenv.Overload: overrides env, "filename" path must exist
func LoadFile(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Overload(filename)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

// LoadDirectory does nothing when configDir is empty, otherwise it will attempt
// to load a list of configuration files located in configDir by using ReadDir
// to obtain a sorted list of files containing a .env suffix.
//
// When the list is empty it will do nothing, otherwise it passes the file list
// to godotenv.Overload to pull them into the current environment.
func LoadDirectory(configDir string) error {
	if configDir == "" {
		return nil
	}

	// Returns entries sorted by filename
	ents, err := os.ReadDir(configDir)
	if err != nil {
		// We mimic the behavior of LoadGlobal here, if an explicit path is
		// provided we return an error.
		return err
	}

	var paths []string
	for _, ent := range ents {
		if ent.IsDir() {
			continue // ignore directories
		}

		// We only read files ending in .env
		name := ent.Name()
		if !strings.HasSuffix(name, ".env") {
			continue
		}

		// ent.Name() does not include the watch dir.
		paths = append(paths, filepath.Join(configDir, name))
	}

	// If at least one path was found we load the configuration files in the
	// directory. We don't call override without config files because it will
	// override the env vars previously set with a ".env", if one exists.
	return loadDirectoryPaths(paths...)
}

func loadDirectoryPaths(p ...string) error {
	// If at least one path was found we load the configuration files in the
	// directory. We don't call override without config files because it will
	// override the env vars previously set with a ".env", if one exists.
	if len(p) > 0 {
		if err := godotenv.Overload(p...); err != nil {
			return err
		}
	}
	return nil
}

// LoadGlobalFromEnv will return a new *GlobalConfiguration value from the
// currently configured environment.
func LoadGlobalFromEnv() (*GlobalConfiguration, error) {
	config := new(GlobalConfiguration)
	if err := loadGlobal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func loadGlobal(config *GlobalConfiguration) error {
	if err := envconfig.Process("", config); err != nil {
		return err
	}

	// if err := config.ApplyDefaults(); err != nil {
	// 	return err
	// }

	if err := config.Validate(); err != nil {
		return err
	}
	// return populateGlobal(config)
	return nil
}
