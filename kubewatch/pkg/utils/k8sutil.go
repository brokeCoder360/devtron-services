/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"flag"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"go.uber.org/zap"
	apps_v1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
)

type K8sUtilImpl struct {
	logger              *zap.SugaredLogger
	httpTransportConfig *k8s.CustomK8sHttpTransportConfig
	defaultK8sConfig    *rest.Config
}

type K8sUtil interface {
	GetK8sClientForConfig(config *rest.Config) (*kubernetes.Clientset, error)
	GetK8sConfigForCluster(clusterInfo *repository.Cluster) *rest.Config
}

func NewK8sUtilImpl(logger *zap.SugaredLogger,
	httpTransportConfig *k8s.CustomK8sHttpTransportConfig,
	defaultK8sConfig *rest.Config) *K8sUtilImpl {
	return &K8sUtilImpl{
		logger:              logger,
		httpTransportConfig: httpTransportConfig,
		defaultK8sConfig:    defaultK8sConfig,
	}
}

func (impl *K8sUtilImpl) GetK8sConfigForCluster(clusterInfo *repository.Cluster) *rest.Config {
	restConfig := &rest.Config{}
	if clusterInfo.ClusterName == commonBean.DEFAULT_CLUSTER {
		restConfig = impl.defaultK8sConfig
	} else {
		restConfig = &rest.Config{
			Host:            clusterInfo.ServerUrl,
			BearerToken:     clusterInfo.Config[commonBean.BearerToken],
			TLSClientConfig: rest.TLSClientConfig{Insecure: clusterInfo.InsecureSkipTlsVerify},
		}
		if !restConfig.TLSClientConfig.Insecure {
			restConfig.TLSClientConfig.KeyData = []byte(clusterInfo.Config[commonBean.TlsKey])
			restConfig.TLSClientConfig.CertData = []byte(clusterInfo.Config[commonBean.CertData])
			restConfig.TLSClientConfig.CAData = []byte(clusterInfo.Config[commonBean.CertificateAuthorityData])
		}
	}
	return restConfig
}

func (impl *K8sUtilImpl) GetK8sClientForConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	var err error
	config, err = impl.httpTransportConfig.OverrideConfigWithCustomTransport(config)
	if err != nil {
		impl.logger.Errorw("error in overriding config with custom transport", "err", err)
		return nil, err
	}
	httpClientFor, err := rest.HTTPClientFor(config)
	if err != nil {
		impl.logger.Errorw("error occurred while overriding k8s pubSubClient", "reason", err)
		return nil, err
	}
	clusterClient, err := kubernetes.NewForConfigAndClient(config, httpClientFor)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return nil, err
	}
	return clusterClient, nil
}

func GetDefaultK8sConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// handled for local dev config
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeConfig := flag.String("kubeconfig", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else {
		return cfg, nil
	}
}

// GetObjectMetaData returns metadata of a given k8s object
func GetObjectMetaData(obj interface{}) meta_v1.ObjectMeta {

	var objectMeta meta_v1.ObjectMeta

	switch object := obj.(type) {
	case *apps_v1.Deployment:
		objectMeta = object.ObjectMeta
	case *api_v1.ReplicationController:
		objectMeta = object.ObjectMeta
	case *apps_v1.ReplicaSet:
		objectMeta = object.ObjectMeta
	case *apps_v1.DaemonSet:
		objectMeta = object.ObjectMeta
	case *api_v1.Service:
		objectMeta = object.ObjectMeta
	case *api_v1.Pod:
		objectMeta = object.ObjectMeta
	case *batch_v1.Job:
		objectMeta = object.ObjectMeta
	case *api_v1.PersistentVolume:
		objectMeta = object.ObjectMeta
	case *api_v1.Namespace:
		objectMeta = object.ObjectMeta
	case *api_v1.Secret:
		objectMeta = object.ObjectMeta
	case *ext_v1beta1.Ingress:
		objectMeta = object.ObjectMeta
	case *api_v1.Event:
		objectMeta = object.ObjectMeta
	}
	return objectMeta
}
