package metrics

import (
	"reflect"
	"testing"

	"github.com/gregwight/mistclient"
)

func TestSiteLabelNames(t *testing.T) {
	expected := []string{
		"site_name",
		"country_code",
		"timezone",
	}

	if !reflect.DeepEqual(SiteLabelNames, expected) {
		t.Errorf("SiteLabelNames = %v, want %v", SiteLabelNames, expected)
	}
}

func TestSiteLabelValues(t *testing.T) {
	site := mistclient.Site{
		Name:        "Test Site Name",
		CountryCode: "GB",
		Timezone:    "Europe/London",
	}

	expected := []string{
		"Test Site Name",
		"GB",
		"Europe/London",
	}

	actual := SiteLabelValues(site)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("SiteLabelValues() = %v, want %v", actual, expected)
	}
}
