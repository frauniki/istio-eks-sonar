package cloud

import (
	"bytes"
	"context"
	"text/template"

	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/frauniki/istio-eks-sonar/config"
	"k8s.io/client-go/tools/clientcmd/api"
)

func GenerateKubeconfigs(ctx context.Context, config *config.GenerateConfig) (map[string]*api.Config, error) {
	cfg, err := aws_config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	kubeconfigs := map[string]*api.Config{}

	for _, eksCluster := range config.EKSClusters {
		innerCfg := cfg
		if eksCluster.AssumeRole != "" {
			assumeRole(&innerCfg, eksCluster.AssumeRole)
		}
		eksClient := eks.NewFromConfig(innerCfg)

		for _, cluster := range eksCluster.Clusters {
			resp, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
				Name: &cluster,
			})
			if err != nil {
				return nil, err
			}
			kubeconfig, err := genKubeconfig(config.ExecTemplateConfig, cluster, []byte(*resp.Cluster.CertificateAuthority.Data), *resp.Cluster.Endpoint)
			if err != nil {
				return nil, err
			}
			kubeconfigs[cluster] = kubeconfig
		}
	}

	return kubeconfigs, nil
}

func genKubeconfig(execConfig config.ExecTemplateConfig, clusterName string, caData []byte, server string) (*api.Config, error) {
	templateData := make(map[string]string, len(execConfig.Data)+1)
	for k, v := range execConfig.Data {
		templateData[k] = v
	}
	templateData["clusterName"] = clusterName

	command, err := templateExec(execConfig.Command, templateData)
	if err != nil {
		return nil, err
	}
	args := make([]string, len(execConfig.Args))
	for _, a := range execConfig.Args {
		arg, err := templateExec(a, templateData)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	env := make([]api.ExecEnvVar, len(execConfig.Env))
	for _, e := range execConfig.Env {
		v, err := templateExec(e.Value, templateData)
		if err != nil {
			return nil, err
		}
		env = append(env, api.ExecEnvVar{
			Name:  e.Name,
			Value: v,
		})
	}

	return &api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				CertificateAuthorityData: caData,
				Server:                   server,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			clusterName: {
				Exec: &api.ExecConfig{
					APIVersion:         execConfig.APIVersion,
					Command:            command,
					Args:               args,
					Env:                env,
					InteractiveMode:    api.ExecInteractiveMode(execConfig.InteractiveMode),
					ProvideClusterInfo: execConfig.ProviderClusterInfo,
				},
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: clusterName,
			},
		},
		CurrentContext: clusterName,
	}, nil
}

func templateExec(tmpl string, data map[string]string) (string, error) {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
