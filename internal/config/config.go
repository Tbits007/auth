package config

import (
	"log"
	"os"
	"time"
	"github.com/ilyakaznacheev/cleanenv"
)


type Config struct {
	Env         string     `yaml:"env" env-default:"dev"`
	HTTPServer  HTTPServer `yaml:"http_server"`
	Postgres    Postgres   `yaml:"postgres"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
    Password    string        `yaml:"password" env-required:"true"`
}

type Postgres struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	DBName   string `yaml:"dbname" env-default:"postgres"`
}

func MustLoad() *Config {
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        log.Fatal("CONFIG_PATH environment variable is not set")
    }

    if _, err := os.Stat(configPath); err != nil {
        log.Fatalf("error opening config file: %s", err)
    }

    var cfg Config

    err := cleanenv.ReadConfig(configPath, &cfg)
    if err != nil {
        log.Fatalf("error reading config file: %s", err)
    }

    return &cfg
}