package fileserver

import (
	"context"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"mime"
	"net/http"
	"sync"
)

const pluginName string = "fileserver"

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if config section exists.
	Has(name string) bool
}

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

type Plugin struct {
	sync.Mutex
	config *Config

	log *zap.Logger
	app *http.Server
}

func (p *Plugin) Init(cfg Configurer, log Logger) error {
	const op = errors.Op("file_server_init")

	if !cfg.Has(pluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(pluginName, &p.config)
	if err != nil {
		return errors.E(op, err)
	}

	p.log = log.NamedLogger(pluginName)

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	p.Lock()
	mux := new(http.ServeMux)

	for _, cfg := range p.config.MimeTypes {
		err := mime.AddExtensionType(cfg.Ext, cfg.MimeType)
		if err != nil {
			p.log.Error(
				"failed to register mime type",
				zap.String("ext", cfg.Ext),
				zap.String("mime-type", cfg.MimeType),
			)
		}
	}

	for _, cfg := range p.config.VirtualHosts {
		fs := http.FileServer(http.Dir(cfg.Root))
		mux.Handle(cfg.Prefix, http.StripPrefix(cfg.Prefix, fs))
	}

	p.app = &http.Server{
		ReadTimeout:  p.config.ReadTimeout,
		WriteTimeout: p.config.WriteTimeout,
		IdleTimeout:  p.config.IdleTimeout,
		Addr:         p.config.Address,
		Handler:      mux,
	}

	go func() {
		p.Unlock()
		p.log.Info("file server started", zap.String("address", p.config.Address))
		err := p.app.ListenAndServe()
		if err != nil {
			errCh <- err
			return
		}
	}()

	return errCh
}

func (p *Plugin) Stop(ctx context.Context) error {
	endCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func() {
		p.Lock()
		defer p.Unlock()

		err := p.app.Shutdown(ctx)
		if err != nil {
			errCh <- err
			return
		}

		endCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-errCh:
		return e
	case <-endCh:
		return nil
	}
}

func (p *Plugin) Name() string {
	return pluginName
}
