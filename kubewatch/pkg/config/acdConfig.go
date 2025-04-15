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

// CATEGORY=ARGOCD_INFORMER
type AcdConfig struct {
	// ACDNamespace is the namespace where all the ArgoCD application objects are published
	// For multi-cluster mode, it will be set to v1.NamespaceAll
	ACDNamespace string `env:"ACD_NAMESPACE" envDefault:"devtroncd" description:"Namespace where all the ArgoCD application objects are published. For multi-cluster mode, it will be set to v1.NamespaceAll" deprecated:"false"`

	// ACDInformer is used to determine whether ArgoCD informer is enabled or not
	ACDInformer bool `env:"ACD_INFORMER" envDefault:"true" description:"Used to determine whether ArgoCD informer is enabled or not" deprecated:"false"`
}

func getAcdConfig() (*AcdConfig, error) {
	acdCfg := &AcdConfig{}
	err := env.Parse(acdCfg)
	return acdCfg, err
}
