package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)


type Config struct {
	Env         string     	  `yaml:"env" env-default:"dev"`
	GRPCServer  GRPCServer 	  `yaml:"grpc_server"`
	Postgres    Postgres   	  `yaml:"postgres"`
	Redis       Redis		  `yaml:"redis"`
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

type Redis struct {
	Address     string        `yaml:"address"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	projectRoot := getProjectRoot()
	envPath := filepath.Join(projectRoot, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Warning: Couldn't load .env file: %v", err)
	}

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

func getProjectRoot() string {
    paths := []string{
        ".env",
        "../.env",
        "../../.env",
        "../../../.env",
    }
    
    for _, path := range paths {
        if _, err := os.Stat(path); err == nil {
            absPath, _ := filepath.Abs(path)
            return filepath.Dir(absPath)
        }
    }
    
    return "" 
}