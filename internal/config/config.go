package config

import (
	"log"
	"os"
	"time"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env-default:"local"`
	Postgress  PostgressConfig  `yaml:"postgress"`
	HttpServer HttpServerConfig `yaml:"http_server"`
}

type PostgressConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"password"`
	DBName   string `yaml:"dbname" env-default:"url_shortener"`
}

type HttpServerConfig struct {
	Address     string        `yaml:"address" env-default:":8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := "D:/go-url-shortener/config/local.yaml" 
	if err := os.Setenv("CONFIG_PATH", configPath); err != nil {
		log.Fatal("CONFIG_PATH is not set")
		return nil
	}


	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	return &cfg
}
