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

package workflow

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type InformerImpl struct {
	logger        *zap.SugaredLogger
	client        *pubsub.PubSubClientServiceImpl
	appConfig     *config.AppConfig
	natsTopicName string
}

func NewCiInformerImpl(logger *zap.SugaredLogger, client *pubsub.PubSubClientServiceImpl, appConfig *config.AppConfig) *InformerImpl {
	return &InformerImpl{
		logger:        logger,
		client:        client,
		appConfig:     appConfig,
		natsTopicName: pubsub.WORKFLOW_STATUS_UPDATE_TOPIC,
	}
}

func NewCdInformerImpl(logger *zap.SugaredLogger, client *pubsub.PubSubClientServiceImpl, appConfig *config.AppConfig) *InformerImpl {
	return &InformerImpl{
		logger:        logger,
		client:        client,
		appConfig:     appConfig,
		natsTopicName: pubsub.CD_WORKFLOW_STATUS_UPDATE,
	}
}

func (impl *InformerImpl) GetSharedInformer(clusterId int, namespace string, k8sConfig *rest.Config) (cache.SharedIndexInformer, error) {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("registered workflow informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	httpClient, err := rest.HTTPClientFor(k8sConfig)
	if err != nil {
		impl.logger.Error("error in getting http client for the default k8s config")
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfigAndClient(k8sConfig, httpClient)
	if err != nil {
		impl.logger.Errorw("error in getting dynamic interface for resource", "err", err)
		return nil, err
	}
	workflowInformer := util.NewWorkflowInformer(dynamicClient, namespace, 0, nil, cache.Indexers{})
	_, err = workflowInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldWf, newWf interface{}) {
			impl.logger.Info("workflow update detected")
			if workflow, ok := newWf.(*unstructured.Unstructured).Object["status"]; ok {
				wfJson, err := json.Marshal(workflow)
				if err != nil {
					impl.logger.Errorw("error occurred while marshalling workflow", "err", err)
					return
				}
				impl.logger.Debugw("sending workflow update event ", "wfJson", string(wfJson))
				var reqBody = wfJson
				if impl.appConfig.GetExternalConfig().External {
					err = publishEventsOnRest(reqBody, impl.natsTopicName, impl.appConfig.GetExternalConfig())
				} else {
					if impl.client == nil {
						impl.logger.Warn("don't publish")
						return
					}
					err = impl.client.Publish(impl.natsTopicName, string(reqBody))
				}
				if err != nil {
					impl.logger.Errorw("Error while publishing Request", "err ", err)
					return
				}
				impl.logger.Debug("workflow update sent")
			}
		},
		DeleteFunc: func(wf interface{}) {},
	})
	if err != nil {
		impl.logger.Errorw("error in adding workflow event handler", "err", err)
		return workflowInformer, err
	}
	return workflowInformer, nil
}
