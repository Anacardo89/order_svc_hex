package config

import "time"

func New() *Config {
	return &Config{
		DB:    DB{},
		Kafka: Kafka{},
	}
}

type Config struct {
	AppHome string `env:"APP_HOME" envDefault:""`
	DB      DB     `yaml:"db"`
	Kafka   Kafka  `yaml:"kafka"`
}

type DB struct {
	DSN             string        `env:"DB_DSN" envDefault:"postgres://user:pass@db:5432/dbname?sslmode=disable"`
	MaxConns        int           `yaml:"max_conns"`
	MinConns        int           `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`  // minutes
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"` // minutes
}

type Kafka struct {
	Brokers []string         `yaml:"brokers"`
	GroupID string           `yaml:"group_id"`
	Topics  map[string]Topic `yaml:"topics"`
}

type Topic struct {
	Name string `yaml:"name"`
	DLQ  string `yaml:"dlq"`
}
