package loggerInjector

const (
	InjectorAgentAnnotation    = "logger.injector.io/agent-inject"
	InjectorInjectedAnnotation = "logger.injector.io/injected"
	InjectorElastic            = "logger-side-car-elastic"
	InjectorClaimName          = "logger.injector.io/app-claim-name"
	InjectorLogTag             = "logger.injector.io/log-tag-name"
	InjectorFlushInterval      = "logger.injector.io/flush-interval"
	InjectorLogPathPattern     = "logger.injector.io/log-path-pattern"
	InjectorStorageClassName   = "logger.injector.io/storage-class-name"
	FluentDConfigData          = "fluent.conf"
	FluentdImageRepository     = "logger-fluentd-image-repository"
	FluentdVolumeSize          = "logger.injector.io/fluentd-vol-size"
	KubeConfigPathEnv          = "KUBE_CONFIG_PATH"
	InClusterConfig            = "IN_CLUSTER_CONFIG"
)
