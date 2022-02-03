package daemonset

import (
	"testing"

	dynatracev1beta2 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareVolumes(t *testing.T) {
	t.Run(`has root volume`, func(t *testing.T) {
		instance := &dynatracev1beta2.DynaKube{}
		volumes := prepareVolumes(instance)

		assert.Contains(t, volumes, getRootVolume())
		assert.NotContains(t, volumes, getCertificateVolume(instance))
	})
	t.Run(`has certificate volume`, func(t *testing.T) {
		instance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				TrustedCAs: testName,
			},
		}
		volumes := prepareVolumes(instance)

		assert.Contains(t, volumes, getRootVolume())
		assert.Contains(t, volumes, getCertificateVolume(instance))
	})
	t.Run(`has tls volume`, func(t *testing.T) {
		instance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				TrustedCAs: testName,
				ActiveGates: []dynatracev1beta2.ActiveGateSpec{
					{
						Capabilities: map[dynatracev1beta2.CapabilityDisplayName]dynatracev1beta2.CapabilityProperties{
							dynatracev1beta2.KubeMonCapability.DisplayName: {},
						},
						ActiveGateProperties: dynatracev1beta2.ActiveGateProperties{
							TlsSecretName: "testing",
						},
					},
				},
			},
		}
		volumes := prepareVolumes(instance)
		assert.Contains(t, volumes, getTLSVolume(instance))
	})
	t.Run(`has all volumes`, func(t *testing.T) {
		instance := &dynatracev1beta2.DynaKube{
			Spec: dynatracev1beta2.DynaKubeSpec{
				TrustedCAs: testName,
				OneAgent: dynatracev1beta2.OneAgentSpec{
					HostMonitoring: &dynatracev1beta2.HostMonitoringSpec{},
				},
				ActiveGates: []dynatracev1beta2.ActiveGateSpec{
					{
						Capabilities: map[dynatracev1beta2.CapabilityDisplayName]dynatracev1beta2.CapabilityProperties{
							dynatracev1beta2.KubeMonCapability.DisplayName: {},
						},
						ActiveGateProperties: dynatracev1beta2.ActiveGateProperties{
							TlsSecretName: "testing",
						},
					},
				},
			},
		}
		dsInfo := HostMonitoring{
			builderInfo{
				instance:       instance,
				hostInjectSpec: &instance.Spec.OneAgent.HostMonitoring.HostInjectSpec,
				clusterId:      "",
			},
			HostMonitoringFeature,
		}
		ds, err := dsInfo.BuildDaemonSet()
		require.NoError(t, err)

		volumes := ds.Spec.Template.Spec.Volumes

		assert.Contains(t, volumes, getRootVolume())
		assert.Contains(t, volumes, getCertificateVolume(instance))
		assert.Contains(t, volumes, getTLSVolume(instance))
	})
}
