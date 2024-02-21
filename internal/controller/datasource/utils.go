package datasource

import (
	"context"
	"fmt"
	"reflect"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/pkg/errors"
	kubeV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func makeJSONData(data *string) (map[string]interface{}, error) {
	jd := make(map[string]interface{})
	if data != nil && *data != "" {
		if err := json.Unmarshal([]byte(*data), &jd); err != nil {
			return nil, errors.Wrap(err, errUnmarshalJson)
		}
	}
	return jd, nil
}

func makeSecureJSONData(data *string) (map[string]string, error) {
	sjd := make(map[string]string)
	if data != nil && *data != "" {
		if err := json.Unmarshal([]byte(*data), &sjd); err != nil {
			return nil, errors.Wrap(err, errUnmarshalSecureJson)
		}
	}
	return sjd, nil
}

func secretToStringMap(secret *kubeV1.Secret) map[string]string {
	sjd := make(map[string]string)
	if secret == nil {
		return sjd
	}
	for key, val := range secret.Data {
		sjd[key] = string(val)
	}
	return sjd
}

func jsonDataWithHeaders(inputJSONData map[string]interface{}, inputSecureJSONData map[string]string, headers map[string]string) (map[string]interface{}, map[string]string) {
	jsonData := make(map[string]interface{})
	for name, value := range inputJSONData {
		jsonData[name] = value
	}

	secureJSONData := make(map[string]string)
	for name, value := range inputSecureJSONData {
		secureJSONData[name] = value
	}

	idx := 1
	for name, value := range headers {
		jsonData[fmt.Sprintf("httpHeaderName%d", idx)] = name
		secureJSONData[fmt.Sprintf("httpHeaderValue%d", idx)] = value
		idx++
	}

	return jsonData, secureJSONData
}

func defaultString(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// nolint: unparam
func defaultBool(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func compareOptional[K comparable](desired *K, actual K, defaultValue K) bool {
	var expected K
	if desired == nil {
		expected = defaultValue
	} else {
		expected = *desired
	}
	return actual == expected
}

func compareMap(desired map[string]interface{}, actual map[string]interface{}) bool {
	if len(desired) != len(actual) {
		return false
	}
	for key, value := range desired {
		if _, ok := actual[key]; !ok {
			return false
		}
		equal, ok := compareComparable(value, actual[key])
		if ok {
			if !equal {
				return false
			}
			continue
		}
		typeA := reflect.TypeOf(desired)
		if typeA == reflect.TypeOf(map[string]interface{}{}) {
			if !compareMap(value.(map[string]interface{}), actual[key].(map[string]interface{})) {
				return false
			}
		}
	}
	return true
}

// compareComparable tries to compare to values of different types. It returns a boolean indicating if the values are
// equal and a boolean indicating if the comparison was successful
func compareComparable(desired interface{}, actual interface{}) (bool, bool) {
	typeA := reflect.TypeOf(desired)
	typeB := reflect.TypeOf(actual)
	if typeA.Comparable() && typeB.Comparable() && typeA.ConvertibleTo(typeB) {
		converted := reflect.ValueOf(actual).Convert(typeA)
		return converted.Equal(reflect.ValueOf(desired)), true
	}
	return false, false
}

func compareMapKeys[T1, T2 comparable](desired map[string]T1, actual map[string]T2) bool {
	if len(desired) != len(actual) {
		return false
	}
	for key := range desired {
		if _, ok := actual[key]; !ok {
			return false
		}
	}
	return true
}

func (c *external) getValueFromSecret(ctx context.Context, selector v1.SecretKeySelector) (*string, error) {
	secret, err := c.getSecret(ctx, selector.SecretReference)
	if resource.IgnoreNotFound(err) != nil {
		return nil, errors.Wrap(err, errGetSecret)
	}

	pwRaw := secret.Data[selector.Key]
	strValue := string(pwRaw)
	return &strValue, nil
}
