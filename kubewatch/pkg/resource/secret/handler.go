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

package secret

import (
	bean2 "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type InformerFactoryImpl struct {
	logger  *zap.SugaredLogger
	k8sUtil utils.K8sUtil
}

func NewInformerFactoryImpl(logger *zap.SugaredLogger,
	k8sUtil utils.K8sUtil) *InformerFactoryImpl {
	return &InformerFactoryImpl{
		logger:  logger,
		k8sUtil: k8sUtil,
	}
}

func (impl *InformerFactoryImpl) GetSharedInformerFactory(config *rest.Config, clusterLabels *bean2.ClusterLabels,
	eventHandlers *bean.EventHandlers[coreV1.Secret], options ...kubeinformers.SharedInformerOption) (kubeinformers.SharedInformerFactory, error) {
	clusterClient, k8sErr := impl.k8sUtil.GetK8sClientForConfig(config)
	if k8sErr != nil {
		middleware.IncUnreachableCluster(clusterLabels)
		return nil, k8sErr
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, 15*time.Minute, options...)
	secretInformer := informerFactory.Core().V1().Secrets()
	_, eventErr := secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			impl.logger.Debugw("event received in cluster secret add informer", "time", time.Now())
			if secretObject, ok := newObj.(*coreV1.Secret); ok {
				eventHandlers.AddFunc(secretObject)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Debugw("event received in cluster secret update informer", "time", time.Now())
			oldSecretObject, validOld := oldObj.(*coreV1.Secret)
			newSecretObject, validNew := newObj.(*coreV1.Secret)
			if validOld && validNew {
				eventHandlers.UpdateFunc(oldSecretObject, newSecretObject)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("event received in secret delete informer", "time", time.Now())
			if secretObject, ok := obj.(*coreV1.Secret); ok {
				eventHandlers.DeleteFunc(secretObject)
			}
		},
	})
	if eventErr != nil {
		impl.logger.Errorw("error in adding event handler for cluster secret informer", "err", eventErr)
		return informerFactory, eventErr
	}
	return informerFactory, nil
}
