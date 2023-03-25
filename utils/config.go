package loggerInjector

import (
	"fmt"
	"os"
)

type Config struct {
	Elastic                Elastic
	FluentdImageRepository string
}

func NewConfig() (*Config, error) {
	config := &Config{
		Elastic: Elastic{
			Host:       os.Getenv(ElasticHostEnvKey),
			Port:       ConvertToIntOrDefault(os.Getenv(ElasticPortEnvKey)),
			Password:   os.Getenv(ElasticPasswordEnvKey),
			User:       os.Getenv(ElasticUserEnvKey),
			SslVerify:  ConvertToBooleanOrDefault(os.Getenv(ElasticSslVerifyEnvKey)),
			Scheme:     os.Getenv(ElasticSchemaEnvKey),
			SslVersion: os.Getenv(ElasticSslVersionEnvKey),
		},
		FluentdImageRepository: os.Getenv(FluentdImageRepositoryEnvKey),
	}

	if !config.EnsureRequirements() {
		return config, fmt.Errorf("missing configurations , please import the required envrionment variables")
	}
	return config, nil
}
func (c *Config) EnsureRequirements() bool {
	return c.Elastic.Host != "" &&
		c.Elastic.Port > 0
}
