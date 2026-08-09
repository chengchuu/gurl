package site

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExtractAPIUsesExportedRootPackage(t *testing.T) {
	t.Parallel()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	cfg := testConfig(t)
	docs, err := ExtractAPI(root, cfg.ModulePath)
	if err != nil {
		t.Fatal(err)
	}
	if docs.PackageName != "gurl" {
		t.Fatalf("package name = %q", docs.PackageName)
	}
	want := map[string]bool{
		"GetQueryParam": false, "SetQueryParam": false, "DelQueryParam": false,
		"GetHashParam": false, "SetHashParam": false, "DelHashParam": false,
		"GetPath": false, "SetPath": false, "GetHost": false, "SetHost": false,
		"GetHostname": false, "SetHostname": false, "GetProtocol": false, "SetProtocol": false,
		"CheckValid": false, "CheckValidHTTPURL": false, "GetURLFileType": false, "GetBaseURL": false,
	}
	for _, item := range docs.Items {
		if _, ok := want[item.Name]; ok {
			want[item.Name] = true
		}
		if strings.Contains(item.Doc, "\n\turl:") {
			t.Errorf("generated API documentation for %s uses a stale parameter name", item.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("generated API missing %s", name)
		}
	}
}
