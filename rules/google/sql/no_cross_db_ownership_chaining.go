package sql

import (
	"github.com/aquasecurity/defsec/provider"
	"github.com/aquasecurity/defsec/rules"
	"github.com/aquasecurity/defsec/severity"
	"github.com/aquasecurity/defsec/state"
)

var CheckNoCrossDbOwnershipChaining = rules.Register(
	rules.Rule{
		Provider:    provider.GoogleProvider,
		Service:     "sql",
		ShortCode:   "no-cross-db-ownership-chaining",
		Summary:     "Cross-database ownership chaining should be disabled",
		Impact:      "Unintended access to sensitive data",
		Resolution:  "Disable cross database ownership chaining",
		Explanation: `Cross-database ownership chaining, also known as cross-database chaining, is a security feature of SQL Server that allows users of databases access to other databases besides the one they are currently using.`,
		Links: []string{ 
			"https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/cross-db-ownership-chaining-server-configuration-option?view=sql-server-ver15",
		},
		Severity: severity.Medium,
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
