package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type UserConfig struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email"`
}

func GetUserConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	hookbridgeDir := filepath.Join(dir, "hookbridge")
	// owner: rwx, group: rx, others: rx   (only owner can write)
	err = os.MkdirAll(hookbridgeDir, 0o755) // returns nil if success, or alr exists
	if err != nil {
		return "", err
	}
	return filepath.Join(hookbridgeDir, "user-config.json"), nil
}

func SaveUserConfig(userConfig *UserConfig) error {
	path, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return err
	}

	// 0o600; read/write only, no execute; and only owner
	return os.WriteFile(path, data, 0o600)
}

func LoadUserConfig() (*UserConfig, error) {
	path, err := GetUserConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &UserConfig{}, nil // no config yet
	} else if err != nil {
		return nil, err
	}

	// use unmarshal if already have data in memory
	// use newdecoder.decode() if have an io.Reader stream
	var userConfig UserConfig
	err = json.Unmarshal(data, &userConfig)
	if err != nil {
		return nil, err
	}

	return &userConfig, nil
}

func IsLoggedIn(u *UserConfig) bool {
	return u != nil && u.AccessToken != "" && u.RefreshToken != "" && u.Email != ""
}

func DeleteUserConfig() error {
	path, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotSignedIn
		}
		return err
	}

	return nil
}
