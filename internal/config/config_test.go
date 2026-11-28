package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestLoadRepoConfig(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "nixpkgs-health-check-by-maintainer.toml"))
	if err != nil {
		t.Fatalf("Load repo config: %v", err)
	}

	if skip, reason := cfg.SkipHydra("hello"); !skip || reason != "Test skip" {
		t.Errorf("SkipHydra(hello) = (%v, %q), want (true, %q)", skip, reason, "Test skip")
	}
	if skip, _ := cfg.SkipUpdate("hello"); !skip {
		t.Errorf("SkipUpdate(hello) = false, want true")
	}
	if skip, _ := cfg.SkipHydra("fishnet"); !skip {
		t.Errorf("SkipHydra(fishnet) = false, want true")
	}
	if skip, _ := cfg.SkipUpdate("fishnet"); skip {
		t.Errorf("SkipUpdate(fishnet) = true, want false")
	}
	if skip, _ := cfg.SkipUpdate("fresh-editor"); skip {
		t.Errorf("SkipUpdate(fresh-editor) = true, want false")
	}
	if !cfg.HasPackageUpdateIgnore("fresh-editor") {
		t.Errorf("HasPackageUpdateIgnore(fresh-editor) = false, want true")
	}
	if cfg.HasPackageUpdateIgnore("biz-ud-gothic") {
		t.Errorf("HasPackageUpdateIgnore(biz-ud-gothic) = true, want false")
	}

	encoded, err := cfg.EncodeUpdateConfig("fresh-editor")
	if err != nil {
		t.Fatalf("EncodeUpdateConfig(fresh-editor): %v", err)
	}
	var decoded UpdateCheckerConfig
	if err := toml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal encoded update config: %v", err)
	}
	if len(decoded.Ignore) != 2 {
		t.Fatalf("got %d ignore rules, want 2: %+v", len(decoded.Ignore), decoded.Ignore)
	}
	if !strings.Contains(decoded.Ignore[0].Pattern, `go: go\.mod requires go >=`) {
		t.Errorf("encoded update config missing global rule: %+v", decoded.Ignore[0])
	}
	if !strings.Contains(decoded.Ignore[1].URL, "fresh-editor/2026-11-06.log") {
		t.Errorf("encoded update config missing package rule: %+v", decoded.Ignore[1])
	}
}

func TestSkipValueAndValidation(t *testing.T) {
	dir := t.TempDir()

	validPath := filepath.Join(dir, "valid.toml")
	validTOML := `
[[global.hydra.ignore]]
url = "https://hydra.nixos.org/build/1"
reason = "global"

[packages.foo.hydra]
skip = true

[[packages.bar.hydra.ignore]]
url = "https://hydra.nixos.org/build/2"
reason = "pkg"

[[packages.bar.hydra.ignore]]
url = "https://hydra.nixos.org/build/1"
reason = "dup"
`
	if err := os.WriteFile(validPath, []byte(validTOML), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(validPath)
	if err != nil {
		t.Fatalf("Load valid config: %v", err)
	}

	if skip, reason := cfg.SkipHydra("foo"); !skip || reason != "Skipped" {
		t.Errorf("SkipHydra(foo) = (%v, %q), want (true, Skipped)", skip, reason)
	}

	gotAll := cfg.AllHydraIgnoreURLs("bar")
	wantAll := []string{"https://hydra.nixos.org/build/1", "https://hydra.nixos.org/build/2"}
	if !slices.Equal(gotAll, wantAll) {
		t.Errorf("AllHydraIgnoreURLs(bar) = %v, want %v", gotAll, wantAll)
	}

	invalidPath := filepath.Join(dir, "invalid.toml")
	invalidTOML := `
[[packages.bar.hydra.ignore]]
reason = "missing url"
`
	if err := os.WriteFile(invalidPath, []byte(invalidTOML), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(invalidPath); err == nil {
		t.Fatal("expected error for hydra ignore rule without url")
	}
}
