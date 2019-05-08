package main

import (
	"context"
	"net/http"
	"time"

	"github.com/etherlabsio/healthcheck"
)

func healthcheckHandler() http.Handler {
	return healthcheck.Handler(

		// WithTimeout allows you to set a max overall timeout.
		healthcheck.WithTimeout(5*time.Second),

		healthcheck.WithChecker(
			"database", healthcheck.CheckerFunc(
				func(ctx context.Context) error {
					_, err := db.ExecOneContext(ctx, "SELECT 'healthcheck check'")
					return err
				},
			),
		),
	)
}
