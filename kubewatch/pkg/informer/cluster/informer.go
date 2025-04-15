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
	"errors"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoCD"
	cdWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/cd"
	ciWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/ci"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/systemExec"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	resourceBean "github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"time"
)

type Informer interface {
	StartDevtronClusterWatcher() error
	StartExternalInformer() error
	StartAll() error
	StopAll()
}

type InformerImpl struct {
	logger                 *zap.SugaredLogger
	appConfig              *config.AppConfig
	k8sUtil                utils.K8sUtil
	informerClient         resource.InformerClient
	clusterRepository      repository.ClusterRepository
	clusterInformerStopper *informerBean.FactoryStopper
	argoCdInformer         *argoCD.InformerImpl
	ciWfInformer           *ciWf.InformerImpl
	cdWfInformer           *cdWf.InformerImpl
	systemExecInformer     *systemExec.InformerImpl
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	clusterRepository repository.ClusterRepository,
	informerClient resource.InformerClient,
	argoCdInformer *argoCD.InformerImpl,
	ciWfInformer *ciWf.InformerImpl,
	cdWfInformer *cdWf.InformerImpl,
	systemExecInformer *systemExec.InformerImpl) *InformerImpl {
	return &InformerImpl{
		logger:             logger,
		appConfig:          appConfig,
		k8sUtil:            k8sUtil,
		informerClient:     informerClient,
		clusterRepository:  clusterRepository,
		argoCdInformer:     argoCdInformer,
		ciWfInformer:       ciWfInformer,
		cdWfInformer:       cdWfInformer,
		systemExecInformer: systemExecInformer,
	}
}

func (impl *InformerImpl) StartDevtronClusterWatcher() error {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start default cluster informer", "time", time.Since(startTime))
	}()
	clusterInfo, err := impl.clusterRepository.FindByName(commonBean.DEFAULT_CLUSTER)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching default cluster", "err", err, "clusterName", commonBean.DEFAULT_CLUSTER)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("default cluster not found", "clusterName", commonBean.DEFAULT_CLUSTER)
		return err
	}
	impl.logger.Debug("starting informer, reading new cluster request for default cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.FieldSelector = informerBean.CLUSTER_MODIFY_EVENT_FIELD_SELECTOR
	})
	// addFunc is called when a new secret is created
	addFunc := func(secretObject *coreV1.Secret) {
		if err := impl.handleClusterChangeEvent(secretObject); err != nil {
			impl.logger.Errorw("error in handling cluster add event", "err", err, "secretObject", secretObject)
		}
	}
	// updateFunc is called when an existing secret is updated
	updateFunc := func(oldSecretObject, newSecretObject *coreV1.Secret) {
		if err := impl.handleClusterChangeEvent(newSecretObject); err != nil {
			impl.logger.Errorw("error in handling cluster update event", "err", err, "newSecretObject", newSecretObject)
		}
	}
	// deleteFunc is called when an existing secret is deleted
	deleteFunc := func(secretObject *coreV1.Secret) {
		if err := impl.handleClusterDeleteEvent(secretObject); err != nil {
			impl.logger.Errorw("error in handling cluster delete event", "err", err, "secretObject", secretObject)
		}
	}
	informerFactory := impl.informerClient.GetSecretInformerFactory()
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	eventHandler := resourceBean.NewEventHandlers[coreV1.Secret]().
		AddFuncHandler(addFunc).
		UpdateFuncHandler(updateFunc).
		DeleteFuncHandler(deleteFunc)
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	secretFactory, err := informerFactory.GetSharedInformerFactory(restConfig, clusterLabels, eventHandler, labelOptions)
	if err != nil {
		impl.logger.Errorw("error in registering default cluster secret informer", "err", err)
		middleware.IncUnregisteredInformers(clusterLabels, middleware.DEFAULT_CLUSTER_SECRET)
		return err
	}
	stopChannel, err := impl.getStopChannel(secretFactory, clusterLabels)
	if err != nil {
		return err
	}
	secretFactory.Start(stopChannel)
	// waiting for cache sync
	synced := secretFactory.WaitForCacheSync(stopChannel)
	for v, ok := range synced {
		if !ok {
			impl.logger.Errorw("failed to sync secret informer for default cluster", "value", v)
			return errors.New("failed to sync secret informer for default cluster")
		}
	}
	impl.logger.Infow("informer started for cluster watcher", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	return nil
}

func (impl *InformerImpl) StartExternalInformer() error {
	return impl.startInformerForCluster(repository.GetDefaultCluster())
}

func (impl *InformerImpl) StartAll() error {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		impl.logger.Error("error in fetching clusters", "err", err)
		return err
	}
	for _, model := range models {
		if err = impl.startInformerForCluster(model); err != nil {
			impl.logger.Errorw("error in starting informer for cluster", "clusterId", model.Id, "err", err)
			// ignore error and continue with other clusters
		}
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for supportedClient := range informerBean.SupportedClientMap {
		clientAdvisor, err := impl.GetClientStopper(supportedClient)
		if err != nil {
			impl.logger.Errorw("error in getting client advisor", "supportedClient", supportedClient, "err", err)
			continue
		}
		clientAdvisor.StopAll()
	}
	if err := impl.stopDevtronClusterWatcher(); err != nil {
		impl.logger.Errorw("error in stopping informer for default cluster", "err", err)
	}
}
