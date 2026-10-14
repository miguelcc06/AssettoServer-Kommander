package api

import "testing"

func TestHostProjectDirDefault(t *testing.T) {
	t.Setenv("HOST_PROJECT_DIR", "")
	if got := hostProjectDir(); got != "/app/host" {
		t.Fatalf("hostProjectDir() = %q, want /app/host", got)
	}
}

func TestHostProjectDirFromEnv(t *testing.T) {
	t.Setenv("HOST_PROJECT_DIR", "/srv/AssettoServer-Kommander")
	if got := hostProjectDir(); got != "/srv/AssettoServer-Kommander" {
		t.Fatalf("hostProjectDir() = %q", got)
	}
}

func TestComposeBaseArgsUsesHostProjectDir(t *testing.T) {
	t.Setenv("HOST_PROJECT_DIR", "/opt/kommander")
	args := composeBaseArgs()
	if len(args) < 2 || args[0] != "--project-directory" || args[1] != "/opt/kommander" {
		t.Fatalf("composeBaseArgs() = %#v", args)
	}
}
