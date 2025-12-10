// package architecture_test contains tests to verify the architectural constraints of the p2p-blockchain module.
// The architectural contraints are defined in the ARC42 documentation.
package architecture_test

import (
	"encoding/json"
	"fmt"
	"testing"

	archgo "github.com/arch-go/arch-go/api"
	config "github.com/arch-go/arch-go/api/configuration"
)

var subsystems = []string{"netzwerkrouting", "wallet", "miner", "blockchain", "app"} // Note technically "app" is not a subsystem, but we treat it as one for the purpose of architecture tests.

func TestArchitecture(t *testing.T) {
	var rules []*config.DependenciesRule

	for _, s := range subsystems {
		// API Rule
		// api can depend on: itself, core (same subsystem), common package
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
		// core can depend on: itself, data (same subsystem), common package
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

		// Data Rule
		// data can depend on: itself, common package
		allowedDataDeps := []string{
			fmt.Sprintf("**.%s.data.**", s),
		}

		allowedDataDeps = append(allowedDataDeps, addCommonDependencies(s)...)

		rules = append(rules, &config.DependenciesRule{
			Package: fmt.Sprintf("**.%s.data.**", s),
			ShouldOnlyDependsOn: &config.Dependencies{
				Internal: allowedDataDeps,
			},
		})

		// Infrastructure Rule
		// infrastructure can depend on: itself, api (same subsystem), core (same subsystem), data (same subsystem), common package
		allowedInfrastructureDeps := []string{
			fmt.Sprintf("**.%s.infrastructure.**", s),
			fmt.Sprintf("**.%s.api.**", s),
			fmt.Sprintf("**.%s.core.**", s),
			fmt.Sprintf("**.%s.data.**", s),
		}

		allowedInfrastructureDeps = append(allowedInfrastructureDeps, addCommonDependencies(s)...)
		allowedInfrastructureDeps = append(allowedInfrastructureDeps, addCommonInfrastructureDependencies()...)

		rules = append(rules, &config.DependenciesRule{
			Package: fmt.Sprintf("**.%s.infrastructure.**", s),
			ShouldOnlyDependsOn: &config.Dependencies{
				Internal: allowedInfrastructureDeps,
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

// addCommonInfrastructureDependencies adds common dependencies for the  infrastructure layer.
// This includes dependencies on protobuf generated packages.
func addCommonInfrastructureDependencies() []string {
	return []string{
		"**.p2p-blockchain.internal.pb.**",
	}
}
