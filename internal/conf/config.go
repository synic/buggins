package conf

import (
	"errors"
	"os"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ModuleOptions map[string]any `yaml:"modules"`
	DiscordToken  string         `yaml:"token"`
	DatabaseURL   string         `yaml:"database_url"`
}

func getDefaults() Config {
	return Config{
		DatabaseURL: "db.sqlite",
	}
}

func ProvideFromFile(location string) func() (Config, error) {
	return func() (Config, error) {
		return New(location)
	}
}

func New(file string) (Config, error) {
	conf := getDefaults()
	data, err := os.ReadFile(file)

	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		return conf, err
	}

	err = validate(conf)

	if err != nil {
		return conf, err
	}

	return conf, nil
}

func validate(conf Config) error {
	if conf.DiscordToken == "" {
		return errors.New("discord token not provided")
	}

	if len(conf.ModuleOptions) == 0 {
		return errors.New("no modules configured")
	}

	return nil
}

func (c Config) Populate(module string, res any) error {
	data, ok := c.ModuleOptions[module]

	if !ok {
		return errors.New("module options not found")
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           res,
	})

	if err != nil {
		return err
	}

	err = decoder.Decode(data)

	if err != nil {
		return err
	}

	return nil
}
