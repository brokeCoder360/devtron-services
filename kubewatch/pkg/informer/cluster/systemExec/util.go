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
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func getTopic(workflowType string) (string, error) {
	switch workflowType {
	case informerBean.CD_WORKFLOW_NAME:
		return pubsub.CD_WORKFLOW_STATUS_UPDATE, nil
	case informerBean.CI_WORKFLOW_NAME:
		return pubsub.WORKFLOW_STATUS_UPDATE_TOPIC, nil
	}
	return "", fmt.Errorf("no topic mapped to workflow type %s", workflowType)
}

func getLatestFinishedAt(pod *coreV1.Pod) metav1.Time {
	var latest metav1.Time
	for _, ctr := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		if r := ctr.State.Running; r != nil { // if we are running, then the finished at time must be now or after
			latest = metav1.Now()
		} else if t := ctr.State.Terminated; t != nil && t.FinishedAt.After(latest.Time) {
			latest = t.FinishedAt
		}
	}
	return latest
}

func getExitCode(pod *coreV1.Pod) *int32 {
	for _, c := range pod.Status.ContainerStatuses {
		if c.Name == common.MainContainerName && c.State.Terminated != nil {
			return pointer.Int32Ptr(c.State.Terminated.ExitCode)
		}
	}
	return nil
}

func getPendingReason(pod *coreV1.Pod) string {
	for _, ctrStatus := range pod.Status.ContainerStatuses {
		if ctrStatus.State.Waiting != nil {
			if ctrStatus.State.Waiting.Message != "" {
				return fmt.Sprintf("%s: %s", ctrStatus.State.Waiting.Reason, ctrStatus.State.Waiting.Message)
			}
			return ctrStatus.State.Waiting.Reason
		}
	}
	// Example:
	// - lastProbeTime: null
	//   lastTransitionTime: 2018-08-29T06:38:36Z
	//   message: '0/3 nodes are available: 2 Insufficient cpu, 3 MatchNodeSelector.'
	//   reason: Unschedulable
	//   status: "False"
	//   type: PodScheduled
	for _, cond := range pod.Status.Conditions {
		if cond.Reason == coreV1.PodReasonUnschedulable {
			if cond.Message != "" {
				return fmt.Sprintf("%s: %s", cond.Reason, cond.Message)
			}
			return cond.Reason
		}
	}
	return ""
}

func isResourceNotFoundErr(err error) bool {
	if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == metav1.StatusReasonNotFound {
		return true
	}
	return false
}

func getWorkflowStatus(podObj *coreV1.Pod, nodeStatus v1alpha1.NodeStatus, templateName string) *v1alpha1.WorkflowStatus {
	workflowStatus := &v1alpha1.WorkflowStatus{}
	workflowPhase := v1alpha1.WorkflowPhase(nodeStatus.Phase)
	if workflowPhase == v1alpha1.WorkflowPending {
		workflowPhase = v1alpha1.WorkflowRunning
	}
	if workflowPhase.Completed() {
		workflowStatus.FinishedAt = nodeStatus.FinishedAt
	}
	workflowStatus.Phase = workflowPhase
	nodeNameVsStatus := make(map[string]v1alpha1.NodeStatus, 1)
	nodeStatus.ID = podObj.Name
	nodeStatus.TemplateName = templateName
	nodeStatus.Name = nodeStatus.ID
	nodeStatus.BoundaryID = getPodOwnerNameByKind(podObj, informerBean.JobKind)
	nodeNameVsStatus[podObj.Name] = nodeStatus
	workflowStatus.Nodes = nodeNameVsStatus
	workflowStatus.Message = nodeStatus.Message
	return workflowStatus
}

func getPodOwnerNameByKind(podObj *coreV1.Pod, ownerKind string) string {
	ownerReferences := podObj.OwnerReferences
	for _, ownerReference := range ownerReferences {
		if ownerReference.Kind == ownerKind {
			return ownerReference.Name
		}
	}
	return podObj.Name
}
