package service

import "sync"

type ConfigurationWithConstructor[T comparable] interface {
	Build() (T, error)
}

type ExternalService[S comparable, CFG ConfigurationWithConstructor[S]] struct {
	mu      sync.Mutex
	service S
	cfg     CFG
}

func (cfg *ExternalService[S, CFG]) GetService() (S, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	var zero S
	if cfg.service == zero {
		service, err := cfg.cfg.Build()
		if err != nil {
			return zero, err
		}
		cfg.service = service
	}

	return cfg.service, nil
}

func FromImplementation[S comparable, CFG ConfigurationWithConstructor[S]](impl S) *ExternalService[S, CFG] {
	return &ExternalService[S, CFG]{
		service: impl,
	}
}

func (cfg *ExternalService[S, CFG]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&cfg.cfg); err != nil {
		return err
	}

	return nil
}
