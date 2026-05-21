package config

import (
	"encoding/json"
	"os"

	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

// SaveToken 保存token到本地
func (m *Manager) SaveToken(token *bdisk.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.TokenPath(), data, 0600)
}

// LoadToken 从本地加载token
func (m *Manager) LoadToken() (*bdisk.Token, error) {
	data, err := os.ReadFile(m.TokenPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, bdisk.ErrNoToken
		}
		return nil, err
	}

	var token bdisk.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// ClearToken 清除本地token
func (m *Manager) ClearToken() error {
	return os.Remove(m.TokenPath())
}
