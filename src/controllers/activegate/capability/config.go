package capability

const (
	ActiveGateContainerName = "activegate"

	ActiveGateVolumeNameGatewayConfig = "ag-lib-gateway-config"
	ActiveGateVolumeNameGatewayTemp   = "ag-lib-gateway-temp"
	ActiveGateVolumeNameGatewayData   = "ag-lib-gateway-data"
	ActiveGateVolumeNameLog           = "ag-log-gateway"
	ActiveGateVolumeNameTmp           = "ag-tmp-gateway"

	ActiveGateDirectoryGatewayConfig = "/var/lib/dynatrace/gateway/config"
	ActiveGateDirectoryGatewayTemp   = "/var/lib/dynatrace/gateway/temp"
	ActiveGateDirectoryGatewayData   = "/var/lib/dynatrace/gateway/data"
	ActiveGateDirectoryLog           = "/var/log/dynatrace/gateway"
	ActiveGateDirectoryTmp           = "/var/tmp/dynatrace/gateway"

	HttpsServicePortName = "https"
	HttpsServicePort     = 443
	HttpServicePortName  = "http"
	HttpServicePort      = 80

	EecContainerName       = ActiveGateContainerName + "-eec"
	StatsdContainerName    = ActiveGateContainerName + "-statsd"
	StatsdIngestPortName   = "statsd"
	StatsdIngestPort       = 18125
	StatsdIngestTargetPort = "statsd-port"
)
