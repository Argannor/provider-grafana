package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
