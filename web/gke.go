package web

import (
	"context"
	"encoding/base64"

	"google.golang.org/api/container/v1"
)

// ClusterInfo contain auth data used to connect to Kubernetes.
type ClusterInfo struct {
	Username string
	Password string
	CAData   []byte
	KeyData  []byte
	CertData []byte
	Endpoint string
}

// GetGKECluster returns cluster config of a GKE.
func GetGKECluster(project, zone, name string) (ClusterInfo, error) {
	var info ClusterInfo
	ctx := context.Background()
	svc, err := container.NewService(ctx)
	if err != nil {
		return info, err
	}
	res, err := svc.Projects.Zones.Clusters.Get(project, zone, name).Context(ctx).Do()
	if err != nil {
		return info, err
	}
	info.Endpoint = res.Endpoint
	if res.MasterAuth != nil {
		info.Username = res.MasterAuth.Username
		info.Password = res.MasterAuth.Password
		info.CAData, err = base64.StdEncoding.DecodeString(res.MasterAuth.ClusterCaCertificate)
		if err != nil {
			return info, err
		}
		info.CertData, err = base64.StdEncoding.DecodeString(res.MasterAuth.ClientCertificate)
		if err != nil {
			return info, err
		}
		info.KeyData, err = base64.StdEncoding.DecodeString(res.MasterAuth.ClientKey)
		if err != nil {
			return info, err
		}
	}
	return info, nil
}
