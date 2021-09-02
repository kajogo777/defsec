package gke

import (
	"github.com/aquasecurity/defsec/provider"
	"github.com/aquasecurity/defsec/rules"
	"github.com/aquasecurity/defsec/severity"
	"github.com/aquasecurity/defsec/state"
)

var CheckNodeShieldingEnabled = rules.Register(
	rules.Rule{
		Provider:    provider.GoogleProvider,
		Service:     "gke",
		ShortCode:   "node-shielding-enabled",
		Summary:     "Shielded GKE nodes not enabled.",
		Impact:      "Node identity and integrity can't be verified without shielded GKE nodes",
		Resolution:  "Enable node shielding",
		Explanation: `CIS GKE Benchmark Recommendation: 6.5.5. Ensure Shielded GKE Nodes are Enabled

Shielded GKE Nodes provide strong, verifiable node identity and integrity to increase the security of GKE nodes and should be enabled on all GKE clusters.`,
		Links: []string{ 
			"https://cloud.google.com/kubernetes-engine/docs/how-to/hardening-your-cluster#shielded_nodes",
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
