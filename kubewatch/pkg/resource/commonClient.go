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

package resource

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
)

type InformerClient interface {
	GetSharedInformerClient(sharedInformerType bean.SharedInformerType) SharedInformer
	GetSecretInformerFactory() InformerFactory[coreV1.Secret]
	GetPodInformerFactory() InformerFactory[coreV1.Pod]
}

type InformerClientImpl struct {
	logger *zap.SugaredLogger
	// client is the pubsub-client to publish events
	// NOTE: pubsub_lib.PubSubClientServiceImpl can be nil for External mode kubewatch
	client    *pubsub.PubSubClientServiceImpl
	appConfig *config.AppConfig
	k8sUtil   utils.K8sUtil
}

func NewInformerClientImpl(logger *zap.SugaredLogger,
	client *pubsub.PubSubClientServiceImpl,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil) *InformerClientImpl {
	return &InformerClientImpl{
		logger:    logger,
		client:    client,
		appConfig: appConfig,
		k8sUtil:   k8sUtil,
	}
}
