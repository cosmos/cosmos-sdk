package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetNodeHomeDirectoryFromHomeFlag(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"app", "--home", "/tmp/test-home"}
	got, err := GetNodeHomeDirectory("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/test-home" {
		t.Fatalf("expected /tmp/test-home, got %q", got)
	}
}

func TestGetNodeHomeDirectoryFromHomeFlagEquals(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"app", "--home=/tmp/test-home-eq"}
	got, err := GetNodeHomeDirectory("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/test-home-eq" {
		t.Fatalf("expected /tmp/test-home-eq, got %q", got)
	}
}

func TestGetNodeHomeDirectoryFromNodeHomeEnv(t *testing.T) {
	orig := os.Args
	origEnv := os.Getenv("NODE_HOME")
	origPrefix := EnvPrefix
	defer func() {
		os.Args = orig
		os.Setenv("NODE_HOME", origEnv)
		EnvPrefix = origPrefix
	}()

	os.Args = []string{"app"}
	EnvPrefix = ""
	os.Setenv("NODE_HOME", "/tmp/node-home-env")

	got, err := GetNodeHomeDirectory("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/node-home-env" {
		t.Fatalf("expected /tmp/node-home-env, got %q", got)
	}
}

func TestGetNodeHomeDirectoryFromEnvPrefix(t *testing.T) {
	orig := os.Args
	origPrefix := EnvPrefix
	defer func() {
		os.Args = orig
		EnvPrefix = origPrefix
		os.Unsetenv("MYAPP_HOME")
	}()

	os.Args = []string{"app"}
	EnvPrefix = "MYAPP"
	os.Setenv("MYAPP_HOME", "/tmp/myapp-home")

	got, err := GetNodeHomeDirectory("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/myapp-home" {
		t.Fatalf("expected /tmp/myapp-home, got %q", got)
	}
}

func TestGetNodeHomeDirectoryFallsBackToUserHome(t *testing.T) {
	orig := os.Args
	origPrefix := EnvPrefix
	origNodeHome := os.Getenv("NODE_HOME")
	defer func() {
		os.Args = orig
		EnvPrefix = origPrefix
		os.Setenv("NODE_HOME", origNodeHome)
	}()

	os.Args = []string{"app"}
	EnvPrefix = ""
	os.Unsetenv("NODE_HOME")

	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine user home dir")
	}

	got, err := GetNodeHomeDirectory("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(userHome, "myapp")
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
