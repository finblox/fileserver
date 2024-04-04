package fileserver

import (
	"github.com/roadrunner-server/errors"
	"time"
)

type Config struct {
	// Address to serve
	Address string `mapstructure:"address"`

	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`

	MimeTypes []*MimeTypeCfg `mapstructure:"mime_types"`

	// per-root configuration
	VirtualHosts []*VirtualHostCfg `mapstructure:"serve"`
}

type MimeTypeCfg struct {
	// Ext HTTP
	Ext string `mapstructure:"ext"`

	// MimeType defines the mime type of the corresponding extension
	MimeType string `mapstructure:"mime_type"`
}

type VirtualHostCfg struct {
	// Prefix HTTP
	Prefix string `mapstructure:"prefix"`

	// Dir contains name of directory to control access to.
	// Default - "."
	Root string `mapstructure:"root"`
}

func (c *Config) Valid() error {
	const op = errors.Op("static_validation")
	if c.Address == "" {
		return errors.E(op, errors.Str("empty address"))
	}

	if c.VirtualHosts == nil {
		return errors.E(op, errors.Str("no configuration to serve"))
	}

	for i := 0; i < len(c.VirtualHosts); i++ {
		if c.VirtualHosts[i].Prefix == "" {
			return errors.E(op, errors.Str("empty prefix"))
		}

		if c.VirtualHosts[i].Root == "" {
			c.VirtualHosts[i].Root = "."
		}
	}

	return nil
}
