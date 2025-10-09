package config

import (
	"ai-team/pkg/types"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigChainMapUnmarshal(t *testing.T) {
	yamlData := `
chains:
  test-chain:
    steps:
      - name: "role1"
        input:
          foo: "bar"
        output_key: "out1"
      - name: "role2"
        input:
          bar: "baz"
        output_key: "out2"
`
	var cfg struct {
		Chains map[string]types.RoleChain `yaml:"chains"`
	}
	if err := yaml.Unmarshal([]byte(yamlData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	chain, ok := cfg.Chains["test-chain"]
	if !ok {
		t.Fatalf("chain 'test-chain' not found in config")
	}
	if len(chain.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(chain.Steps))
	}
	if chain.Steps[0].Name != "role1" || chain.Steps[1].Name != "role2" {
		t.Errorf("unexpected step names: %+v", chain.Steps)
	}
	if !reflect.DeepEqual(chain.Steps[0].Input["foo"], "bar") {
		t.Errorf("unexpected input for step 0: %+v", chain.Steps[0].Input)
	}
}
