package todo

// Config represents the app config
type Config struct {
	AppPort string `envconfig:"APP_PORT"`

	DBHost     string `envconfig:"DB_HOST"`
	DBPort     string `envconfig:"DB_PORT"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBName     string `envconfig:"DB_NAME"`
}
