package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string  `yaml:"env" env-required:"true"`
	Binance Binance `yaml:"binance" env-required:"true"`
}

type Binance struct {
	Depth BinanceDepth
}

type BinanceDepth struct {
	Symbols []string `yaml:"symbols"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if path == "" {
		panic("config is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config" + err.Error())
	}

	return &config
}

/*
Take path to config
Priority: flag > env > default
default - empty string
*/
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path_to_config_file")

	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res

}
