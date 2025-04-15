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

package argoCD

import (
	"fmt"
	"github.com/devtron-labs/common-lib/async"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	informerErr "github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	resourceBean "github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	"time"
)

type InformerImpl struct {
	logger                *zap.SugaredLogger
	appConfig             *config.AppConfig
	k8sUtil               utils.K8sUtil
	informerClient        resource.InformerClient
	argoCdInformerStopper map[int]*informerBean.SharedStopper
	asyncRunnable         *async.Runnable
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                logger,
		appConfig:             appConfig,
		k8sUtil:               k8sUtil,
		informerClient:        informerClient,
		asyncRunnable:         asyncRunnable,
		argoCdInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if !impl.appConfig.GetAcdConfig().ACDInformer || impl.appConfig.GetExternalConfig().External {
		impl.logger.Debugw("argo cd informer is not enabled for cluster, skipping...", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName, "appConfig", impl.appConfig)
		return nil
	}
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start argo cd informer", "clusterId", clusterInfo.Id, "time", time.Since(startTime))
	}()
	stopper, ok := impl.argoCdInformerStopper[clusterInfo.Id]
	if ok && stopper.HasInformer() {
		impl.logger.Debug(fmt.Sprintf("argo cd application informer for %s already exist", clusterInfo.ClusterName))
		return informerErr.AlreadyExists
	}
	impl.logger.Infow("starting informer for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	applicationInformer := impl.informerClient.GetSharedInformerClient(resourceBean.ApplicationResourceType)
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	acdInformer, err := applicationInformer.GetSharedInformer(clusterInfo.Id, impl.appConfig.GetACDNamespace(), restConfig)
	if err != nil {
		impl.logger.Errorw("error in registering acd informer", "err", err, "clusterId", clusterInfo.Id)
		middleware.IncUnregisteredInformers(clusterLabels, middleware.ARGO_CD)
		return err
	}
	stopChannel, err := impl.getStopChannel(clusterLabels)
	if err != nil {
		return err
	}
	runnable := func() {
		acdInformer.Run(stopChannel)
		impl.logger.Infow("informer started for argocd", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	}
	impl.asyncRunnable.Execute(runnable)
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getArgoCdStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.argoCdInformerStopper, clusterId)
		impl.logger.Infow("argo cd informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, clusterId := range impl.getStoppableClusterIds() {
		if err := impl.StopInformerForCluster(clusterId); err != nil {
			impl.logger.Errorw("error in stopping argo cd informer", "clusterId", clusterId, "err", err)
		}
	}
}
