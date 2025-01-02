package domain

import "testing"

func TestRunStage_Validate(t *testing.T) {
	type fields struct {
		Name        string
		Command     string
		Validator   func(s *RunStage) error
		ContOnError bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "no validator", fields: fields{Name: StageRun, Validator: nil, Command: "", ContOnError: false}, wantErr: false},
		{name: "failure", fields: fields{Name: StageRun, Validator: defaultRunStageValidator, Command: "", ContOnError: false}, wantErr: true},
		{name: "success", fields: fields{Name: StageRun, Validator: defaultRunStageValidator, Command: "echo 'hello'", ContOnError: false}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RunStage{
				Name:        tt.fields.Name,
				Command:     tt.fields.Command,
				Validator:   tt.fields.Validator,
				ContOnError: tt.fields.ContOnError,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("RunStage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildStage_Validate(t *testing.T) {
	type fields struct {
		Name           string
		DockerfilePath string
		Validator      func(s *BuildStage) error
		ContOnError    bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "no validator", fields: fields{Name: StageBuild, Validator: nil, DockerfilePath: "", ContOnError: false}, wantErr: false},
		{name: "failure", fields: fields{Name: StageBuild, Validator: defaultBuildStageValidator, DockerfilePath: "", ContOnError: false}, wantErr: true},
		{name: "success", fields: fields{Name: StageBuild, Validator: defaultBuildStageValidator, DockerfilePath: "Dockerfile", ContOnError: false}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BuildStage{
				Name:           tt.fields.Name,
				DockerfilePath: tt.fields.DockerfilePath,
				Validator:      tt.fields.Validator,
				ContOnError:    tt.fields.ContOnError,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("BuildStage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeployStage_Validate(t *testing.T) {
	type fields struct {
		Name         string
		ClusterName  string
		ManifestPath string
		Validator    func(s *DeployStage) error
		ContOnError  bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "no validator", fields: fields{Name: StageDeploy, Validator: nil, ClusterName: "", ManifestPath: "", ContOnError: false}, wantErr: false},
		{name: "failure", fields: fields{Name: StageDeploy, Validator: defaultDeployStageValidator, ClusterName: "", ManifestPath: "", ContOnError: false}, wantErr: true},
		{name: "success", fields: fields{Name: StageDeploy, Validator: defaultDeployStageValidator, ClusterName: "cluster", ManifestPath: "manifest.yaml", ContOnError: false}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DeployStage{
				Name:         tt.fields.Name,
				ClusterName:  tt.fields.ClusterName,
				ManifestPath: tt.fields.ManifestPath,
				Validator:    tt.fields.Validator,
				ContOnError:  tt.fields.ContOnError,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DeployStage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
