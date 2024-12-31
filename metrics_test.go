package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func loadMockData(t *testing.T, filePath string) string {
	t.Helper()

	data, err := os.ReadFile(filePath) // Use os.ReadFile for modern Go
	if err != nil {
		t.Fatalf("Failed to load mock data from file %s: %v", filePath, err)
	}
	return string(data)
}

func runTestWithMockData(t *testing.T, mainFile string, expectedMetrics map[string]string) {

	// Load mock JSON outputs from files
	mainOutput := loadMockData(t, mainFile)
	backupsOutput := loadMockData(t, "assets/backups.json")

	// Replace execute function to return appropriate mock data
	execute = func(lockKey string, parameters string) string {
		switch parameters {
		case "--output json --list":
			return mainOutput
		case "--output json --list-backups":
			return backupsOutput
		default:
			return "{}" // Return empty JSON for unexpected cases
		}
	}

	// Create test server
	handler := getMetrics()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to the handler
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Validate expected metrics
	for metric, expected := range expectedMetrics {
		if err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), metric); err != nil {
			t.Errorf("Unexpected metrics for %s:\n%s", metric, err)
		}
	}
}

func TestMetrics(t *testing.T) {
	registerMetrics()
	tests := []struct {
		name            string
		mainFile        string
		expectedMetrics map[string]string
	}{
		{
			name:     "With Replication Lags",
			mainFile: "assets/with-replication-lags.json",
			expectedMetrics: map[string]string{
				"replikator_replication_lag": `
					# HELP replikator_replication_lag Replication lag from master server
					# TYPE replikator_replication_lag gauge
					replikator_replication_lag{state="running"} 5
				`,
				"replikator_replication_lags": `
					# HELP replikator_replication_lags Replication lag per channel
					# TYPE replikator_replication_lags gauge
					replikator_replication_lags{channel="worst"} 5
					replikator_replication_lags{channel="aurora"} 0
					replikator_replication_lags{channel="mysql-rds"} 5
				`,
				"replikator_backup_timestamp_seconds": `
					# HELP replikator_backup_timestamp_seconds Backup timestamp in seconds
					# TYPE replikator_backup_timestamp_seconds gauge
					replikator_backup_timestamp_seconds{backup="backup-20241225-1600"} 1.735142404e+09
					replikator_backup_timestamp_seconds{backup="backup-20241226-0400"} 1.735185605e+09
					replikator_backup_timestamp_seconds{backup="backup-20241226-1600"} 1.735228805e+09
					replikator_backup_timestamp_seconds{backup="backup-20241227-0400"} 1.735272006e+09
					replikator_backup_timestamp_seconds{backup="backup-20241227-1600"} 1.735315206e+09
					replikator_backup_timestamp_seconds{backup="backup-20241228-0400"} 1.735358404e+09
					replikator_backup_timestamp_seconds{backup="backup-20241228-1600"} 1.735401605e+09
					replikator_backup_timestamp_seconds{backup="backup-20241229-0400"} 1.735444806e+09
					replikator_backup_timestamp_seconds{backup="backup-20241229-1600"} 1.735488005e+09
					replikator_backup_timestamp_seconds{backup="backup-20241230-0400"} 1.735531206e+09
				`,
			},
		},
		{
			name:     "Without Replication Lags",
			mainFile: "assets/without-replication-lags.json",
			expectedMetrics: map[string]string{
				"replikator_replication_lag": `
					# HELP replikator_replication_lag Replication lag from master server
					# TYPE replikator_replication_lag gauge
					replikator_replication_lag{state="running"} 0
				`,
				"replikator_replication_lags": "",
			},
		},
		{
			name:     "MySQL Stopped",
			mainFile: "assets/mysql-stopped.json",
			expectedMetrics: map[string]string{
				"replikator_replication_lag": `
					# HELP replikator_replication_lag Replication lag from master server
					# TYPE replikator_replication_lag gauge
					replikator_replication_lag{state="stopped"} -1
				`,
			},
		},
		{
			name:     "Replication Stopped",
			mainFile: "assets/replication-stopped.json",
			expectedMetrics: map[string]string{
				"replikator_replication_lag": `
					# HELP replikator_replication_lag Replication lag from master server
					# TYPE replikator_replication_lag gauge
					replikator_replication_lag{state="stopped"} -1
				`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTestWithMockData(t, tc.mainFile, tc.expectedMetrics)
		})
	}
}
