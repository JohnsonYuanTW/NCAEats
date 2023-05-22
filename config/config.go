package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ChannelSecret      string        `envconfig:"CHANNEL_SECRET"`
	ChannelAccessToken string        `envconfig:"CHANNEL_ACCESS_TOKEN"`
	SSLCertfilePath    string        `envconfig:"SSL_CERTIFICATE_FILE"`
	SSLKeyPath         string        `envconfig:"SSL_KEY_PATH"`
	SiteURL            string        `envconfig:"SITE_URL"`
	Port               string        `envconfig:"PORT"`
	DBUsername         string        `envconfig:"DB_USERNAME"`
	DBPassword         string        `envconfig:"DB_PASSWORD"`
	DBURL              string        `envconfig:"DB_URL"`
	DBName             string        `envconfig:"DB_NAME"`
	DBPort             string        `envconfig:"DB_PORT"`
	DBMaxIdleConns     int           `envconfig:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns     int           `envconfig:"DB_MAX_OPEN_CONNS"`
	DBConnMaxLifetime  time.Duration `envconfig:"DB_CONN_MAX_LIFETIME"`
}

func LoadEnvVariables() (*Config, error) {
	s := &Config{}
	if err := envconfig.Process("", s); err != nil {
		return nil, err
	}
	return s, nil
}
