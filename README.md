# logger injector
## Worker that run into K8s cluster level or outside  looking into pods that match some metadata and inject a sidecar fluentd container to push the app level logs to Elastic search

# Design 
[![N|Solid](https://raw.githubusercontent.com/ragoob/logger-injector/develop/Injector.png)](#)

# How it works
- The injector deamon run on cluster level and watch all objects changes and if it contains special annotations the inector start modify the deployment or stateful 
  pod to add fluentd side car containers configured to look into the pod logs volume and stream the logs to elastic search 
  
## Features

- automatic Injector (fluentd sidecar container)
- You do not need to configure anything to your app just write your logs into volume
- Support multiple worker threads
- One time install per cluster
- Now support watching deployments only
- support in-cluster configuration or kube config if you need to run it outside K8s cluster
## Tech Stack
- Go

## Todo
- Support other K8s objects
- Add more option to control elastic search and file formats
- Support replication



## Injector Configurations
  | Variable       | Type         |Description| 
| :------------- |:-------------| :-----|
| ELASTIC_HOST   | string       | elastic search host 
| ELASTIC_PORT    | number        |  elastic search port  
| ELASTIC_PASSWORD    | number        |    elastic user password
| ELASTIC_USER    | string        |     elastic user   |
| ELASTIC_SSL_VERIFY    | boolean       |    elastic skip ssl verify  default false |
| ELASTIC_SCHEME    | string       |    elastic http/https  default https  |
| ELASTIC_SSL_VERSION    |string       |    elastic tls version  default TLSv1_2 |
| FLUENTD_IMAGE_REPOSITORY    |string       |  fluentd image default fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch |
| IN_CLUSTER_CONFIG    |boolean       |  Set it true if the app will run inside the cluster  |
| KUBE_CONFIG_PATH    |boolean       |  working when IN_CLUSTER_CONFIG is false and the default value is user home dir |

## App required annotations 
| name | type | Description |
| --------------- | --------------- | --------------- |
| logger.injector.io/agent-inject | boolean  | Required to be true |
| logger.injector.io/log-tag-name | string | Elastic search tag for create index |
| logger.injector.io/flush-interval | string | Fluentd flush interval default 1m |
| logger.injector.io/flush-interval | string | Fluentd flush interval default 1m |
| logger.injector.io/log-path-pattern | string | Your log file pattern such as log*.txt default log*.log |
| logger.injector.io/storage-class-name | string | Storage class Name to create PVC for fluentd buffer default emptyDir{} |
| logger.injector.io/fluentd-vol-size | string | Volume storage for fluentd buffer PV default 1 Gi |

## How to install
- Open build directory and copy default.properties to the build dir and populate your environment variables value
- run ``` deploy.sh ```
