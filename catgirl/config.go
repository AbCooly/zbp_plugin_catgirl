package catgirl

import (
	"encoding/json"
	"errors"
	"github.com/FloatTech/floatbox/file"
	"os"
)

// 配置结构体
type serverConfig struct {
	Cookie string `json:"cookie"`
	file   string
}

func newServerConfig(file string) *serverConfig {
	return &serverConfig{
		file: file,
	}
}

func (cfg *serverConfig) update(token string) (err error) {

	if token != "" {
		cfg.Cookie = token
	}
	reader, err := os.Create(cfg.file)
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewEncoder(reader).Encode(cfg)
}

func (cfg *serverConfig) load() (err error) {
	if cfg.Cookie != "" {
		return
	}
	if file.IsNotExist(cfg.file) {
		err = errors.New("no server config")
		return
	}
	reader, err := os.Open(cfg.file)
	if err != nil {
		return
	}
	defer reader.Close()
	err = json.NewDecoder(reader).Decode(cfg)
	return
}
