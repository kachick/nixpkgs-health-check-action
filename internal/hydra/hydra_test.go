package hydra

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseEvalStreamSnapshots(t *testing.T) {
	dirs := []string{
		filepath.Join("testdata", "fixtures"),
		filepath.Join("testdata", "test-snapshots"),
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir(%s): %v", dir, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".expected.json") {
				continue
			}

			t.Run(filepath.Join(filepath.Base(dir), name), func(t *testing.T) {
				rawBytes, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				expectedPath := filepath.Join(dir, strings.TrimSuffix(name, ".json")+".expected.json")
				expectedBytes, err := os.ReadFile(expectedPath)
				if err != nil {
					t.Fatal(err)
				}

				got, err := ParseEvalStream(bytes.NewReader(rawBytes), nil)
				if err != nil {
					t.Fatalf("ParseEvalStream: %v", err)
				}

				// Existing jq expected files contain one JSON array per top-level input object.
				want := decodeExpectedStream(t, expectedBytes)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("mismatch:\ngot:  %+v\nwant: %+v", got, want)
				}
			})
		}
	}
}

func decodeExpectedStream(t *testing.T, data []byte) []BuildResult {
	t.Helper()
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	var all []BuildResult
	for {
		var batch []BuildResult
		err := json.UnmarshalDecode(dec, &batch)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("decode expected json stream: %v", err)
		}
		all = append(all, batch...)
	}
	if all == nil {
		all = []BuildResult{}
	}
	return all
}

func TestCheckIgnoreAndStaleRules(t *testing.T) {
	ignoredURL := "https://hydra.nixos.org/build/201"
	otherURL := "https://hydra.nixos.org/build/202"
	staleURL := "https://hydra.nixos.org/build/999"

	sample := `{
		"1/foo": {
			"now_fail": [
				{"name": "foo-1.0", "job_name": "foo.x86_64-linux", "arch": "x86_64-linux", "status": "Failed", "build_url": "` + ignoredURL + `"},
				{"name": "foo-1.0", "job_name": "foo.aarch64-linux", "arch": "aarch64-linux", "status": "Failed", "build_url": "` + otherURL + `"}
			]
		}
	}`

	run := func(_ context.Context, _ string) ([]byte, error) {
		return []byte(sample), nil
	}

	res, err := Check(context.Background(), run, "foo", []string{ignoredURL, staleURL}, []string{ignoredURL, staleURL})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(res.Warnings) != 1 || res.Warnings[0].URL == nil || *res.Warnings[0].URL != ignoredURL {
		t.Errorf("unexpected warnings: %+v", res.Warnings)
	}
	if len(res.Errors) != 1 || res.Errors[0].URL == nil || *res.Errors[0].URL != otherURL {
		t.Errorf("unexpected errors: %+v", res.Errors)
	}
	if len(res.StaleIgnore) != 1 || res.StaleIgnore[0] != staleURL {
		t.Errorf("unexpected stale ignore: %+v", res.StaleIgnore)
	}
}
