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

package config

import "github.com/caarlos0/env"

// CATEGORY=CLUSTER_MODE
type ClusterConfig struct {
	// SystemExecClusterType uses CLUSTER_TYPE env variable; for backward compatibility
	//	- AllClusterType: All clusters are enabled for SystemExec informer
	//	- InClusterType: Only default cluster is enabled for SystemExec informer
	SystemExecClusterType string `env:"CLUSTER_TYPE" envDefault:"IN_CLUSTER" description:"Determines cluster mode for System Executor informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER" deprecated:"false"`

	// ClusterArgoCDType defines whether all clusters are enabled for ArgoCD informer
	//	- AllClusterType: All clusters are enabled for ArgoCD informer
	//	- InClusterType: Only default cluster is enabled for ArgoCD informer
	ClusterArgoCDType string `env:"CLUSTER_ARGO_CD_TYPE" envDefault:"IN_CLUSTER" description:"Determines cluster mode for ArgoCD informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER" deprecated:"false"`

	// ClusterCiArgoWfType defines whether all clusters are enabled for CI ArgoWorkflow informer
	//	- AllClusterType: All clusters are enabled for CI ArgoWorkflow informer
	//	- InClusterType: Only default cluster is enabled for CI ArgoWorkflow informer
	ClusterCiArgoWfType string `env:"CLUSTER_CI_ARGO_WF_TYPE" envDefault:"IN_CLUSTER" description:"Determines cluster mode for CI ArgoWorkflow informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER" deprecated:"false"`

	// ClusterCdArgoWfType defines whether all clusters are enabled for CD ArgoWorkflow informer
	//	- AllClusterType: All clusters are enabled for CD ArgoWorkflow informer
	//	- InClusterType: Only default cluster is enabled for CD ArgoWorkflow informer
	ClusterCdArgoWfType string `env:"CLUSTER_CD_ARGO_WF_TYPE" envDefault:"IN_CLUSTER" description:"Determines cluster mode for CD ArgoWorkflow informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER" deprecated:"false"`
}

func getClusterConfig() (*ClusterConfig, error) {
	clusterCfg := &ClusterConfig{}
	err := env.Parse(clusterCfg)
	if err != nil {
		return nil, err
	}
	return clusterCfg, err
}
