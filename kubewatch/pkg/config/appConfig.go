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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type AppConfig struct {
	ExternalConfig *ExternalConfig
	ClusterCfg     *ClusterConfig
	CiConfig       *CiConfig
	CdConfig       *CdConfig
	AcdConfig      *AcdConfig
	Timeout        *Timeout
}

func (app *AppConfig) GetClusterConfig() *ClusterConfig {
	return app.ClusterCfg
}

func (app *AppConfig) GetExternalConfig() *ExternalConfig {
	return app.ExternalConfig
}

func (app *AppConfig) GetCiWfNamespace() string {
	if app.GetExternalConfig().External {
		return app.GetExternalConfig().Namespace
	} else if app.IsMultiClusterCiArgoWfType() {
		return metav1.NamespaceAll
	} else {
		return app.GetCiConfig().DefaultNamespace
	}
}

func (app *AppConfig) GetCdWfNamespace() string {
	if app.GetExternalConfig().External {
		return app.GetExternalConfig().Namespace
	} else if app.IsMultiClusterCdArgoWfType() {
		return metav1.NamespaceAll
	} else {
		return app.GetCdConfig().DefaultNamespace
	}
}

func (app *AppConfig) GetCiConfig() *CiConfig {
	return app.CiConfig
}

func (app *AppConfig) GetCdConfig() *CdConfig {
	return app.CdConfig
}

func (app *AppConfig) GetAcdConfig() *AcdConfig {
	return app.AcdConfig
}

func (app *AppConfig) GetACDNamespace() string {
	if app.IsMultiClusterArgoCD() {
		return metav1.NamespaceAll
	} else {
		return app.AcdConfig.ACDNamespace
	}
}

func (app *AppConfig) GetTimeout() *Timeout {
	return app.Timeout
}

func (app *AppConfig) IsDBAvailable() bool {
	return !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterArgoCD() bool {
	return app.GetClusterConfig().ClusterArgoCDType == AllClusterType && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterCdArgoWfType() bool {
	return app.GetClusterConfig().ClusterCdArgoWfType == AllClusterType && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterCiArgoWfType() bool {
	return app.GetClusterConfig().ClusterCiArgoWfType == AllClusterType && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterSystemExec() bool {
	return app.GetClusterConfig().SystemExecClusterType == AllClusterType && !app.GetExternalConfig().External
}

func GetAppConfig() (*AppConfig, error) {
	clusterConfig, err := getClusterConfig()
	if err != nil {
		return nil, err
	}
	externalConfig, err := getExternalConfig()
	if err != nil {
		return nil, err
	}
	ciConfig, err := getCiConfig()
	if err != nil {
		return nil, err
	}
	cdConfig, err := getCdConfig()
	if err != nil {
		return nil, err
	}
	acdConfig, err := getAcdConfig()
	if err != nil {
		return nil, err
	}
	timeout, err := getTimeout()
	if err != nil {
		return nil, err
	}
	return &AppConfig{
		ClusterCfg:     clusterConfig,
		ExternalConfig: externalConfig,
		CiConfig:       ciConfig,
		CdConfig:       cdConfig,
		AcdConfig:      acdConfig,
		Timeout:        timeout,
	}, nil
}
