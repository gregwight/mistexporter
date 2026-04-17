package collector

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gregwight/mistclient"
	"github.com/gregwight/mistexporter/internal/config"
	"github.com/gregwight/mistexporter/internal/filter"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// testAPIServerHandler serves mock API responses from the testdata directory.
func testAPIServerHandler(t *testing.T, dataDir string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token test-api-key" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		path := filepath.Join(dataDir, strings.TrimPrefix(r.URL.Path, "/"))
		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func TestNew(t *testing.T) {
	f, err := filter.New(nil)
	if err != nil {
		t.Fatalf("filter.New failed: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c, err := New(&mistclient.APIClient{}, "test-org", f, true, true, logger)
	if err != nil {
		t.Fatalf("New() returned an unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("New() returned a nil collector")
	}
	if c.orgID != "test-org" {
		t.Errorf("New() orgID = %q, want %q", c.orgID, "test-org")
	}
	if c.logger == nil {
		t.Error("New() logger is nil")
	}
}

func TestCollect(t *testing.T) {
	testCases := []struct {
		name              string
		filterCfg         *config.SiteFilter
		enableOrgMetrics  bool
		enableSiteMetrics bool
		handler           http.HandlerFunc
		expectedFile      string
		lint              bool
	}{
		{
			name:              "success",
			filterCfg:         nil,
			enableOrgMetrics:  true,
			enableSiteMetrics: true,
			handler:           testAPIServerHandler(t, "testdata"),
			expectedFile:      "testdata/success.prom",
			lint:              true,
		},
		{
			name:              "org metrics disabled",
			filterCfg:         nil,
			enableOrgMetrics:  false,
			enableSiteMetrics: true,
			handler:           testAPIServerHandler(t, "testdata"),
			expectedFile:      "testdata/site_only.prom",
			lint:              true,
		},
		{
			name:              "site metrics disabled",
			filterCfg:         nil,
			enableOrgMetrics:  true,
			enableSiteMetrics: false,
			handler:           testAPIServerHandler(t, "testdata"),
			expectedFile:      "testdata/org_only.prom",
			lint:              true,
		},
		{
			name:              "org and site metrics disabled",
			filterCfg:         nil,
			enableOrgMetrics:  false,
			enableSiteMetrics: false,
			handler:           testAPIServerHandler(t, "testdata"),
			expectedFile:      "testdata/empty.prom",
		},
		{
			name:              "api error on org endpoints",
			filterCfg:         nil,
			enableOrgMetrics:  true,
			enableSiteMetrics: true,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/v1/orgs/") {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				testAPIServerHandler(t, "testdata")(w, r)
			}),
			expectedFile: "testdata/org_error.prom",
		},
		{
			name:              "api error on site stats",
			filterCfg:         nil,
			enableOrgMetrics:  true,
			enableSiteMetrics: true,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/v1/sites/") && strings.HasSuffix(r.URL.Path, "/stats") {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				testAPIServerHandler(t, "testdata")(w, r)
			}),
			expectedFile: "testdata/site_stats_error.prom",
		},
		{
			name:              "with site filter",
			filterCfg:         &config.SiteFilter{Include: []string{"Test Site 1"}},
			enableOrgMetrics:  true,
			enableSiteMetrics: true,
			handler:           testAPIServerHandler(t, "testdata"),
			expectedFile:      "testdata/filtered.prom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)

			client, err := mistclient.New(&mistclient.Config{BaseURL: server.URL, APIKey: "test-api-key"}, nil)
			if err != nil {
				t.Fatalf("failed to create mist client: %v", err)
			}
			// We close the server inside the test run to ensure all client connections are terminated
			// before the next test case starts.
			t.Cleanup(server.Close)

			siteFilter, err := filter.New(tc.filterCfg)
			if err != nil {
				t.Fatalf("filter.New failed: %v", err)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			collector, err := New(client, "test-org-id", siteFilter, tc.enableOrgMetrics, tc.enableSiteMetrics, logger)
			if err != nil {
				t.Fatalf("New() returned an unexpected error: %v", err)
			}

			expected, err := os.Open(tc.expectedFile)
			if err != nil {
				t.Fatalf("failed to open expected metrics file %s: %v", tc.expectedFile, err)
			}
			defer expected.Close()

			if err := testutil.CollectAndCompare(collector, expected); err != nil {
				t.Errorf("unexpected metrics collected:\n%v", err)
			}

			if tc.lint {
				problems, err := testutil.CollectAndLint(collector)
				if err != nil {
					t.Errorf("metric linting failed with an error: %v", err)
				}
				if len(problems) > 0 {
					t.Errorf("metric linting failed with problems:\n%v", problems)
				}
			}
		})
	}
}
