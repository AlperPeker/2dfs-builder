package oci

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestBuildMeasurementSummaryAggregatesRemoteRetries(t *testing.T) {
	measurement := BuildMeasurement{
		Allotments: []AllotmentMeasurement{
			{RemoteRetryCount: 2},
			{RemoteRetryCount: 1},
			{RemoteRetryCount: 0},
		},
	}

	summary := measurement.Summary()

	if summary.TotalRemoteRetryCount != 3 {
		t.Fatalf(
			"expected total remote retry count 3, got %d",
			summary.TotalRemoteRetryCount,
		)
	}
}

func TestExportCSVWritesRemoteRetryMetricsInCorrectColumns(t *testing.T) {
	measurement := BuildMeasurement{
		Success: true,
		Allotments: []AllotmentMeasurement{
			{
				AllotmentKey:     "key-test",
				RemoteRetryCount: 2,
			},
		},
	}

	outputPath, err := measurement.ExportCSV(t.TempDir())
	if err != nil {
		t.Fatalf("ExportCSV returned error: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("failed to open exported CSV: %v", err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("failed to read exported CSV: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected header + summary + allotment rows, got %d rows", len(records))
	}

	header := records[0]
	summaryRow := records[1]
	allotmentRow := records[2]

	if len(summaryRow) != len(header) {
		t.Fatalf(
			"summary row has %d columns, header has %d",
			len(summaryRow),
			len(header),
		)
	}

	if len(allotmentRow) != len(header) {
		t.Fatalf(
			"allotment row has %d columns, header has %d",
			len(allotmentRow),
			len(header),
		)
	}

	columnIndex := make(map[string]int, len(header))
	for i, column := range header {
		columnIndex[column] = i
	}

	remoteRetryIndex, ok := columnIndex["remote_retry_count"]
	if !ok {
		t.Fatal("remote_retry_count column not found")
	}

	totalRemoteRetryIndex, ok := columnIndex["total_remote_retry_count"]
	if !ok {
		t.Fatal("total_remote_retry_count column not found")
	}

	if allotmentRow[remoteRetryIndex] != "2" {
		t.Fatalf(
			"expected allotment remote_retry_count 2, got %q",
			allotmentRow[remoteRetryIndex],
		)
	}

	if summaryRow[remoteRetryIndex] != "" {
		t.Fatalf(
			"expected summary remote_retry_count to be empty, got %q",
			summaryRow[remoteRetryIndex],
		)
	}

	if summaryRow[totalRemoteRetryIndex] != "2" {
		t.Fatalf(
			"expected summary total_remote_retry_count 2, got %q",
			summaryRow[totalRemoteRetryIndex],
		)
	}

	if allotmentRow[totalRemoteRetryIndex] != "" {
		t.Fatalf(
			"expected allotment total_remote_retry_count to be empty, got %q",
			allotmentRow[totalRemoteRetryIndex],
		)
	}
}
