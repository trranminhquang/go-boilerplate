package db

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/trranminhquang/go-boilerplate/internal/conf"
)

// Connection is the interface a storage provider must implement.
type Connection struct {
	*pop.Connection
}

// Dial will connect to that storage engine
func Dial(config *conf.GlobalConfiguration) (*Connection, error) {
	if config.DB.Driver == "" && config.DB.URL != "" {
		u, err := url.Parse(config.DB.URL)
		if err != nil {
			return nil, errors.Wrap(err, "parsing db connection url")
		}
		config.DB.Driver = u.Scheme
	}

	driver := ""
	if config.DB.Driver != "postgres" {
		logrus.Warn("DEPRECATION NOTICE: only PostgreSQL is supported by Supabase's GoTrue, will be removed soon")
	} else {
		// pop v5 uses pgx as the default PostgreSQL driver
		driver = "pgx"
	}

	// if driver != "" && (config.Tracing.Enabled || config.Metrics.Enabled) {
	// 	instrumentedDriver, err := otelsql.Register(driver)
	// 	if err != nil {
	// 		logrus.WithError(err).Errorf("unable to instrument sql driver %q for use with OpenTelemetry", driver)
	// 	} else {
	// 		logrus.Debugf("using %s as an instrumented driver for OpenTelemetry", instrumentedDriver)
	// 		// sqlx needs to be informed that the new instrumented
	// 		// driver has the same semantics as the
	// 		// non-instrumented driver
	// 		sqlx.BindDriver(instrumentedDriver, sqlx.BindType(driver))

	// 		driver = instrumentedDriver
	// 	}
	// }

	options := make(map[string]string)

	if config.DB.HealthCheckPeriod != time.Duration(0) {
		options["pool_health_check_period"] = config.DB.HealthCheckPeriod.String()
	}

	if config.DB.ConnMaxIdleTime != time.Duration(0) {
		options["pool_max_conn_idle_time"] = config.DB.ConnMaxIdleTime.String()
	}

	db, err := pop.NewConnection(&pop.ConnectionDetails{
		Dialect:         config.DB.Driver,
		Driver:          driver,
		URL:             config.DB.URL,
		Pool:            config.DB.MaxPoolSize,
		IdlePool:        config.DB.MaxIdlePoolSize,
		ConnMaxLifetime: config.DB.ConnMaxLifetime,
		ConnMaxIdleTime: config.DB.ConnMaxIdleTime,
		Options:         options,
	})
	if err != nil {
		return nil, errors.Wrap(err, "opening database connection")
	}
	if err := db.Open(); err != nil {
		return nil, errors.Wrap(err, "checking database connection")
	}

	// if config.Metrics.Enabled {
	// 	registerOpenTelemetryDatabaseStats(db)
	// }

	return &Connection{db}, nil
}

type CommitWithError struct {
	Err error
}

func (e *CommitWithError) Error() string {
	return e.Err.Error()
}

func (e *CommitWithError) Cause() error {
	return e.Err
}

// NewCommitWithError creates an error that can be returned in a pop transaction
// without rolling back the transaction. This should only be used in cases where
// you want the transaction to commit but return an error message to the user.
func NewCommitWithError(err error) *CommitWithError {
	return &CommitWithError{Err: err}
}

func (c *Connection) Transaction(fn func(*Connection) error) error {
	if c.TX == nil {
		var returnErr error
		if terr := c.Connection.Transaction(func(tx *pop.Connection) error {
			err := fn(&Connection{tx})
			switch err.(type) {
			case *CommitWithError:
				returnErr = err
				return nil
			default:
				return err
			}
		}); terr != nil {
			// there exists a race condition when the context deadline is exceeded
			// and whether the transaction has been committed or not
			// e.g. if the context deadline has exceeded but the transaction has already been committed,
			// it won't be possible to perform a rollback on the transaction since the transaction has been closed
			if !errors.Is(terr, sql.ErrTxDone) {
				return terr
			}
		}
		return returnErr
	}
	return fn(c)
}

// WithContext returns a new connection with an updated context. This is
// typically used for tracing as the context contains trace span information.
func (c *Connection) WithContext(ctx context.Context) *Connection {
	return &Connection{c.Connection.WithContext(ctx)}
}
