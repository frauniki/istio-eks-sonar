package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type GenerateConfig struct {
	SecretNamePrefix   string             `yaml:"secretNamePrefix"`
	TargetNamespaces   []string           `yaml:"targetNamespaces"`
	EKSClusters        []EKSClusters      `yaml:"eksClusters"`
	ExecTemplateConfig ExecTemplateConfig `yaml:"execTemplateConfig"`
}

type EKSClusters struct {
	AssumeRole string   `yaml:"assumeRole"`
	Clusters   []string `yaml:"clusters"`
}

type ExecTemplateConfig struct {
	APIVersion string   `yaml:"apiVersion"`
	Command    string   `yaml:"command"`
	Args       []string `yaml:"args"`
	Env        []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"env"`
	InteractiveMode     string `yaml:"interactiveMode"`
	ProviderClusterInfo bool   `yaml:"providerClusterInfo"`

	Data map[string]string `yaml:"data"`
}

func LoadConfig(path string) (*GenerateConfig, error) {
	cfg := GenerateConfig{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
