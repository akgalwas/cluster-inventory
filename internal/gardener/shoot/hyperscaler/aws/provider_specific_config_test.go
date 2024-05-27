package aws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestControlPlaneConfig(t *testing.T) {
	t.Run("Create Control Plane config", func(t *testing.T) {
		// when
		controlPlaneConfigBytes, err := GetControlPlaneConfig()

		// then
		require.NoError(t, err)

		var controlPlaneConfig ControlPlaneConfig
		err = json.Unmarshal(controlPlaneConfigBytes, &controlPlaneConfig)
		assert.NoError(t, err)

		assert.Equal(t, apiVersion, controlPlaneConfig.TypeMeta.APIVersion)
		assert.Equal(t, controlPlaneConfigKind, controlPlaneConfig.TypeMeta.Kind)
	})
}

func TestInfrastructureConfig(t *testing.T) {
	for tname, tcase := range map[string]struct {
		givenNodesCidr   string
		givenZoneNames   []string
		expectedAwsZones []Zone
	}{
		"Regular 10.250.0.0/16": {
			givenNodesCidr: "10.250.0.0/16",
			givenZoneNames: []string{
				"eu-central-1a",
				"eu-central-1b",
				"eu-central-1c",
			},
			expectedAwsZones: []Zone{
				{
					Name:     "eu-central-1a",
					Workers:  "10.250.0.0/19",
					Public:   "10.250.32.0/20",
					Internal: "10.250.48.0/20",
				},
				{
					Name:     "eu-central-1b",
					Workers:  "10.250.64.0/19",
					Public:   "10.250.96.0/20",
					Internal: "10.250.112.0/20",
				},
				{
					Name:     "eu-central-1c",
					Workers:  "10.250.128.0/19",
					Public:   "10.250.160.0/20",
					Internal: "10.250.176.0/20",
				},
			},
		},
		"Regular 10.180.0.0/23": {
			givenNodesCidr: "10.180.0.0/23",
			givenZoneNames: []string{
				"eu-central-1a",
				"eu-central-1b",
				"eu-central-1c",
			},
			expectedAwsZones: []Zone{
				{
					Name:     "eu-central-1a",
					Workers:  "10.180.0.0/26",
					Public:   "10.180.0.64/27",
					Internal: "10.180.0.96/27",
				},
				{
					Name:     "eu-central-1b",
					Workers:  "10.180.0.128/26",
					Public:   "10.180.0.192/27",
					Internal: "10.180.0.224/27",
				},
				{
					Name:     "eu-central-1c",
					Workers:  "10.180.1.0/26",
					Public:   "10.180.1.64/27",
					Internal: "10.180.1.96/27",
				},
			},
		},
	} {
		t.Run(tname, func(t *testing.T) {
			// when
			infrastructureConfigBytes, err := GetInfrastructureConfig(tcase.givenNodesCidr, tcase.givenZoneNames)

			// then
			assert.NoError(t, err)

			// when
			var infrastructureConfig InfrastructureConfig
			err = json.Unmarshal(infrastructureConfigBytes, &infrastructureConfig)

			// then
			require.NoError(t, err)
			assert.Equal(t, apiVersion, infrastructureConfig.TypeMeta.APIVersion)
			assert.Equal(t, infrastructureConfigKind, infrastructureConfig.TypeMeta.Kind)

			assert.Equal(t, tcase.givenNodesCidr, *infrastructureConfig.Networks.VPC.CIDR)
			for i, actualZone := range infrastructureConfig.Networks.Zones {
				assertIPRanges(t, tcase.expectedAwsZones[i], actualZone)
			}
		})
	}
}

func assertIPRanges(t *testing.T, expectedZone Zone, actualZone Zone) {
	assert.Equal(t, expectedZone.Name, actualZone.Name)
	assert.Equal(t, expectedZone.Internal, actualZone.Internal)
	assert.Equal(t, expectedZone.Workers, actualZone.Workers)
	assert.Equal(t, expectedZone.Public, actualZone.Public)
}

func TestWorkerConfig(t *testing.T) {
	t.Run("Create worker config", func(t *testing.T) {
		// when
		configBytes, err := GetWorkerConfig()

		// then
		require.NoError(t, err)
		var config WorkerConfig

		err = json.Unmarshal(configBytes, &config)
		require.NoError(t, err)

		assert.Equal(t, awsIMDSv2HTTPPutResponseHopLimit, *config.InstanceMetadataOptions.HTTPPutResponseHopLimit)
		assert.Equal(t, HTTPTokensRequired, *config.InstanceMetadataOptions.HTTPTokens)
	})
}
