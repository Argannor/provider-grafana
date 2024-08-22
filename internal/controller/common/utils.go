package common

import (
	"fmt"
	"reflect"

	kubeV1 "k8s.io/api/core/v1"
)

func SecretToStringMap(secret *kubeV1.Secret) map[string]string {
	sjd := make(map[string]string)
	if secret == nil {
		return sjd
	}
	for key, val := range secret.Data {
		sjd[key] = string(val)
	}
	return sjd
}

func JsonDataWithHeaders(inputJSONData map[string]interface{}, inputSecureJSONData map[string]string, headers map[string]string) (map[string]interface{}, map[string]string) {
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

func DefaultString(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// nolint: unparam
func DefaultBool(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func CompareOptional[K comparable](desired *K, actual K, defaultValue K) bool {
	var expected K
	if desired == nil {
		expected = defaultValue
	} else {
		expected = *desired
	}
	return actual == expected
}

func CompareMap(desired map[string]interface{}, actual map[string]interface{}) (bool, error) {
	if len(desired) != len(actual) {
		return false, nil
	}
	for key, value := range desired {
		if _, ok := actual[key]; !ok {
			return false, nil
		}
		equal, ok := compareComparable(value, actual[key])
		if ok {
			if !equal {
				return false, nil
			}
			continue
		}
		typeA := reflect.TypeOf(desired)
		if typeA == reflect.TypeOf(map[string]interface{}{}) {
			desiredValueType := reflect.TypeOf(value)
			actualValueType := reflect.TypeOf(actual[key])
			if desiredValueType != actualValueType {
				return false, nil
			}
			switch desiredValueType {
			case reflect.TypeOf(map[string]interface{}{}):
				return CompareMap(value.(map[string]interface{}), actual[key].(map[string]interface{}))
			case reflect.TypeOf([]interface{}{}):
				return CompareSlice(value.([]interface{}), actual[key].([]interface{}))
			default:
				return false, fmt.Errorf("Unsupported map type %s of value %v", desiredValueType, value)
			}
		}
		return false, fmt.Errorf("Unsupported type %s of value %v", typeA, value)
	}
	return true, nil
}

func CompareSlice(desired []interface{}, actual []interface{}) (bool, error) {
	if len(desired) != len(actual) {
		return false, nil
	}
	for i, value := range desired {
		equal, ok := compareComparable(value, actual[i])
		if ok {
			if !equal {
				return false, nil
			}
			continue
		}
		typeA := reflect.TypeOf(value)
		switch typeA {
		case reflect.TypeOf(map[string]interface{}{}):
			return CompareMap(value.(map[string]interface{}), actual[i].(map[string]interface{}))
		case reflect.TypeOf([]interface{}{}):
			return CompareSlice(value.([]interface{}), actual[i].([]interface{}))
		default:
			return false, fmt.Errorf("Unsupported type %s of value %v", typeA, value)
		}
	}
	return true, nil

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

func CompareMapKeys[T1, T2 comparable](desired map[string]T1, actual map[string]T2) bool {
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

func AsInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}
