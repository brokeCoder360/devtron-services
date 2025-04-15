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

package bean

type ClusterInfo struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	BearerToken string `json:"bearerToken"`
	ServerUrl   string `json:"serverUrl"`
}

const (
	CLUSTER_MODIFY_EVENT_SECRET_TYPE    = "cluster.request/modify"
	CLUSTER_MODIFY_EVENT_FIELD_SELECTOR = "type==cluster.request/modify"
	INFORMER_ALREADY_EXIST_MESSAGE      = "INFORMER_ALREADY_EXIST"
	CLUSTER_ACTION_ADD                  = "add"
	CLUSTER_ACTION_UPDATE               = "update"
	CLUSTER_ACTION_DELETE               = "delete"
	SECRET_FIELD_ACTION                 = "action"
	SECRET_FIELD_CLUSTER_ID             = "cluster_id"

	POD_DELETED_MESSAGE     = "pod deleted"
	EXIT_CODE_143_ERROR     = "Error (exit code 143)"
	CI_WORKFLOW_NAME        = "ci"
	CD_WORKFLOW_NAME        = "cd"
	WORKFLOW_LABEL_SELECTOR = "devtron.ai/purpose==workflow"
	WORKFLOW_TYPE_LABEL_KEY = "workflowType"
	JobKind                 = "Job"
	NodeNoLongerExists      = "PodGC: node no longer exists"
	UPDATE_EVENT            = "update_event"
	DELETE_EVENT            = "delete_event"
)

type ClusterLabels struct {
	ClusterName string
	ClusterId   int
}

func NewClusterLabels(clusterName string, clusterId int) *ClusterLabels {
	return &ClusterLabels{
		ClusterName: clusterName,
		ClusterId:   clusterId,
	}
}
