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

package cluster

import (
	"fmt"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/informer/bean"
)

type StopAdvisor interface {
	StopInformerForCluster(clusterId int) error
	StopAll()
}

type ClientAdvisor interface {
	StartInformerForCluster(clusterInfo *repository.Cluster) error
	StopAdvisor
}

func (impl *InformerImpl) GetClient(clientType bean.ClientType, clusterInfo *repository.Cluster) (ClientAdvisor, error) {
	if !impl.IsMultiClusterMode(clientType) && !clusterInfo.IsDefault() {
		impl.logger.Debugw("informer is not supported for cluster, skipping...", "clientType", clientType, "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName, "appConfig", impl.appConfig)
		return NewUnimplementedAdvisor(), nil
	}
	return impl.GetClientAdvisor(clientType)
}

func (impl *InformerImpl) GetClientStopper(clientType bean.ClientType) (StopAdvisor, error) {
	return impl.GetClientAdvisor(clientType)
}

func (impl *InformerImpl) GetClientAdvisor(clientType bean.ClientType) (ClientAdvisor, error) {
	switch clientType {
	case bean.ArgoCDClientType:
		return impl.argoCdInformer, nil
	case bean.CiArgoWorkflowClientType:
		return impl.ciWfInformer, nil
	case bean.CdArgoWorkflowClientType:
		return impl.cdWfInformer, nil
	case bean.SystemExecutorClientType:
		return impl.systemExecInformer, nil
	default:
		return NewUnimplementedAdvisor(), fmt.Errorf("client type %q not supported", clientType)
	}
}

func (impl *InformerImpl) IsMultiClusterMode(clientType bean.ClientType) bool {
	switch clientType {
	case bean.ArgoCDClientType:
		return impl.appConfig.IsMultiClusterArgoCD()
	case bean.CiArgoWorkflowClientType:
		return impl.appConfig.IsMultiClusterCiArgoWfType()
	case bean.CdArgoWorkflowClientType:
		return impl.appConfig.IsMultiClusterCdArgoWfType()
	case bean.SystemExecutorClientType:
		return impl.appConfig.IsMultiClusterSystemExec()
	default:
		return false
	}
}
