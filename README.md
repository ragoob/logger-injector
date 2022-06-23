# logger injector
## Worker that run into K8s cluster level or outside  looking into pods that match some metadata in start inject a sidecar fluentd container to push the app level logs to Elastic search

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
- Now support watching deployment
- support in-cluster configuration or kube config if you need to run it outside K8s cluster

## Todo
- Support other K8s objects
- Add more option to control elastic search and file formats
- Support replication
