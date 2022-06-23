package loggerInjector

type Elastic struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	SslVersion string `json:"ssl_version"`
	Scheme     string `json:"scheme"`
	SslVerify  bool   `json:"ssl_verify"`
}
