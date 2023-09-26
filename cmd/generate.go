package cmd

import (
	"context"

	"github.com/frauniki/istio-eks-sonar/cloud"
	"github.com/frauniki/istio-eks-sonar/config"
	"github.com/frauniki/istio-eks-sonar/kube"
	"github.com/spf13/cobra"
)

var generateCmd = func() *cobra.Command {
	var flags struct {
		config string
	}

	cmd := &cobra.Command{
		Use: "generate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(flags.config)
		},
	}

	cmd.Flags().StringVarP(&flags.config, "config", "c", "", "config file")

	return cmd
}()

func init() {
	rootCmd.AddCommand(generateCmd)
}

func generate(configPath string) error {
	ctx := context.Background()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	kubeconfigs, err := cloud.GenerateKubeconfigs(ctx, cfg)
	if err != nil {
		return err
	}
	if err != kube.CreateRemoteSecrets(ctx, cfg, kubeconfigs) {
		return err
	}

	return nil
}
