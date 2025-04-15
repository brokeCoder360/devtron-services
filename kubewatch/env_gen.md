

## ARGOCD_INFORMER Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ACD_INFORMER | bool |true | Used to determine whether ArgoCD informer is enabled or not |  | false |
 | ACD_NAMESPACE | string |devtroncd | Namespace where all the ArgoCD application objects are published. For multi-cluster mode, it will be set to v1.NamespaceAll |  | false |


## CD_ARGO_WORKFLOW Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | CD_DEFAULT_NAMESPACE | string |devtron-cd | Namespace where all CD workflows objects are scheduled. For multi-cluster mode, it will be set to v1.NamespaceAll |  | false |
 | CD_INFORMER | bool |true | Used to determine whether CD informer is enabled or not |  | false |


## CI_ARGO_WORKFLOW Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | CI_INFORMER | bool |true | Used to determine whether CI informer is enabled or not |  | false |
 | DEFAULT_NAMESPACE | string |devtron-ci | Namespace where all CI workflows objects are scheduled. For multi-cluster mode, it will be set to v1.NamespaceAll |  | false |


## CLUSTER_MODE Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | CLUSTER_ARGO_CD_TYPE | string |IN_CLUSTER | Determines cluster mode for ArgoCD informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER |  | false |
 | CLUSTER_CD_ARGO_WF_TYPE | string |IN_CLUSTER | Determines cluster mode for CD ArgoWorkflow informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER |  | false |
 | CLUSTER_CI_ARGO_WF_TYPE | string |IN_CLUSTER | Determines cluster mode for CI ArgoWorkflow informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER |  | false |
 | CLUSTER_TYPE | string |IN_CLUSTER | Determines cluster mode for System Executor informer; for multiple cluster mode, it will be set to ALL_CLUSTER; for single cluster mode, it will be set to IN_CLUSTER |  | false |


## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | APP | string |kubewatch |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | ENABLE_STATSVIZ | bool |false |  |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | int |25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | int |300 |  |  | false |
 | K8s_TCP_KEEPALIVE | int |30 |  |  | false |
 | K8s_TCP_TIMEOUT | int |30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | int |10 |  |  | false |
 | LOG_LEVEL | int |-1 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NATS_SERVER_HOST | string |nats://devtron-nats.devtroncd:4222 |  |  | false |
 | PG_ADDR | string |127.0.0.1 |  |  | false |
 | PG_DATABASE | string |orchestrator |  |  | false |
 | PG_EXPORT_PROM_METRICS | bool |true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | bool |true |  |  | false |
 | PG_LOG_ALL_QUERY | bool |false |  |  | false |
 | PG_LOG_SLOW_QUERY | bool |true |  |  | false |
 | PG_PASSWORD | string | |  |  | false |
 | PG_PORT | string |5432 |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | int64 |5000 |  |  | false |
 | PG_USER | string | |  |  | false |
 | RUNTIME_CONFIG_LOCAL_DEV | LocalDevMode |false |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | bool |false |  |  | false |


## EXTERNAL_KUBEWATCH Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | CD_EXTERNAL_LISTENER_URL | string |http://devtroncd-orchestrator-service-prod.devtroncd:80 | URL of the orchestrator |  | false |
 | CD_EXTERNAL_NAMESPACE | string | | Namespace where the external kubewatch is set up |  | false |
 | CD_EXTERNAL_ORCHESTRATOR_TOKEN | string | | Token used to authenticate with the orchestrator |  | false |
 | CD_EXTERNAL_REST_LISTENER | bool |false | Used to determine whether it's an external kubewatch or internal kubewatch |  | false |


## GRACEFUL_SHUTDOWN Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | SLEEP_TIMEOUT | int |5 | Graceful shutdown timeout in seconds |  | false |

