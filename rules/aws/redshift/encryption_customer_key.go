package redshift

import (
	"github.com/aquasecurity/defsec/provider"
	"github.com/aquasecurity/defsec/rules"
	"github.com/aquasecurity/defsec/severity"
	"github.com/aquasecurity/defsec/state"
)

var CheckEncryptionCustomerKey = rules.Register(
	rules.Rule{
		Provider:    provider.AWSProvider,
		Service:     "redshift",
		ShortCode:   "encryption-customer-key",
		Summary:     "Redshift clusters should use at rest encryption",
		Impact:      "Data may be leaked if infrastructure is compromised",
		Resolution:  "Enable encryption using CMK",
		Explanation: `Redshift clusters that contain sensitive data or are subject to regulation should be encrypted at rest to prevent data leakage should the infrastructure be compromised.`,
		Links: []string{ 
			"https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-db-encryption.html",
		},
		Severity: severity.High,
	},
	func(s *state.State) (results rules.Results) {
		for _, x := range s.AWS.S3.Buckets {
			if x.Encryption.Enabled.IsFalse() {
				results.Add(
					"",
					x.Encryption.Enabled.Metadata(),
					x.Encryption.Enabled.Value(),
				)
			}
		}
		return
	},
)
