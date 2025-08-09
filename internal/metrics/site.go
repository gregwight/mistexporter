package metrics

import "github.com/gregwight/mistclient"

// SiteLabelNames defines the labels attached site metrics.
var SiteLabelNames = []string{
	"site_name",
	"country_code",
	"timezone",
}

// SiteLabelValues generates label values for site metrics.
func SiteLabelValues(s mistclient.Site) []string {
	return []string{
		s.Name,
		s.CountryCode,
		s.Timezone,
	}
}
