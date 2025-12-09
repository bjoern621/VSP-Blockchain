package architecture_test

import (
	"encoding/json"
	"testing"

	archgo "github.com/arch-go/arch-go/api"
	config "github.com/arch-go/arch-go/api/configuration"
)

func TestArchitecture(t *testing.T) {
	configuration := config.Config{
		DependenciesRules: []*config.DependenciesRule{
			{
				Package: "**.core.**",
				ShouldNotDependsOn: &config.Dependencies{
					Internal: []string{
						"**.infrastructure.**",
						"**.api.**",
					},
				},
			},
			{
				Package: "**.netzwerkrouting.**",
				ShouldNotDependsOn: &config.Dependencies{
					Internal: []string{
						"**.blockchain.**",
					},
				},
			},
		},
	}
	moduleInfo := config.Load("s3b/vsp-blockchain/p2p-blockchain")

	result := archgo.CheckArchitecture(moduleInfo, configuration)

	if !result.Pass {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		t.Fatalf("Architecture tests failed:\n%s", string(jsonBytes))
	}
}
