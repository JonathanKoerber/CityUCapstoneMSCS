package ics_node

type DeviceConfig struct {
	ImageName      string
	IP             string
	Port           string // "502"
	Net            string // "tcp"
	BridgeName     string
	Protocol       string
	DeviceName     string
	ContextDir     string
	DockerfilePath string
	Dockerfile     string
	ContainerName  string
	Context        map[string]interface{}
}
