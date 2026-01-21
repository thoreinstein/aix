package opencode

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMCPManager_List(t *testing.T) {
	t.Run("empty config returns empty slice", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		servers, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if servers == nil {
			t.Error("List() returned nil, want empty slice")
		}
		if len(servers) != 0 {
			t.Errorf("List() returned %d servers, want 0", len(servers))
		}
	})

	t.Run("returns servers from config", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		// Add test servers
		server1 := &MCPServer{Name: "server1", Command: []string{"cmd1"}}
		server2 := &MCPServer{Name: "server2", Command: []string{"cmd2"}}

		if err := mgr.Add(server1); err != nil {
			t.Fatalf("Add(server1) error = %v", err)
		}
		if err := mgr.Add(server2); err != nil {
			t.Fatalf("Add(server2) error = %v", err)
		}

		servers, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(servers) != 2 {
			t.Errorf("List() returned %d servers, want 2", len(servers))
		}
	})

	t.Run("returns servers sorted by name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		// Add in non-alphabetical order
		if err := mgr.Add(&MCPServer{Name: "zebra", Command: []string{"z"}}); err != nil {
			t.Fatalf("Add(zebra) error = %v", err)
		}
		if err := mgr.Add(&MCPServer{Name: "alpha", Command: []string{"a"}}); err != nil {
			t.Fatalf("Add(alpha) error = %v", err)
		}
		if err := mgr.Add(&MCPServer{Name: "middle", Command: []string{"m"}}); err != nil {
			t.Fatalf("Add(middle) error = %v", err)
		}

		servers, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(servers) != 3 {
			t.Fatalf("List() returned %d servers, want 3", len(servers))
		}
		if servers[0].Name != "alpha" {
			t.Errorf("servers[0].Name = %q, want %q", servers[0].Name, "alpha")
		}
		if servers[1].Name != "middle" {
			t.Errorf("servers[1].Name = %q, want %q", servers[1].Name, "middle")
		}
		if servers[2].Name != "zebra" {
			t.Errorf("servers[2].Name = %q, want %q", servers[2].Name, "zebra")
		}
	})
}

func TestMCPManager_Get(t *testing.T) {
	t.Run("returns server by name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		server := &MCPServer{
			Name:        "test-server",
			Command:     []string{"npx", "server"},
			Type:        "local",
			Environment: map[string]string{"FOO": "bar"},
		}
		if err := mgr.Add(server); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		got, err := mgr.Get("test-server")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != server.Name {
			t.Errorf("Get().Name = %q, want %q", got.Name, server.Name)
		}
		if len(got.Command) != len(server.Command) {
			t.Errorf("Get().Command len = %d, want %d", len(got.Command), len(server.Command))
		}
		if got.Type != server.Type {
			t.Errorf("Get().Type = %q, want %q", got.Type, server.Type)
		}
	})

	t.Run("returns ErrMCPServerNotFound for nonexistent server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		_, err := mgr.Get("nonexistent")
		if !errors.Is(err, ErrMCPServerNotFound) {
			t.Errorf("Get() error = %v, want %v", err, ErrMCPServerNotFound)
		}
	})
}

func TestMCPManager_Add(t *testing.T) {
	t.Run("adds new server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		server := &MCPServer{
			Name:    "new-server",
			Command: []string{"run-server"},
		}

		if err := mgr.Add(server); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		got, err := mgr.Get("new-server")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != server.Name {
			t.Errorf("Get().Name = %q, want %q", got.Name, server.Name)
		}
	})

	t.Run("updates existing server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		original := &MCPServer{
			Name:    "update-test",
			Command: []string{"original"},
		}
		if err := mgr.Add(original); err != nil {
			t.Fatalf("Add(original) error = %v", err)
		}

		updated := &MCPServer{
			Name:    "update-test",
			Command: []string{"updated", "with", "args"},
		}
		if err := mgr.Add(updated); err != nil {
			t.Fatalf("Add(updated) error = %v", err)
		}

		got, err := mgr.Get("update-test")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if len(got.Command) != 3 {
			t.Errorf("Add() did not update: Command len = %d, want 3", len(got.Command))
		}
	})

	t.Run("returns ErrInvalidMCPServer for nil server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		err := mgr.Add(nil)
		if !errors.Is(err, ErrInvalidMCPServer) {
			t.Errorf("Add(nil) error = %v, want %v", err, ErrInvalidMCPServer)
		}
	})

	t.Run("returns ErrInvalidMCPServer for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		err := mgr.Add(&MCPServer{Name: ""})
		if !errors.Is(err, ErrInvalidMCPServer) {
			t.Errorf("Add(empty name) error = %v, want %v", err, ErrInvalidMCPServer)
		}
	})
}

func TestMCPManager_Remove(t *testing.T) {
	t.Run("removes existing server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		server := &MCPServer{Name: "to-remove", Command: []string{"cmd"}}
		if err := mgr.Add(server); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		if err := mgr.Remove("to-remove"); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}

		_, err := mgr.Get("to-remove")
		if !errors.Is(err, ErrMCPServerNotFound) {
			t.Errorf("Get() after Remove() error = %v, want %v", err, ErrMCPServerNotFound)
		}
	})

	t.Run("idempotent - no error if not exists", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		err := mgr.Remove("never-existed")
		if err != nil {
			t.Errorf("Remove(nonexistent) error = %v, want nil", err)
		}
	})
}

func TestMCPManager_Enable(t *testing.T) {
	t.Run("enables disabled server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		server := &MCPServer{Name: "enable-test", Command: []string{"cmd"}, Disabled: true}
		if err := mgr.Add(server); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		if err := mgr.Enable("enable-test"); err != nil {
			t.Fatalf("Enable() error = %v", err)
		}

		got, err := mgr.Get("enable-test")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Disabled {
			t.Error("Enable() did not set Disabled to false")
		}
	})

	t.Run("returns ErrMCPServerNotFound for nonexistent server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		err := mgr.Enable("nonexistent")
		if !errors.Is(err, ErrMCPServerNotFound) {
			t.Errorf("Enable() error = %v, want %v", err, ErrMCPServerNotFound)
		}
	})
}

func TestMCPManager_Disable(t *testing.T) {
	t.Run("disables enabled server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		server := &MCPServer{Name: "disable-test", Command: []string{"cmd"}, Disabled: false}
		if err := mgr.Add(server); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		if err := mgr.Disable("disable-test"); err != nil {
			t.Fatalf("Disable() error = %v", err)
		}

		got, err := mgr.Get("disable-test")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if !got.Disabled {
			t.Error("Disable() did not set Disabled to true")
		}
	})

	t.Run("returns ErrMCPServerNotFound for nonexistent server", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		err := mgr.Disable("nonexistent")
		if !errors.Is(err, ErrMCPServerNotFound) {
			t.Errorf("Disable() error = %v, want %v", err, ErrMCPServerNotFound)
		}
	})
}

func TestMCPManager_loadConfig(t *testing.T) {
	t.Run("missing file returns empty config", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewMCPManager(paths)

		// Don't create any config file
		servers, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(servers) != 0 {
			t.Errorf("List() with missing file returned %d servers, want 0", len(servers))
		}
	})

	t.Run("empty mcp map in file is initialized", func(t *testing.T) {
		paths := testPaths(t)

		// Create config file with null mcp
		configPath := paths.MCPConfigPath()
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte(`{"mcp": null}`), 0o644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		mgr := NewMCPManager(paths)

		// Should be able to add servers without error
		err := mgr.Add(&MCPServer{Name: "test", Command: []string{"cmd"}})
		if err != nil {
			t.Errorf("Add() after null mcp error = %v", err)
		}
	})
}

func TestAtomicWriteJSON(t *testing.T) {
	t.Run("creates file atomically", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		data := map[string]string{"key": "value"}
		if err := atomicWriteJSON(path, data); err != nil {
			t.Fatalf("atomicWriteJSON() error = %v", err)
		}

		// Verify file exists and has correct content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		var got map[string]string
		if err := json.Unmarshal(content, &got); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if got["key"] != "value" {
			t.Errorf("atomicWriteJSON() content[\"key\"] = %q, want %q", got["key"], "value")
		}
	})

	t.Run("creates parent directory if needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nested", "dir", "test.json")

		data := map[string]string{"nested": "true"}
		if err := atomicWriteJSON(path, data); err != nil {
			t.Fatalf("atomicWriteJSON() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("atomicWriteJSON() did not create file")
		}
	})

	t.Run("adds trailing newline", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "newline.json")

		if err := atomicWriteJSON(path, map[string]string{}); err != nil {
			t.Fatalf("atomicWriteJSON() error = %v", err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if len(content) == 0 || content[len(content)-1] != '\n' {
			t.Error("atomicWriteJSON() did not add trailing newline")
		}
	})
}

func TestMCPManager_RoundTrip(t *testing.T) {
	paths := testPaths(t)
	mgr := NewMCPManager(paths)

	original := &MCPServer{
		Name:        "full-server",
		Command:     []string{"npx", "run", "server"},
		Type:        "local",
		URL:         "",
		Environment: map[string]string{"API_KEY": "secret"},
		Headers:     map[string]string{"Authorization": "Bearer token"},
		Disabled:    false,
	}

	if err := mgr.Add(original); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := mgr.Get("full-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if len(got.Command) != len(original.Command) {
		t.Errorf("len(Command) = %d, want %d", len(got.Command), len(original.Command))
	}
	if got.Type != original.Type {
		t.Errorf("Type = %q, want %q", got.Type, original.Type)
	}
	if got.Environment["API_KEY"] != original.Environment["API_KEY"] {
		t.Errorf("Environment[API_KEY] = %q, want %q", got.Environment["API_KEY"], original.Environment["API_KEY"])
	}
	if got.Disabled != original.Disabled {
		t.Errorf("Disabled = %v, want %v", got.Disabled, original.Disabled)
	}
}
