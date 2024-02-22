package datasource

import (
	"context"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
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

func (c *external) getValueFromSecret(ctx context.Context, selector v1.SecretKeySelector) (*string, error) {
	secret, err := c.getSecret(ctx, selector.SecretReference)
	if resource.IgnoreNotFound(err) != nil {
		return nil, errors.Wrap(err, errGetSecret)
	}

	pwRaw := secret.Data[selector.Key]
	strValue := string(pwRaw)
	return &strValue, nil
}
