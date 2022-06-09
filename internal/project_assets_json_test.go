package internal_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/dotnet-publish/internal"
	"github.com/sclevine/spec"
)

func testTargets(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect  = NewWithT(t).Expect
		targets internal.Targets
	)
	var input []byte = []byte(`{
    ".NETCoreApp,Version=v3.1": {
      "Microsoft.AspNetCore.Diagnostics.HealthChecks/2.2.0-rc1": {
        "type": "package",
        "dependencies": {
          "Microsoft.AspNetCore.Http.Abstractions": "2.2.0",
          "Microsoft.Net.Http.Headers": "2.2.0"
        },
        "compile": {
          "lib/netstandard2.0/Microsoft.AspNetCore.Diagnostics.HealthChecks.dll": {}
        },
        "runtime": {
          "lib/netstandard2.0/Microsoft.AspNetCore.Diagnostics.HealthChecks.dll": {}
        }
      }
    },
    ".NETCoreApp,Version=v6.0": {
      "Consul/0.7.2.6": {
        "type": "package",
        "dependencies": {
          "NETStandard.Library": "1.6.1",
          "System.Threading.Thread": "4.0.0"
        },
        "compile": {
          "lib/netstandard1.3/Consul.dll": {}
        },
        "runtime": {
          "lib/netstandard1.3/Consul.dll": {}
        }
      }
    }
  }
`)
	var badInput []byte = []byte(`{
    ".NETCoreApp,Version=v3.1": {
      "Consul/0.7.2.6": {
        "type": "package",
        "dependencies": {
          "NETStandard.Library": "1.6.1",
          "System.Threading.Thread": "4.0.0"
        },
        "compile": {
          "lib/netstandard1.3/Consul.dll": {}
        },
        "runtime": {
          "lib/netstandard1.3/Consul.dll": {}
        }
      },
      "Microsoft.AspNetCore.Diagnostics.HealthChecks/2.2.0-rc1": {
        "type": "package",
        "dependencies": {
          "Microsoft.AspNetCore.Http.Abstractions": "2.2.0",
          "Microsoft.Net.Http.Headers": "2.2.0"
        },
        "compile": {
          "lib/netstandard2.0/Microsoft.AspNetCore.Diagnostics.HealthChecks.dll": {}
        },
        "runtime": {
          "lib/netstandard2.0/Microsoft.AspNetCore.Diagnostics.HealthChecks.dll": {}
        }
      }
    },
    ".NETCoreApp,Version=v6.0": {
			"some-garbage" : true
    }
  }
`)

	context("UnmarshalJSON", func() {
		it("correctly unmarshals JSON", func() {

			err := json.Unmarshal(input, &targets)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(targets)).To(Equal(2))
			// Must use ContainElement since maps are non-ordered; array ends up different after each unmarshalling
			Expect(targets).To(ContainElement(internal.Target{
				Name: ".NETCoreApp,Version=v3.1",
				Dependencies: internal.Dependencies([]internal.ProjectDependency{
					{
						Name:                "Microsoft.AspNetCore.Diagnostics.HealthChecks/2.2.0-rc1",
						Type:                "package",
						RuntimeDependencies: []string{"lib/netstandard2.0/Microsoft.AspNetCore.Diagnostics.HealthChecks.dll"},
					},
				}),
			}))
			Expect(targets).To(ContainElement(internal.Target{
				Name: ".NETCoreApp,Version=v6.0",
				Dependencies: internal.Dependencies([]internal.ProjectDependency{
					{
						Name:                "Consul/0.7.2.6",
						Type:                "package",
						RuntimeDependencies: []string{"lib/netstandard1.3/Consul.dll"},
					},
				}),
			}))
		})
		context("failure cases", func() {
			context("when a target has an invalid dependency list structure", func() {
				it("returns an error", func() {

					err := json.Unmarshal(badInput, &targets)
					Expect(err).To(MatchError(ContainSubstring("cannot unmarshal bool into Go value of type internal.ProjectDependency")), err.Error())
				})
			})
		})
	})
}
func testDependencies(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		deps   internal.Dependencies
	)
	var input []byte = []byte(`{
      "Consul/0.7.2.6": {
        "type": "package",
        "dependencies": {
          "NETStandard.Library": "1.6.1",
          "System.Threading.Thread": "4.0.0"
        },
        "compile": {
          "lib/netstandard1.3/Consul.dll": {}
        },
        "runtime": {
          "lib/netstandard1.3/Consul.dll": {}
        }
      },
      "Microsoft.Win32.Registry/4.6.0": {
        "type": "package",
        "dependencies": {
          "System.Security.AccessControl": "4.6.0",
          "System.Security.Principal.Windows": "4.6.0"
        },
        "compile": {
          "ref/netstandard2.0/_._": {}
        },
        "runtime": {
          "lib/netstandard2.0/Microsoft.Win32.Registry.dll": {}
        },
        "runtimeTargets": {
          "runtimes/unix/lib/netstandard2.0/Microsoft.Win32.Registry.dll": {
            "assetType": "runtime",
            "rid": "unix"
          }
        }
      }
}`)
	var badInput []byte = []byte(`{
      "Consul/0.7.2.6": {
        "type": "package",
        "dependencies": {
          "NETStandard.Library": "1.6.1",
          "System.Threading.Thread": "4.0.0"
        },
        "compile": {
          "lib/netstandard1.3/Consul.dll": {}
        },
        "runtime": {
          "lib/netstandard1.3/Consul.dll": {}
        }
      },
			"some-garbage" : true
}`)

	context("UnmarshalJSON", func() {
		it("correctly unmarshals JSON", func() {

			err := json.Unmarshal(input, &deps)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(deps)).To(Equal(2))
			Expect(deps).To(ContainElement(internal.ProjectDependency{
				Name:                "Consul/0.7.2.6",
				Type:                "package",
				RuntimeDependencies: []string{"lib/netstandard1.3/Consul.dll"},
			}))
			Expect(deps).To(ContainElement(internal.ProjectDependency{
				Name:                "Microsoft.Win32.Registry/4.6.0",
				Type:                "package",
				RuntimeDependencies: []string{"lib/netstandard2.0/Microsoft.Win32.Registry.dll"},
				RuntimeTargets: internal.RuntimeTargets([]internal.RuntimeTarget{
					{
						FileName:          "runtimes/unix/lib/netstandard2.0/Microsoft.Win32.Registry.dll",
						AssetType:         "runtime",
						RuntimeIdentifier: "unix",
					},
				}),
			}))
		})

		context("failure cases", func() {
			context("when a target has an invalid dependency list structure", func() {
				it("returns an error", func() {

					err := json.Unmarshal(badInput, &deps)
					Expect(err).To(MatchError(ContainSubstring("cannot unmarshal bool into Go value of type internal.ProjectDependency")), err.Error())
				})
			})
		})
	})
}

func testRuntimeTargets(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		rts    internal.RuntimeTargets
	)
	var input []byte = []byte(`{
          "runtimes/unix/lib/netstandard2.0/Microsoft.Win32.Registry.dll": {
            "assetType": "runtime",
            "rid": "unix"
          },
          "runtimes/win/lib/netstandard2.0/Microsoft.Win32.Registry.dll": {
            "assetType": "runtime",
            "rid": "win"
          }
        }`)
	var badInput []byte = []byte(`{
          "runtimes/unix/lib/netstandard2.0/Microsoft.Win32.Registry.dll": {
            "assetType": "runtime",
            "rid": "unix"
          },
          "some-garbage": true
        }`)

	context("UnmarshalJSON", func() {
		it("correctly unmarshals JSON", func() {

			err := json.Unmarshal(input, &rts)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(rts)).To(Equal(2))
			Expect(rts).To(ContainElement(internal.RuntimeTarget{
				FileName:          "runtimes/unix/lib/netstandard2.0/Microsoft.Win32.Registry.dll",
				AssetType:         "runtime",
				RuntimeIdentifier: "unix",
			}))
			Expect(rts).To(ContainElement(internal.RuntimeTarget{
				FileName:          "runtimes/win/lib/netstandard2.0/Microsoft.Win32.Registry.dll",
				AssetType:         "runtime",
				RuntimeIdentifier: "win",
			}))
		})

		context("failure cases", func() {
			context("when a target has an invalid dependency list structure", func() {
				it("returns an error", func() {

					err := json.Unmarshal(badInput, &rts)
					Expect(err).To(MatchError(ContainSubstring("cannot unmarshal bool into Go value of type map[string]interface {}")), err.Error())
				})
			})
		})
	})
}

func testRuntimeDependencies(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect  = NewWithT(t).Expect
		runtime internal.RuntimeDependencies
	)
	var input []byte = []byte(`{
          "lib/netstandard1.3/Consul.dll": {},
          "lib/netstandard2.0/Microsoft.Diagnostics.FastSerialization.dll": {}
        }`)
	var badInput []byte = []byte(`[ "not-a-map" ]`)

	context("UnmarshalJSON", func() {
		it("correctly unmarshals JSON", func() {

			err := json.Unmarshal(input, &runtime)
			Expect(err).NotTo(HaveOccurred())

			Expect(runtime).To(Equal(internal.RuntimeDependencies([]string{
				"lib/netstandard1.3/Consul.dll",
				"lib/netstandard2.0/Microsoft.Diagnostics.FastSerialization.dll",
			})))
		})

		context("failure cases", func() {
			context("when a target has an invalid dependency list structure", func() {
				it("returns an error", func() {

					err := json.Unmarshal(badInput, &runtime)
					Expect(err).To(MatchError(ContainSubstring("cannot unmarshal array into Go value of type map[string]interface {}")))
				})
			})
		})
	})
}
