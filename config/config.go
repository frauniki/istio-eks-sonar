package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type GenerateConfig struct {
	SecretNamePrefix string        `yaml:"secretNamePrefix" default:"istio-remote-secret-"`
	TargetNamespaces []string      `yaml:"targetNamespaces" validate:"required"`
	ClusterName      string        `yaml:"clusterName"`
	EKSClusters      []EKSClusters `yaml:"eksClusters" validate:"required"`
	ExecTemplate     ExecTemplate  `yaml:"execTemplate" validate:"required"`
}

type EKSClusters struct {
	AssumeRoleARN string   `yaml:"assumeRoleARN"`
	Clusters      []string `yaml:"clusters" validate:"required"`
}

type ExecTemplate struct {
	APIVersion string   `yaml:"apiVersion" default:"client.authentication.k8s.io/v1beta1"`
	Command    string   `yaml:"command" default:"aws-iam-authenticator"`
	Args       []string `yaml:"args" default:"[\"token\", \"-i\", \"{{ .clusterName }}\"]"`
	Env        []struct {
		Name  string `yaml:"name" validate:"required"`
		Value string `yaml:"value" validate:"required"`
	} `yaml:"env"`
	InteractiveMode     string `yaml:"interactiveMode" default:"IfAvailable"`
	ProviderClusterInfo bool   `yaml:"providerClusterInfo" default:"false"`
}

func LoadConfig(path string) (*GenerateConfig, error) {
	cfg := &GenerateConfig{}
	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	validate := validator.New()
	dec := yaml.NewDecoder(buf, yaml.Validator(validate))

	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
