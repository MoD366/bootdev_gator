package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Dburl       string `json:"db_url"`
	CurrentUser string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	fullpath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	cfg, err := os.ReadFile(fullpath)
	if err != nil {
		return Config{}, err
	}

	var conf Config
	if err = json.Unmarshal(cfg, &conf); err != nil {
		return Config{}, err
	}

	return conf, nil
}

func (conf Config) SetUser(name string) error {
	conf.CurrentUser = name

	jsonData, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	fullpath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(fullpath, jsonData, 0644)

	return err
}

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return homedir + "/" + configFileName, nil
}
