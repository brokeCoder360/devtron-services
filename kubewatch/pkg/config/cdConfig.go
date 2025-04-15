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

// CATEGORY=CD_ARGO_WORKFLOW
type CdConfig struct {
	// DefaultNamespace is the namespace where all CD workflows are scheduled
	DefaultNamespace string `env:"CD_DEFAULT_NAMESPACE" envDefault:"devtron-cd" description:"Namespace where all CD workflows objects are scheduled. For multi-cluster mode, it will be set to v1.NamespaceAll" deprecated:"false"`

	// CdInformer is used to determine whether CD informer is enabled or not
	CdInformer bool `env:"CD_INFORMER" envDefault:"true" description:"Used to determine whether CD informer is enabled or not" deprecated:"false"`
}

func getCdConfig() (*CdConfig, error) {
	cdConfig := &CdConfig{}
	err := env.Parse(cdConfig)
	return cdConfig, err
}
