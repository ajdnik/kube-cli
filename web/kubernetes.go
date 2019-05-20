package web

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
)

func UpdateDeployment(namespace, name, container, docker string, info ClusterInfo) error {
	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:     "https://" + info.Endpoint,
		Username: info.Username,
		Password: info.Password,
		AuthProvider: &api.AuthProviderConfig{
			Name: "gcp",
		},
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CertData: info.CertData,
			CAData:   info.CAData,
			KeyData:  info.KeyData,
		},
	})
	if err != nil {
		return err
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		res, err := client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		found := false
		for i, c := range res.Spec.Template.Spec.Containers {
			if c.Name == container {
				res.Spec.Template.Spec.Containers[i].Image = docker
				found = true
			}
		}
		if !found {
			return fmt.Errorf("container spec for %v not found in %v deployment", container, name)
		}
		_, err = client.AppsV1().Deployments(namespace).Update(res)
		return err
	})
	return err
}