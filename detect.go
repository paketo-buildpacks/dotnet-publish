package dotnetpublish

import (
	"fmt"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

type BuildPlanMetadata struct {
	Version string `toml:"version,omitempty"`
	Build   bool   `toml:"build"`
	Launch  bool   `toml:"launch"`
}

//go:generate faux --interface ProjectParser --output fakes/project_parser.go
type ProjectParser interface {
	ASPNetIsRequired(path string) (bool, error)
	NodeIsRequired(path string) (bool, error)
	NPMIsRequired(path string) (bool, error)
}

//go:generate faux --interface BuildpackYMLParser --output fakes/buildpack_yml_parser.go
type BuildpackYMLParser interface {
	ParseProjectPath(path string) (projectFilePath string, err error)
}

func Detect(parser ProjectParser, buildpackYMLParser BuildpackYMLParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var projectPath string
		projectPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
		}

		matches, err := filepath.Glob(filepath.Join(context.WorkingDir, projectPath, "*.?sproj"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if len(matches) == 0 {
			return packit.DetectResult{}, packit.Fail.WithMessage("no project file found")
		}

		projectFilePath := matches[0]

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "dotnet-sdk",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
			{
				Name: "dotnet-runtime",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
			{
				Name: "icu",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
		}

		aspNetReq, err := parser.ASPNetIsRequired(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if aspNetReq {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-aspnetcore",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			})
		}

		nodeReq, err := parser.NodeIsRequired(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if nodeReq {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "node",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			})
		}

		npmReq, err := parser.NPMIsRequired(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if npmReq {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "npm",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "build"},
				},
				Requires: requirements,
			},
		}, nil
	}
}
