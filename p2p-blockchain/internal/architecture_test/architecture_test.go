package architecture_test

import (
	"encoding/json"
	"fmt"
	"testing"

	archgo "github.com/arch-go/arch-go/api"
	config "github.com/arch-go/arch-go/api/configuration"
)

var subsystems = []string{"netzwerkrouting", "wallet", "miner", "blockchain", "app"}

func TestArchitecture(t *testing.T) {
	var rules []*config.DependenciesRule

	for _, s := range subsystems {
		// API Rule
		// api can depend on: itself, core (same subsystem), api (other subsystems)
		allowedApiDeps := []string{
			fmt.Sprintf("**.%s.api.**", s),
			fmt.Sprintf("**.%s.core.**", s),
		}

		allowedApiDeps = append(allowedApiDeps, addCommonDependencies(s)...)

		rules = append(rules, &config.DependenciesRule{
			Package: fmt.Sprintf("**.%s.api.**", s),
			ShouldOnlyDependsOn: &config.Dependencies{
				Internal: allowedApiDeps,
			},
		})

		// Core Rule
		// core can depend on: itself, data (same subsystem), api (other subsystems)
		allowedCoreDeps := []string{
			fmt.Sprintf("**.%s.core.**", s),
			fmt.Sprintf("**.%s.data.**", s),
		}

		allowedCoreDeps = append(allowedCoreDeps, addCommonDependencies(s)...)

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

// addCommonDependencies adds common dependencies to the given rules for a specific subsystem.
// This includes (1) dependencies on the common package and (2) APIs of other subsystems.
func addCommonDependencies(currentSubsystem string) []string {
	commonDeps := []string{
		"**.p2p-blockchain.internal.common.**",
	}

	// Add other subsystems' APIs
	for _, other := range subsystems {
		if other != currentSubsystem {
			commonDeps = append(commonDeps, fmt.Sprintf("**.%s.api.**", other))
		}
	}

	out := append(commonDeps, commonDeps...)

	return out
}
