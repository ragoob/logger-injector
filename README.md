# logger injector
## Worker that run into K8s cluster level looking into pods that match some metadata in start inject a sidecar fluentd container to push the app level logs to Elastic search

[![N|Solid](https://www.fluentd.org/images/miscellany/fluentd-logo.png)](https://nodesource.com/products/nsolid)

## Features

- automatic Injector (fluentd sidecar container)
- You do not need to configure anything to your app just write your logs into PVC
- Support multiple worker threads
- One time install per cluster



