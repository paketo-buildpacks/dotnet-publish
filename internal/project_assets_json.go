package internal

import (
	"encoding/json"
	"fmt"
)

type ProjectAssetsJSON struct {
	Targets Targets
}

type Targets []Target

type Target struct {
	Name         string
	Dependencies Dependencies
}

type Dependencies []ProjectDependency

type ProjectDependency struct {
	Name           string
	Type           string         `json:"type"`
	Runtime        Runtime        `json:"runtime"`
	RuntimeTargets RuntimeTargets `json:"runtimeTargets"`
}

type Runtime string

type RuntimeTargets []RuntimeTarget

type RuntimeTarget struct {
	FileName          string
	AssetType         string
	RuntimeIdentifier string
}

func (ts *Targets) UnmarshalJSON(data []byte) error {
	var result []Target
	var v map[string]Dependencies
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	for name, deps := range v {
		result = append(result, Target{
			Name:         name,
			Dependencies: deps,
		})
	}
	*ts = Targets(result)
	return nil
}

func (ds *Dependencies) UnmarshalJSON(data []byte) error {
	var result []ProjectDependency
	var v map[string]ProjectDependency
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	for name, dep := range v {
		dep.Name = name
		result = append(result, dep)
	}
	*ds = Dependencies(result)
	return nil
}

func (r *Runtime) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v) != 1 {
		return fmt.Errorf("dependency runtime file malformed")
	}
	for key := range v {
		*r = Runtime(key)
	}
	return nil
}

func (ts *RuntimeTargets) UnmarshalJSON(data []byte) error {
	result := []RuntimeTarget{}
	var v map[string]map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	for key := range v {
		result = append(result, RuntimeTarget{
			FileName:          key,
			AssetType:         v[key]["assetType"].(string),
			RuntimeIdentifier: v[key]["rid"].(string),
		})
	}
	*ts = RuntimeTargets(result)
	return nil
}
