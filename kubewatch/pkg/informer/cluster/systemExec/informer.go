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

package systemExec

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	resourceBean "github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"log"
	"time"
)

type InformerImpl struct {
	logger                  *zap.SugaredLogger
	appConfig               *config.AppConfig
	k8sUtil                 utils.K8sUtil
	pubSubClient            *pubsub.PubSubClientServiceImpl
	informerClient          resource.InformerClient
	systemWfInformerStopper map[int]*informerBean.FactoryStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	informerClient resource.InformerClient) *InformerImpl {
	return &InformerImpl{
		logger:                  logger,
		appConfig:               appConfig,
		k8sUtil:                 k8sUtil,
		pubSubClient:            pubSubClient,
		informerClient:          informerClient,
		systemWfInformerStopper: make(map[int]*informerBean.FactoryStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if impl.appConfig.GetExternalConfig().External {
		impl.logger.Debugw("argo workflow setup is not done for cluster, skipping...", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName, "appConfig", impl.appConfig)
		return nil
	}
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start system executor informer", "clusterId", clusterInfo.Id, "time", time.Since(startTime))
	}()
	impl.logger.Infow("starting system executor informer for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = informerBean.WORKFLOW_LABEL_SELECTOR
	})
	// updateFunc is called when an existing pod is updated
	updateFunc := func(oldPodObj, newPodObj *coreV1.Pod) {
		// atleast one of the pod version will be not nil
		if !foundAnyUpdateInPodStatus(oldPodObj, newPodObj) {
			podName := newPodObj.Name
			logArgs := []interface{}{"podName", podName, "newPodStatusObj", newPodObj.Status}
			if oldPodObj != nil {
				logArgs = append(logArgs, "oldPodStatusObj", oldPodObj.Status)
			}
			impl.logger.Debugw("no significant pod updates are detected so skipping the pod update event", logArgs...)
			return
		}

		if newPodObj != nil {
			var workflowType string
			if newPodObj.Labels != nil {
				if val, ok := newPodObj.Labels[informerBean.WORKFLOW_TYPE_LABEL_KEY]; ok {
					workflowType = val
				}
			}
			impl.logger.Debugw("event received in pods update informer", "time", time.Now(), "podObjStatus", newPodObj.Status)
			nodeStatus := impl.assessNodeStatus(informerBean.UPDATE_EVENT, newPodObj)
			workflowStatus := getWorkflowStatus(newPodObj, nodeStatus, workflowType)
			if workflowStatus.Message == "" && workflowStatus.Phase == v1alpha1.WorkflowFailed {
				impl.logger.Debugw("skipping the failed workflow update event as message is empty", "workflow", workflowStatus)
				return
			}
			wfJson, err := json.Marshal(workflowStatus)
			if err != nil {
				impl.logger.Errorw("error occurred while marshalling workflowJson", "err", err)
				return
			}
			impl.logger.Debugw("sending system executor workflow update event", "workflow", string(wfJson))
			if impl.pubSubClient == nil {
				log.Println("don't publish")
				return
			}
			topic, err := getTopic(workflowType)
			if err != nil {
				impl.logger.Errorw("error while getting Topic")
				return
			}
			err = impl.pubSubClient.Publish(topic, string(wfJson))
			if err != nil {
				impl.logger.Errorw("error while publishing Request", "err", err)
				return
			}
			impl.logger.Debugw("system executor workflow update sent", "workflowType", workflowType)
		}
	}
	// deleteFunc is called when an existing pod is deleted
	deleteFunc := func(podObj *coreV1.Pod) {
		var workflowType string
		if podObj.Labels != nil {
			if val, ok := podObj.Labels[informerBean.WORKFLOW_TYPE_LABEL_KEY]; ok {
				workflowType = val
			}
		}
		impl.logger.Debugw("event received in Pods delete informer", "time", time.Now(), "podObjStatus", podObj.Status)
		nodeStatus := impl.assessNodeStatus(informerBean.DELETE_EVENT, podObj)
		nodeStatus, reTriggerRequired := impl.checkIfPodDeletedAndUpdateMessage(podObj.Name, podObj.Namespace, nodeStatus, restConfig)
		if !reTriggerRequired {
			//not sending this deleted event if it's not a re-trigger case
			return
		}
		workflowStatus := getWorkflowStatus(podObj, nodeStatus, workflowType)
		wfJson, err := json.Marshal(workflowStatus)
		if err != nil {
			impl.logger.Errorw("error occurred while marshalling workflowJson", "err", err)
			return
		}
		impl.logger.Debugw("sending system executor cd workflow delete event", "workflow", string(wfJson))
		if impl.pubSubClient == nil {
			log.Println("don't publish")
			return
		}
		topic, err := getTopic(workflowType)
		if err != nil {
			impl.logger.Errorw("error while getting topic")
			return
		}
		err = impl.pubSubClient.Publish(topic, string(wfJson))
		if err != nil {
			impl.logger.Errorw("error while publishing request", "err", err)
			return
		}
		impl.logger.Debugw("workflow update sent", "workflowType", workflowType)
	}
	podInformerFactory := impl.informerClient.GetPodInformerFactory()
	eventHandler := resourceBean.NewEventHandlers[coreV1.Pod]().
		UpdateFuncHandler(updateFunc).
		DeleteFuncHandler(deleteFunc)
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	podFactory, err := podInformerFactory.GetSharedInformerFactory(restConfig, clusterLabels, eventHandler, labelOptions)
	if err != nil {
		impl.logger.Errorw("error in adding event handler for cluster pod informer", "err", err)
		middleware.IncUnregisteredInformers(clusterLabels, middleware.SYSTEM_EXECUTOR)
		return err
	}
	stopChannel, err := impl.getStopChannel(podFactory, clusterLabels)
	if err != nil {
		return err
	}
	podFactory.Start(stopChannel)
	impl.logger.Infow("informer started for system executor", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getSystemWfStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.systemWfInformerStopper, clusterId)
		impl.logger.Infow("system executor informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, clusterId := range impl.getStoppableClusterIds() {
		if err := impl.StopInformerForCluster(clusterId); err != nil {
			impl.logger.Errorw("error in stopping informer for cluster", "clusterId", clusterId, "err", err)
			// continue stopping other informers
		}
	}
}
