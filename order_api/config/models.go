package config

import "time"

func New() *Config {
	return &Config{
		Server: Server{},
		GRPC:   GRPC{},
		Kafka:  Kafka{},
		Loki:   Loki{},
		Tempo:  Tempo{},
	}
}

type Config struct {
	AppHome string `env:"APP_HOME" envDefault:""`
	Server  Server `yaml:"server"`
	GRPC    GRPC
	Kafka   Kafka `yaml:"kafka"`
	Loki    Loki
	Tempo   Tempo
}

type Server struct {
	Host              string        `env:"HOST" envDefault:"localhost"`
	Port              string        `env:"PORT" envDefault:"8080"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

type GRPC struct {
	Host string `env:"GRPC_HOST" envDefault:"localhost"`
	Port string `env:"GRPC_PORT" envDefault:"50051"`
}

type Kafka struct {
	Brokers string            `env:"KAFKA_BROKER" envDefault:"kafka:9092"`
	GroupID string            `yaml:"group_id"`
	Topics  map[string]string `yaml:"topics"`
}

type Loki struct {
	Endpoint string `env:"LOKI_ENDPOINT" envDefault:"http://loki:3100/loki/api/v1/push"`
}

type Tempo struct {
	Endpoint string `env:"TEMPO_ENDPOINT" envDefault:"tempo:4317"`
}
