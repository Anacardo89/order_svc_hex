package config

import "time"

func New() *Config {
	return &Config{
		Server: Server{},
		DB:     DB{},
		Kafka:  Kafka{},
	}
}

type Config struct {
	AppHome string `env:"APP_HOME" envDefault:""`
	Server  Server
	DB      DB    `yaml:"db"`
	Kafka   Kafka `yaml:"kafka"`
}

type Server struct {
	Host string `env:"HOST" envDefault:"localhost"`
	Port string `env:"PORT" envDefault:"8080"`
}

type DB struct {
	DSN             string        `env:"DB_DSN" envDefault:"postgres://user:pass@db:5432/dbname?sslmode=disable"`
	MaxConns        int           `yaml:"max_conns"`
	MinConns        int           `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`  // minutes
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"` // minutes
}

type Kafka struct {
	Brokers        string           `env:"KAFKA_BROKER" envDefault:"kafka:9092"`
	GroupID        string           `yaml:"group_id"`
	Topics         map[string]Topic `yaml:"topics"`
	WorkerPoolSize int              `yaml:"worker_pool_size"`
	BatchSize      int              `yaml:"batch_size"`
	BatchTimeout   time.Duration    `yaml:"batch_timeout"`
	QueueSize      int              `yaml:"queue_size"`
}

type Topic struct {
	Name string `yaml:"name"`
	DLQ  string `yaml:"dlq"`
}
