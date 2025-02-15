package parser

import (
	"fmt"

	"github.com/blang/semver/v4"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kong/kubernetes-ingress-controller/v2/internal/util"
)

// kongHeaderRegexPrefix is a reserved prefix string that Kong uses to determine if it should parse a header value
// as a regex
const kongHeaderRegexPrefix = "~*"

// MinRegexHeaderKongVersion is the minimum Kong version that supports regex header matches
var MinRegexHeaderKongVersion = semver.MustParse("2.8.0")

// -----------------------------------------------------------------------------
// Translate Utilities - Gateway
// -----------------------------------------------------------------------------

// convertGatewayMatchHeadersToKongRouteMatchHeaders takes an input list of Gateway APIs HTTPHeaderMatch
// and converts these header matching rules to the format expected by go-kong.
func convertGatewayMatchHeadersToKongRouteMatchHeaders(headers []gatewayv1alpha2.HTTPHeaderMatch) (map[string][]string, error) {
	// iterate through each provided header match checking for invalid
	// options and otherwise converting to kong type format.
	convertedHeaders := make(map[string][]string)
	for _, header := range headers {
		if _, exists := convertedHeaders[string(header.Name)]; exists {
			return nil, fmt.Errorf("multiple header matches for the same header are not allowed: %s",
				string(header.Name))
		}
		if header.Type != nil && *header.Type == gatewayv1alpha2.HeaderMatchRegularExpression {
			if util.GetKongVersion().LT(MinRegexHeaderKongVersion) {
				return nil, fmt.Errorf("Kong version %s does not support HeaderMatchRegularExpression",
					util.GetKongVersion().String())
			}
			convertedHeaders[string(header.Name)] = []string{kongHeaderRegexPrefix + header.Value}
		} else if header.Type == nil || *header.Type == gatewayv1alpha2.HeaderMatchExact {
			convertedHeaders[string(header.Name)] = []string{header.Value}
		} else {
			return nil, fmt.Errorf("unknown/unsupported header match type: %s", string(*header.Type))
		}
	}

	return convertedHeaders, nil
}
