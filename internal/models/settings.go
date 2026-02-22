package models

import (
	"database/sql"
)

type SettingsStore struct {
	DB *sql.DB
}

func (s *SettingsStore) Get(key string) (string, error) {
	var val string
	err := s.DB.QueryRow("SELECT setting_value FROM settings WHERE setting_key = ?", key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *SettingsStore) Set(key, value string) error {
	_, err := s.DB.Exec("INSERT INTO settings (setting_key, setting_value) VALUES (?, ?) ON DUPLICATE KEY UPDATE setting_value = ?", key, value, value)
	return err
}

func (s *SettingsStore) GetAll() (map[string]string, error) {
	rows, err := s.DB.Query("SELECT setting_key, setting_value FROM settings ORDER BY setting_key")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, rows.Err()
}
