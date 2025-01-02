package domain

import "fmt"

const (
	StageRun    = "run"
	StageBuild  = "build"
	StageDeploy = "deploy"
)

type Stage interface {
	Validate() error
	ContinueOnError() bool
}

// RunStage is a stage that runs an arbitrary command (lint, test) and implements the Stage interface
type RunStage struct {
	Name        string                  `json:"name"`
	Command     string                  `json:"command"`
	Validator   func(s *RunStage) error `json:"-"`
	ContOnError bool                    `json:"cont_on_error"`
}

func NewRunStage(name, command string, continueOnError bool) *RunStage {
	return &RunStage{Name: name, Command: command, Validator: defaultRunStageValidator, ContOnError: continueOnError}
}

func (s *RunStage) Validate() error {
	if s.Validator != nil {
		return s.Validator(s)
	}
	return nil
}

func (s *RunStage) ContinueOnError() bool {
	return s.ContOnError
}

// defaultRunStageValidator is the default validator func for a "run" stage.
func defaultRunStageValidator(s *RunStage) error {
	if s.Command == "" {
		return fmt.Errorf("command is required for run stage")
	}
	return nil
}

// BuildStage is a stage that builds a docker image from a given Dockerfile and implements the Stage interface
type BuildStage struct {
	Name           string                    `json:"name"`
	DockerfilePath string                    `json:"dockerfile_path"`
	Validator      func(s *BuildStage) error `json:"-"`
	ContOnError    bool                      `json:"cont_on_error"`
}

func NewBuildStage(name, dockerfilePath string, continueOnError bool) *BuildStage {
	return &BuildStage{Name: name, DockerfilePath: dockerfilePath, Validator: defaultBuildStageValidator, ContOnError: continueOnError}
}

func (s *BuildStage) Validate() error {
	if s.Validator != nil {
		return s.Validator(s)
	}
	return nil
}

func (s *BuildStage) ContinueOnError() bool {
	return s.ContOnError
}

// defaultBuildStageValidator is the default validator func for a "build" stage.
func defaultBuildStageValidator(s *BuildStage) error {
	if s.DockerfilePath == "" {
		return fmt.Errorf("dockerfile path is required for build stage")
	}
	return nil
}

// DeployStage is a stage that deploys a kubernetes manifest to a given cluster and implements the Stage interface
type DeployStage struct {
	Name         string                     `json:"name"`
	ClusterName  string                     `json:"cluster_name"`
	ManifestPath string                     `json:"manifest_path"`
	Validator    func(s *DeployStage) error `json:"-"`
	ContOnError  bool                       `json:"cont_on_error"`
}

func NewDeployStage(name, clusterName, manifestPath string, continueOnError bool) *DeployStage {
	return &DeployStage{Name: name, ClusterName: clusterName, ManifestPath: manifestPath, Validator: defaultDeployStageValidator, ContOnError: continueOnError}
}

func (s *DeployStage) Validate() error {
	if s.Validator != nil {
		return s.Validator(s)
	}
	return nil
}

func (s *DeployStage) ContinueOnError() bool {
	return s.ContOnError
}

// defaultDeployStageValidator is the default validator for a "deploy" stage.
func defaultDeployStageValidator(s *DeployStage) error {
	if s.ClusterName == "" {
		return fmt.Errorf("cluster name is required for deploy stage")
	}
	if s.ManifestPath == "" {
		return fmt.Errorf("manifest path is required for deploy stage")
	}
	return nil
}
