/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package datasource

import (
	"context"
	"testing"

	"github.com/argannor/provider-grafana/apis/oss/v1alpha1"
	"github.com/argannor/provider-grafana/internal/controller/common"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func TestObserve(t *testing.T) {
	type fields struct {
		service common.GrafanaAPI
		logger  logging.Logger
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		// TODO: Add test cases.
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.fields.service, logger: tc.fields.logger}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	headers := map[string][]byte{
		"Test": []byte("Test-Value"),
	}
	headersSecret := &v1.Secret{
		Data: headers,
	}
	cr := &v1alpha1.DataSource{
		Spec: v1alpha1.DataSourceSpec{
			ForProvider: v1alpha1.DataSourceParameters{
				AccessMode:        nil,
				BasicAuthEnabled:  nil,
				BasicAuthUsername: strRef("admin"),
				DatabaseName:      nil,
				IsDefault:         boolRef(true),
				JSONDataEncoded:   strRef("{\"public\": { \"value\": 1 } }"),
				Name:              nil,
				OrgID:             strRef("1"),
				Type:              strRef("prometheus"),
				UID:               nil,
				URL:               nil,
				Username:          nil,
			},
		},
	}
	atGrafana := &models.DataSource{
		Access:        "proxy",
		AccessControl: nil,
		BasicAuth:     false,
		BasicAuthUser: "admin",
		Database:      "",
		ID:            0,
		IsDefault:     true,
		JSONData: map[string]interface{}{
			"public":          map[string]interface{}{"value": 1},
			"httpHeaderName1": "Test",
		},
		Name:             "",
		OrgID:            1,
		ReadOnly:         false,
		SecureJSONFields: map[string]bool{"secret": true, "httpHeaderValue1": true},
		Type:             "prometheus",
		TypeLogoURL:      "",
		UID:              "",
		URL:              "",
		User:             "",
		Version:          0,
		WithCredentials:  false,
	}
	probe, err := isUpToDate(cr, atGrafana, 1, headersSecret, strRef("{ \"secret\": \"secretValue\" }"))
	assert.Nil(t, err)
	assert.True(t, probe)
}

func TestIsNotUpToDate(t *testing.T) {
	headers := map[string][]byte{
		"Test": []byte("Test-Value"),
	}
	headersSecret := &v1.Secret{
		Data: headers,
	}
	cr := &v1alpha1.DataSource{
		Spec: v1alpha1.DataSourceSpec{
			ForProvider: v1alpha1.DataSourceParameters{
				AccessMode:        nil,
				BasicAuthEnabled:  nil,
				BasicAuthUsername: strRef("admin"),
				DatabaseName:      nil,
				IsDefault:         boolRef(true),
				JSONDataEncoded:   strRef("{\"public\": { \"value\": 1 } }"),
				Name:              nil,
				OrgID:             strRef("1"),
				Type:              strRef("prometheus"),
				UID:               nil,
				URL:               nil,
				Username:          nil,
			},
		},
	}
	atGrafana := &models.DataSource{
		Access:        "proxy",
		AccessControl: nil,
		BasicAuth:     false,
		BasicAuthUser: "admin2",
		Database:      "",
		ID:            0,
		IsDefault:     true,
		JSONData: map[string]interface{}{
			"public":          map[string]interface{}{"value": 1},
			"httpHeaderName1": "Test",
		},
		Name:             "",
		OrgID:            1,
		ReadOnly:         false,
		SecureJSONFields: map[string]bool{"secret": true, "httpHeaderValue1": true},
		Type:             "prometheus",
		TypeLogoURL:      "",
		UID:              "",
		URL:              "",
		User:             "",
		Version:          0,
		WithCredentials:  false,
	}
	probe, err := isUpToDate(cr, atGrafana, 1, headersSecret, strRef("{ \"secret\": \"secretValue\" }"))
	assert.Nil(t, err)
	assert.False(t, probe)
}

func strRef(s string) *string {
	return &s
}
func boolRef(b bool) *bool {
	return &b
}
