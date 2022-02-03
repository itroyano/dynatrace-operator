package capability

import (
	"path/filepath"
	"strings"

	dynatracev1beta2 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

const (
	MultiActiveGateName       = "activegate"
	trustStoreVolume          = "truststore-volume"
	k8scrt2jksPath            = "/opt/dynatrace/gateway/k8scrt2jks.sh"
	activeGateCacertsPath     = "/opt/dynatrace/gateway/jre/lib/security/cacerts"
	activeGateSslPath         = "/var/lib/dynatrace/gateway/ssl"
	k8sCertificateFile        = "k8s-local.jks"
	k8scrt2jksWorkingDir      = "/var/lib/dynatrace/gateway"
	initContainerTemplateName = "certificate-loader"

	jettyCerts = "server-certs"

	secretsRootDir = "/var/lib/dynatrace/secrets/"
)

type baseFunc func() *capabilityBase

var activeGateCapabilities = map[dynatracev1beta2.CapabilityDisplayName]baseFunc{
	dynatracev1beta2.KubeMonCapability.DisplayName:       kubeMonBase,
	dynatracev1beta2.RoutingCapability.DisplayName:       routingBase,
	dynatracev1beta2.MetricsIngestCapability.DisplayName: metricsIngestBase,
	dynatracev1beta2.DynatraceApiCapability.DisplayName:  dynatraceApiBase,
}

type Configuration struct {
	SetDnsEntryPoint     bool
	SetReadinessPort     bool
	SetCommunicationPort bool
	CreateService        bool
	ServiceAccountOwner  string
}

type Capability interface {
	Enabled() bool
	ShortName() string
	ArgName() string
	Properties() *dynatracev1beta2.ActiveGateProperties
	Config() Configuration
	InitContainersTemplates() []corev1.Container
	ContainerVolumeMounts() []corev1.VolumeMount
	Volumes() []corev1.Volume
}

type capabilityBase struct {
	enabled    bool
	shortName  string
	argName    string
	properties *dynatracev1beta2.ActiveGateProperties
	Configuration
	initContainersTemplates []corev1.Container
	containerVolumeMounts   []corev1.VolumeMount
	volumes                 []corev1.Volume
}

func (c *capabilityBase) Enabled() bool {
	return c.enabled
}

func (c *capabilityBase) Properties() *dynatracev1beta2.ActiveGateProperties {
	return c.properties
}

func (c *capabilityBase) Config() Configuration {
	return c.Configuration
}

func (c *capabilityBase) ShortName() string {
	return c.shortName
}

func (c *capabilityBase) ArgName() string {
	return c.argName
}

// Note:
// Caller must set following fields:
//   Image:
//   Resources:
func (c *capabilityBase) InitContainersTemplates() []corev1.Container {
	return c.initContainersTemplates
}

func (c *capabilityBase) ContainerVolumeMounts() []corev1.VolumeMount {
	return c.containerVolumeMounts
}

func (c *capabilityBase) Volumes() []corev1.Volume {
	return c.volumes
}

func CalculateStatefulSetName(capability Capability, instanceName string) string {
	return instanceName + "-" + capability.ShortName()
}

type MultiCapability struct {
	capabilityBase
}

func (c *capabilityBase) setTlsConfig(agSpec *dynatracev1beta2.ActiveGateSpec) {
	if agSpec == nil {
		return
	}

	if agSpec.TlsSecretName != "" {
		c.volumes = append(c.volumes,
			corev1.Volume{
				Name: jettyCerts,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: agSpec.TlsSecretName,
					},
				},
			})
		c.containerVolumeMounts = append(c.containerVolumeMounts,
			corev1.VolumeMount{
				ReadOnly:  true,
				Name:      jettyCerts,
				MountPath: filepath.Join(secretsRootDir, "tls"),
			})
	}
}

func NewMultiCapability(activeGate *dynatracev1beta2.ActiveGateSpec) *MultiCapability {
	mc := MultiCapability{
		capabilityBase{
			shortName: MultiActiveGateName,
		},
	}
	if activeGate == nil || len(activeGate.Capabilities) == 0 {
		mc.CreateService = true // necessary for cleaning up service if created
		return &mc
	}
	mc.enabled = true
	mc.properties = &activeGate.ActiveGateProperties
	capabilityNames := []string{}
	for capName := range activeGate.Capabilities {
		capabilityGenerator, ok := activeGateCapabilities[capName]
		if !ok {
			continue
		}
		capGen := capabilityGenerator()

		capabilityNames = append(capabilityNames, capGen.argName)
		mc.initContainersTemplates = append(mc.initContainersTemplates, capGen.initContainersTemplates...)
		mc.containerVolumeMounts = append(mc.containerVolumeMounts, capGen.containerVolumeMounts...)
		mc.volumes = append(mc.volumes, capGen.volumes...)

		if !mc.CreateService {
			mc.CreateService = capGen.CreateService
		}
		if !mc.SetCommunicationPort {
			mc.SetCommunicationPort = capGen.SetCommunicationPort
		}
		if !mc.SetDnsEntryPoint {
			mc.SetDnsEntryPoint = capGen.SetDnsEntryPoint
		}
		if !mc.SetReadinessPort {
			mc.SetReadinessPort = capGen.SetReadinessPort
		}
		if mc.ServiceAccountOwner == "" {
			mc.ServiceAccountOwner = capGen.ServiceAccountOwner
		}
	}
	mc.argName = strings.Join(capabilityNames, ",")
	mc.setTlsConfig(activeGate)
	return &mc

}

func kubeMonBase() *capabilityBase {
	c := capabilityBase{
		shortName: dynatracev1beta2.KubeMonCapability.ShortName,
		argName:   dynatracev1beta2.KubeMonCapability.ArgumentName,
		Configuration: Configuration{
			ServiceAccountOwner: "kubernetes-monitoring",
		},
		initContainersTemplates: []corev1.Container{
			{
				Name:            initContainerTemplateName,
				ImagePullPolicy: corev1.PullAlways,
				WorkingDir:      k8scrt2jksWorkingDir,
				Command:         []string{"/bin/bash"},
				Args:            []string{"-c", k8scrt2jksPath},
				VolumeMounts: []corev1.VolumeMount{
					{
						ReadOnly:  false,
						Name:      trustStoreVolume,
						MountPath: activeGateSslPath,
					},
				},
			},
		},
		containerVolumeMounts: []corev1.VolumeMount{{
			ReadOnly:  true,
			Name:      trustStoreVolume,
			MountPath: activeGateCacertsPath,
			SubPath:   k8sCertificateFile,
		}},
		volumes: []corev1.Volume{{
			Name: trustStoreVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}},
		},
	}
	return &c
}

func routingBase() *capabilityBase {
	c := capabilityBase{
		shortName: dynatracev1beta2.RoutingCapability.ShortName,
		argName:   dynatracev1beta2.RoutingCapability.ArgumentName,
		Configuration: Configuration{
			SetDnsEntryPoint:     true,
			SetReadinessPort:     true,
			SetCommunicationPort: true,
			CreateService:        true,
		},
	}
	return &c
}

func metricsIngestBase() *capabilityBase {
	c := capabilityBase{
		shortName: dynatracev1beta2.MetricsIngestCapability.ShortName,
		argName:   dynatracev1beta2.MetricsIngestCapability.ArgumentName,
		Configuration: Configuration{
			SetDnsEntryPoint:     true,
			SetReadinessPort:     true,
			SetCommunicationPort: true,
			CreateService:        true,
		},
	}
	return &c
}

func dynatraceApiBase() *capabilityBase {
	c := capabilityBase{
		shortName: dynatracev1beta2.DynatraceApiCapability.ShortName,
		argName:   dynatracev1beta2.DynatraceApiCapability.ArgumentName,
		Configuration: Configuration{
			SetDnsEntryPoint:     true,
			SetReadinessPort:     true,
			SetCommunicationPort: true,
			CreateService:        true,
		},
	}
	return &c
}
