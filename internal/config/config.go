package config

import (
	"log"
	"os"
	"time"
	"github.com/ilyakaznacheev/cleanenv"
)


type Config struct {
	Env         string     	  `yaml:"env" env-default:"dev"`
	GRPCServer  GRPCServer 	  `yaml:"grpc_server"`
	Postgres    Postgres   	  `yaml:"postgres"`
	Auth	 	Auth 		  `yaml:"auth"`	
}

type Auth struct {
	TokenTTL 	time.Duration `yaml:"tokenTTL"`	
	SecretKey 	string		  `yaml:"secretKey"`
}

type GRPCServer struct {  
    Port    int           `yaml:"port"`  
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