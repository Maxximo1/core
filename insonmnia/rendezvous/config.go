package rendezvous

import (
	"crypto/ecdsa"
	"net"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type Config struct {
	// Listening address.
	Addr       net.Addr
	PrivateKey *ecdsa.PrivateKey
}

type config struct {
	Addr string             `yaml:"endpoint" required:"true"`
	Eth  accounts.EthConfig `yaml:"ethereum"`
}

// NewConfig loads a new Rendezvous server config from a file.
func NewConfig(path string) (*Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	privateKey, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, err
	}

	return &Config{
		Addr:       addr,
		PrivateKey: privateKey,
	}, nil
}
