package config

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Global   GlobalConfig             `toml:"global"`
	Packages map[string]PackageConfig `toml:"packages"`
}

type GlobalConfig struct {
	Hydra         ServiceConfig `toml:"hydra"`
	NixpkgsUpdate ServiceConfig `toml:"nixpkgs-update"`
}

type PackageConfig struct {
	Hydra         ServiceConfig `toml:"hydra"`
	NixpkgsUpdate ServiceConfig `toml:"nixpkgs-update"`
}

type ServiceConfig struct {
	Skip   SkipValue    `toml:"skip"`
	Ignore []IgnoreRule `toml:"ignore"`
}

// SkipValue supports both boolean (skip = true) and string (skip = "reason") in TOML.
type SkipValue struct {
	Skip   bool
	Reason string
}

func (s *SkipValue) UnmarshalTOML(v any) error {
	switch val := v.(type) {
	case bool:
		s.Skip = val
		if val {
			s.Reason = "Skipped"
		}
		return nil
	case string:
		if val != "" {
			s.Skip = true
			s.Reason = val
		}
		return nil
	default:
		return fmt.Errorf("skip must be a boolean or string, got %T", v)
	}
}

type IgnoreRule struct {
	URL     string `toml:"url,omitempty"`
	Pattern string `toml:"pattern,omitempty"`
	Reason  string `toml:"reason,omitempty"`
}

type UpdateCheckerConfig struct {
	Ignore []IgnoreRule `toml:"ignore"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat config file %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("config file %q is not a regular file", path)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode config file %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	for i, rule := range c.Global.Hydra.Ignore {
		if rule.URL == "" {
			return fmt.Errorf("global.hydra.ignore[%d] requires a non-empty string url", i)
		}
	}
	for pname, pkg := range c.Packages {
		for i, rule := range pkg.Hydra.Ignore {
			if rule.URL == "" {
				return fmt.Errorf("packages.%s.hydra.ignore[%d] requires a non-empty string url", pname, i)
			}
		}
	}
	return nil
}

func (c *Config) SkipHydra(pname string) (bool, string) {
	if c == nil {
		return false, ""
	}
	s := c.Packages[pname].Hydra.Skip
	return s.Skip, s.Reason
}

func (c *Config) SkipUpdate(pname string) (bool, string) {
	if c == nil {
		return false, ""
	}
	s := c.Packages[pname].NixpkgsUpdate.Skip
	return s.Skip, s.Reason
}

func (c *Config) PackageHydraIgnoreURLs(pname string) []string {
	if c == nil {
		return nil
	}
	rules := c.Packages[pname].Hydra.Ignore
	urls := make([]string, 0, len(rules))
	for _, r := range rules {
		urls = append(urls, r.URL)
	}
	return urls
}

func (c *Config) AllHydraIgnoreURLs(pname string) []string {
	if c == nil {
		return nil
	}
	urls := make([]string, 0, len(c.Global.Hydra.Ignore)+len(c.Packages[pname].Hydra.Ignore))
	for _, r := range c.Global.Hydra.Ignore {
		urls = append(urls, r.URL)
	}
	for _, r := range c.Packages[pname].Hydra.Ignore {
		urls = append(urls, r.URL)
	}
	slices.Sort(urls)
	return slices.Compact(urls)
}

func (c *Config) HasPackageUpdateIgnore(pname string) bool {
	if c == nil {
		return false
	}
	return len(c.Packages[pname].NixpkgsUpdate.Ignore) > 0
}

func (c *Config) EncodeUpdateConfig(pname string) ([]byte, error) {
	var rules []IgnoreRule
	if c != nil {
		rules = append(rules, c.Global.NixpkgsUpdate.Ignore...)
		rules = append(rules, c.Packages[pname].NixpkgsUpdate.Ignore...)
	}
	if rules == nil {
		rules = []IgnoreRule{}
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(UpdateCheckerConfig{Ignore: rules}); err != nil {
		return nil, fmt.Errorf("encode nixpkgs-update config for %q: %w", pname, err)
	}
	return buf.Bytes(), nil
}
