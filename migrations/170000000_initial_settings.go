package migrations

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

type Config struct {
	AppName string `env:"APP_NAME" envDefault:"Linkz"`
	AppURL  string `env:"APP_URL" envDefault:"http://localhost:8090"`
}

func init() {
	// Parse environment variables
	cfg, err := env.ParseAs[Config]()

	if err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	m.Register(func(app core.App) error {
		settings := app.Settings()

		// for all available settings fields you could check
		// https://github.com/pocketbase/pocketbase/blob/develop/core/settings_model.go#L121-L130
		settings.Meta.AppName = "Linkz"
		settings.Meta.AppURL = cfg.AppURL
		settings.Logs.MaxDays = 2

		return app.Save(settings)
	}, nil)
}
