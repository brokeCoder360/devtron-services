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

package informer

import (
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	"time"
)

type Runner interface {
	Start() error
	Stop()
}

type RunnerImpl struct {
	logger          *zap.SugaredLogger
	appConfig       *config.AppConfig
	k8sUtil         utils.K8sUtil
	clusterInformer *cluster.InformerImpl
}

func NewRunnerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	clusterInformer *cluster.InformerImpl) *RunnerImpl {
	return &RunnerImpl{
		logger:          logger,
		appConfig:       appConfig,
		k8sUtil:         k8sUtil,
		clusterInformer: clusterInformer,
	}
}

func (impl *RunnerImpl) Start() error {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start informer", "time", time.Since(startTime))
	}()
	if impl.appConfig.IsDBAvailable() {
		err := impl.clusterInformer.StartDevtronClusterWatcher()
		if err != nil {
			impl.logger.Errorw("error in starting default cluster informer", "err", err)
			return err
		}
		if err = impl.clusterInformer.StartAll(); err != nil {
			impl.logger.Errorw("error in starting default cluster informer", "err", err)
			return err
		}
	} else {
		err := impl.clusterInformer.StartExternalInformer()
		if err != nil {
			impl.logger.Errorw("error in starting external informer", "err", err)
			return err
		}
	}
	return nil
}

func (impl *RunnerImpl) Stop() {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start default cluster informer", "time", time.Since(startTime))
	}()
	impl.clusterInformer.StopAll()
}
