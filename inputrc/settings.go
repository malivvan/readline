package inputrc

import (
	"os"
)

type Bind struct {
	Action string
	Macro  bool
}

type Config struct {
	ReadFileFunc func(string) ([]byte, error)
	Vars         map[string]interface{}
	Binds        map[string]map[string]Bind
	Funcs        map[string]func(string, string) error
}

type ConfigOption func(*Config)

type Handler interface {
	ReadFile(name string) ([]byte, error)

	Do(typ string, param string) error

	Set(name string, value interface{}) error

	Get(name string) interface{}

	Bind(keymap, sequence, action string, macro bool) error
}

func NewConfig() *Config {
	return &Config{
		Vars:  make(map[string]interface{}),
		Binds: make(map[string]map[string]Bind),
		Funcs: make(map[string]func(string, string) error),
	}
}

func NewDefaultConfig(opts ...ConfigOption) *Config {
	cfg := &Config{
		ReadFileFunc: os.ReadFile,
		Vars:         DefaultVars(),
		Binds:        DefaultBinds(),
		Funcs:        make(map[string]func(string, string) error),
	}
	for _, o := range opts {
		o(cfg)
	}

	return cfg
}

func WithConfigReadFileFunc(readFileFunc func(string) ([]byte, error)) ConfigOption {
	return func(cfg *Config) {
		cfg.ReadFileFunc = readFileFunc
	}
}

func (cfg *Config) Bind(keymap, sequence, action string, macro bool) error {
	if cfg.Binds[keymap] == nil {
		cfg.Binds[keymap] = make(map[string]Bind)
	}

	cfg.Binds[keymap][sequence] = Bind{
		Action: action,
		Macro:  macro,
	}

	return nil
}

func (cfg *Config) Do(name, value string) error {
	if f, ok := cfg.Funcs[name]; ok {
		return f(name, value)
	}

	if f, ok := cfg.Funcs[""]; ok {
		return f(name, value)
	}

	return nil
}

func (cfg *Config) Get(name string) interface{} {
	return cfg.Vars[name]
}

func (cfg *Config) GetBool(name string) bool {
	if v, ok := cfg.Vars[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}

	return false
}

func (cfg *Config) GetInt(name string) int {
	if v, ok := cfg.Vars[name]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}

	return 0
}

func (cfg *Config) GetString(name string) string {
	if v, ok := cfg.Vars[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func (cfg *Config) ReadFile(name string) ([]byte, error) {
	if cfg.ReadFileFunc != nil {
		return cfg.ReadFileFunc(name)
	}

	return nil, os.ErrNotExist
}

func (cfg *Config) Set(name string, value interface{}) error {
	cfg.Vars[name] = value
	return nil
}
