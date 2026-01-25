package translate

import (
	"testing"
)

func TestYAMLToTOML(t *testing.T) {
	yamlInput := []byte("name: test\ndescription: a test\nenabled: true\n")
	tomlOutput, err := YAMLToTOML(yamlInput)
	if err != nil {
		t.Fatalf("YAMLToTOML failed: %v", err)
	}

	if len(tomlOutput) == 0 {
		t.Error("expected non-empty TOML output")
	}
}

func TestTOMLToYAML(t *testing.T) {
	tomlInput := []byte("name = \"test\"\ndescription = \"a test\"\nenabled = true\n")
	yamlOutput, err := TOMLToYAML(tomlInput)
	if err != nil {
		t.Fatalf("TOMLToYAML failed: %v", err)
	}

	if len(yamlOutput) == 0 {
		t.Error("expected non-empty YAML output")
	}
}
