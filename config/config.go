package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port string `envconfig:"LISTEN_PORT" default:"8080"`

	ServerTimeout time.Duration `envconfig:"SERVER_TIMEOUT" default:"3m"`

	JWTSecretKey       []byte        `envconfig:"JWT_SECRET_KEY" default:"32Thdl3XHtanj3rKsn0lkS38HbMUh1p8ZLZRy5w3MS0="`
	JWTTokenExpiration time.Duration `envconfig:"JWT_TOKEN_EXPIRATION" default:"24h"`

	SMTPHost     string `envconfig:"SMTP_HOST" default:"smtp-relay.sendinblue.com"`
	SMTPPort     string `envconfig:"SMTP_PORT" default:"587"`
	SMTPLogin    string `envconfig:"SMTP_LOGIN" default:""`
	SMTPPassword string `envconfig:"SMTP_PASSWORD" default:""`

	Email_CompanyName string `envconfig:"EMAIL_COMPANY_NAME" default:"Sonamusica Music Studio"`
	Email_Sender      string `envconfig:"EMAIL_SENDER" default:"no-reply@sonamusicaid.com"`
	Email_BaseAppURL  string `envconfig:"EMAIL_BASE_APP_URL" default:"http://localhost:8080/"`

	LogoURL string `envconfig:"LOGO_URL" default:""`

	DBHost              string `envconfig:"DB_HOST" default:"127.0.0.1"`
	DBPort              string `envconfig:"DB_PORT" default:"3306"`
	DBName              string `envconfig:"DB_NAME" default:"sonamusica-backend"`
	DBUser              string `envconfig:"DB_USER" default:"backend-user"`
	DBPassword          string `envconfig:"DB_PASSWORD" default:"p4ssw0rd"`
	DBMaxOpenConnection int    `envconfig:"DB_MAX_OPEN_CONNECTION" default:"3"`
}

var doOnce sync.Once

func Get() Config {
	doOnce.Do(func() {
		LoadEnvFile()
	})

	conf := Config{}
	envconfig.MustProcess("", &conf)

	return conf
}

func LoadEnvFile() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env
}

func (c *Config) GetMySQLURI() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
