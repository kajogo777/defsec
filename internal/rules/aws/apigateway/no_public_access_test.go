package apigateway

import (
	"testing"

	"github.com/aquasecurity/defsec/internal/types"

	"github.com/aquasecurity/defsec/pkg/state"

	"github.com/aquasecurity/defsec/pkg/providers/aws/apigateway"
	"github.com/aquasecurity/defsec/pkg/scan"

	"github.com/stretchr/testify/assert"
)

func TestCheckNoPublicAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    apigateway.APIGateway
		expected bool
	}{
		{
			name: "API GET method without authorization",
			input: apigateway.APIGateway{
				APIs: []apigateway.API{
					{
						Metadata:     types.NewTestMetadata(),
						ProtocolType: types.String(apigateway.ProtocolTypeREST, types.NewTestMetadata()),
						RESTMethods: []apigateway.RESTMethod{
							{
								Metadata:          types.NewTestMetadata(),
								HTTPMethod:        types.String("GET", types.NewTestMetadata()),
								APIKeyRequired:    types.Bool(false, types.NewTestMetadata()),
								AuthorizationType: types.String(apigateway.AuthorizationNone, types.NewTestMetadata()),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "API OPTION method without authorization",
			input: apigateway.APIGateway{
				APIs: []apigateway.API{
					{
						Metadata:     types.NewTestMetadata(),
						ProtocolType: types.String(apigateway.ProtocolTypeREST, types.NewTestMetadata()),
						RESTMethods: []apigateway.RESTMethod{
							{
								Metadata:          types.NewTestMetadata(),
								HTTPMethod:        types.String("OPTION", types.NewTestMetadata()),
								APIKeyRequired:    types.Bool(true, types.NewTestMetadata()),
								AuthorizationType: types.String(apigateway.AuthorizationNone, types.NewTestMetadata()),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "API GET method with IAM authorization",
			input: apigateway.APIGateway{
				APIs: []apigateway.API{
					{
						Metadata:     types.NewTestMetadata(),
						ProtocolType: types.String(apigateway.ProtocolTypeREST, types.NewTestMetadata()),
						RESTMethods: []apigateway.RESTMethod{
							{
								Metadata:          types.NewTestMetadata(),
								HTTPMethod:        types.String("GET", types.NewTestMetadata()),
								APIKeyRequired:    types.Bool(false, types.NewTestMetadata()),
								AuthorizationType: types.String(apigateway.AuthorizationIAM, types.NewTestMetadata()),
							},
						},
					},
				},
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var testState state.State
			testState.AWS.APIGateway = test.input
			results := CheckNoPublicAccess.Evaluate(&testState)
			var found bool
			for _, result := range results {
				if result.Status() == scan.StatusFailed && result.Rule().LongID() == CheckNoPublicAccess.Rule().LongID() {
					found = true
				}
			}
			if test.expected {
				assert.True(t, found, "Rule should have been found")
			} else {
				assert.False(t, found, "Rule should not have been found")
			}
		})
	}
}
