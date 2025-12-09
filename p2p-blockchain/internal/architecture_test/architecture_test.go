package architecture_test

import (
	"encoding/json"
	"fmt"
	"testing"

	archgo "github.com/arch-go/arch-go/api"
	config "github.com/arch-go/arch-go/api/configuration"
)

func TestArchitecture(t *testing.T) {
	subsystems := []string{"netzwerkrouting", "wallet", "miner", "blockchain", "app"}
	var rules []*config.DependenciesRule

	for _, s := range subsystems {
		// API Rule
		// api can depend on: core (same subsystem), api (other subsystems)
		allowedApiDeps := []string{
			fmt.Sprintf("**.%s.api.**", s),
			fmt.Sprintf("**.%s.core.**", s),
		}
		// Add other subsystems' APIs
		for _, other := range subsystems {
			if other != s {
				allowedApiDeps = append(allowedApiDeps, fmt.Sprintf("**.%s.api.**", other))
			}
		}

		rules = append(rules, &config.DependenciesRule{
			Package: fmt.Sprintf("**.%s.api.**", s),
			ShouldOnlyDependsOn: &config.Dependencies{
				Internal: allowedApiDeps,
			},
		})

		// Core Rule
		// core can depend on: data (same subsystem), api (other subsystems)
		allowedCoreDeps := []string{
			fmt.Sprintf("**.%s.core.**", s),
			fmt.Sprintf("**.%s.data.**", s),
			"**.p2p-blockchain.internal.common.**",
		}
		for _, other := range subsystems {
			if other != s {
				allowedCoreDeps = append(allowedCoreDeps, fmt.Sprintf("**.%s.api.**", other))
			}
		}

		rules = append(rules, &config.DependenciesRule{
			Package: fmt.Sprintf("**.%s.core.**", s),
			ShouldOnlyDependsOn: &config.Dependencies{
				Internal: allowedCoreDeps,
			},
		})
	}

	configuration := config.Config{
		DependenciesRules: rules,
	}
	moduleInfo := config.Load("s3b/vsp-blockchain/p2p-blockchain")

	result := archgo.CheckArchitecture(moduleInfo, configuration)

	if !result.Pass {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		t.Fatalf("Architecture tests failed:\n%s", string(jsonBytes))
	}
}
