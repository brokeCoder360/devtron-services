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

package argoWf

import (
	"github.com/devtron-labs/common-lib/async"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	resourceBean "github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	"time"
)

type InformerImpl struct {
	logger                  *zap.SugaredLogger
	appConfig               *config.AppConfig
	k8sUtil                 utils.K8sUtil
	informerClient          resource.InformerClient
	asyncRunnable           *async.Runnable
	argoWfCdInformerStopper map[int]*informerBean.SharedStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                  logger,
		appConfig:               appConfig,
		k8sUtil:                 k8sUtil,
		informerClient:          informerClient,
		asyncRunnable:           asyncRunnable,
		argoWfCdInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if !impl.appConfig.GetCdConfig().CdInformer {
		impl.logger.Debugw("cd argo workflow informer is not enabled, skipping...", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName, "appConfig", impl.appConfig)
		return nil
	}
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start cd argo workflow informer", "clusterId", clusterInfo.Id, "time", time.Since(startTime))
	}()
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	cdWfInformer := impl.informerClient.GetSharedInformerClient(resourceBean.CdWorkflowResourceType)
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	workflowInformer, err := cdWfInformer.GetSharedInformer(clusterInfo.Id, impl.appConfig.GetCdWfNamespace(), restConfig)
	if err != nil {
		impl.logger.Errorw("error in starting workflow informer", "err", err)
		middleware.IncUnregisteredInformers(clusterLabels, middleware.CD_STAGE_ARGO_WORLFLOW)
	}
	stopChan, err := impl.getCdArgoWfStopChannel(clusterLabels)
	if err != nil {
		return err
	}
	runnable := func() {
		workflowInformer.Run(stopChan)
		impl.logger.Infow("informer started for cd argo workflow", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	}
	impl.asyncRunnable.Execute(runnable)
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getCdArgoWfStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.argoWfCdInformerStopper, clusterId)
		impl.logger.Infow("argo workflow cd informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, clusterId := range impl.getStoppableClusterIds() {
		if err := impl.StopInformerForCluster(clusterId); err != nil {
			impl.logger.Errorw("error in stopping argo workflow ci informer", "clusterId", clusterId, "err", err)
		}
	}
}
