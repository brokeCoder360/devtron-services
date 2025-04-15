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
	"fmt"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	informerErr "github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"golang.org/x/exp/maps"
)

func (impl *InformerImpl) getCiArgoWfStopper(clusterId int) (*informerBean.SharedStopper, bool) {
	stopper, ok := impl.argoWfCiInformerStopper[clusterId]
	if ok {
		return stopper, stopper.HasInformer()
	}
	return stopper, false
}

func (impl *InformerImpl) getStoppableClusterIds() []int {
	return maps.Keys(impl.argoWfCiInformerStopper)
}

func (impl *InformerImpl) getCiArgoWfStopChannel(clusterLabels *informerBean.ClusterLabels) (chan struct{}, error) {
	stopChannel := make(chan struct{})
	stopper, ok := impl.argoWfCiInformerStopper[clusterLabels.ClusterId]
	if ok && stopper.HasInformer() {
		impl.logger.Debug(fmt.Sprintf("ci argo workflow informer for %s already exist", clusterLabels.ClusterName))
		return stopChannel, informerErr.AlreadyExists
	}
	stopper = stopper.GetStopper(stopChannel)
	impl.argoWfCiInformerStopper[clusterLabels.ClusterId] = stopper
	return stopChannel, nil
}
