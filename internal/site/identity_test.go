package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSourceIdentity(t *testing.T) {
	t.Parallel()
	root := testProjectRoot(t)
	cfg := testConfig(t)
	if err := ValidateSourceIdentity(root, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSourceIdentityRejectsOwnerDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source, err := os.ReadFile(filepath.Join(testProjectRoot(t), "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.ReplaceAll(string(source), "chengchuu", "different-owner")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSourceIdentity(root, testConfig(t)); err == nil {
		t.Fatal("ValidateSourceIdentity accepted a drifted GitHub owner")
	}
}
