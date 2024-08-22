package common

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CompareMap(t *testing.T) {
	desired := map[string]interface{}{
		"oauthPassThru": false,
		"tracesToLogsV2": map[string]interface{}{
			"customQuery":   false,
			"datasourceUid": "auid",
			"tags": []interface{}{
				map[string]interface{}{
					"key":    "k8s.namespace.name",
					"values": "k8s_namespace_name",
				},
			},
		},
		"tracesToMetrics": map[string]interface{}{
			"datasourceUid": "buid",
			"tags": []interface{}{
				map[string]interface{}{
					"key":    "k8s.namespace.name",
					"values": "k8s_namespace_name",
				},
			},
		},
	}
	actual := map[string]interface{}{
		"oauthPassThru": false,
		"tracesToLogsV2": map[string]interface{}{
			"customQuery":   false,
			"datasourceUid": "auid",
			"tags": []interface{}{
				map[string]interface{}{
					"key":    "k8s.namespace.name",
					"values": "k8s_namespace_name",
				},
			},
		},
		"tracesToMetrics": map[string]interface{}{
			"datasourceUid": "buid",
			"tags": []interface{}{
				map[string]interface{}{
					"key":    "k8s.namespace.name",
					"values": "k8s_namespace_name",
				},
			},
		},
	}
	probe, err := CompareMap(desired, actual)
	assert.Nil(t, err)
	assert.True(t, probe)
}

func Test_CompareOptional(t *testing.T) {
	desired := "Test"
	assert.True(t, CompareOptional(&desired, "Test", ""))
	assert.False(t, CompareOptional(&desired, "Test1", ""))
	assert.False(t, CompareOptional(&desired, "", "Test1"))
	assert.True(t, CompareOptional(nil, "", ""))
	assert.True(t, CompareOptional(nil, "default", "default"))
	assert.False(t, CompareOptional(nil, "non-default", "default"))
}
