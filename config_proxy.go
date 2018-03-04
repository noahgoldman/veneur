package veneur

type ProxyConfig struct {
	ConsulRefreshInterval  string `yaml:"consul_refresh_interval"`
	ConsulTraceServiceName string `yaml:"consul_trace_service_name"`
	Debug                  bool   `yaml:"debug"`
	EnableProfiling        bool   `yaml:"enable_profiling"`
	ForwardAddress         string `yaml:"forward_address"`
	ForwardTimeout         string `yaml:"forward_timeout"`
	GrpcAddress            string `yaml:"grpc_address"`
	HTTPAddress            string `yaml:"http_address"`
	SentryDsn              string `yaml:"sentry_dsn"`
	SsfDestinationAddress  string `yaml:"ssf_destination_address"`
	StatsAddress           string `yaml:"stats_address"`
	TraceAddress           string `yaml:"trace_address"`
	TraceAPIAddress        string `yaml:"trace_api_address"`
}
