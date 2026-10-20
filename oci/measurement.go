package oci

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type AllotmentResolution string

const (
	ResolutionLocal   AllotmentResolution = "local"
	ResolutionRemote  AllotmentResolution = "remote"
	ResolutionRebuild AllotmentResolution = "rebuild"
	ResolutionFailed  AllotmentResolution = "failed"
)

type LookupResult string

const (
	LookupNotChecked LookupResult = "not_checked"
	LookupHit        LookupResult = "hit"
	LookupMiss       LookupResult = "miss"
	LookupError      LookupResult = "error"
)

type AllotmentMeasurement struct {
	AllotmentKey string

	Resolution AllotmentResolution

	RemoteKeyLookup  LookupResult
	RemoteBlobLookup LookupResult

	TotalDuration             time.Duration
	RemoteKeyLookupDuration   time.Duration
	RemoteBlobRestoreDuration time.Duration
	BuildDuration             time.Duration
	CompressionDuration       time.Duration
	RemoteBlobPushDuration    time.Duration

	RemoteRetryCount int

	RemoteArtifactBytesDownloaded int64
	RemoteArtifactBytesUploaded   int64
}

type BuildMeasurement struct {
	TotalDuration time.Duration
	Success       bool

	Allotments []AllotmentMeasurement
}

type BuildSummary struct {
	TotalDuration time.Duration
	Success       bool

	TotalAllotments int

	LocalHitCount      int
	RemoteRestoreCount int
	RebuildCount       int

	RemoteKeyHitCount  int
	RemoteKeyMissCount int

	RemoteBlobHitCount  int
	RemoteBlobMissCount int

	TotalArtifactBytesDownloaded int64
	TotalArtifactBytesUploaded   int64

	TotalRemoteRetryCount int
}

func newAllotmentMeasurement() AllotmentMeasurement {
	return AllotmentMeasurement{
		Resolution:       ResolutionFailed,
		RemoteKeyLookup:  LookupNotChecked,
		RemoteBlobLookup: LookupNotChecked,
	}
}

func newBuildMeasurement(allotmentCount int) *BuildMeasurement {
	measurement := &BuildMeasurement{
		Allotments: make([]AllotmentMeasurement, allotmentCount),
	}

	for i := range measurement.Allotments {
		measurement.Allotments[i] = newAllotmentMeasurement()
	}

	return measurement
}

func (m *BuildMeasurement) Summary() BuildSummary {
	summary := BuildSummary{
		TotalDuration:   m.TotalDuration,
		Success:         m.Success,
		TotalAllotments: len(m.Allotments),
	}

	for _, allotment := range m.Allotments {
		switch allotment.Resolution {
		case ResolutionLocal:
			summary.LocalHitCount++
		case ResolutionRemote:
			summary.RemoteRestoreCount++
		case ResolutionRebuild:
			summary.RebuildCount++
		}

		switch allotment.RemoteKeyLookup {
		case LookupHit:
			summary.RemoteKeyHitCount++
		case LookupMiss:
			summary.RemoteKeyMissCount++
		}

		switch allotment.RemoteBlobLookup {
		case LookupHit:
			summary.RemoteBlobHitCount++
		case LookupMiss:
			summary.RemoteBlobMissCount++
		}

		summary.TotalArtifactBytesDownloaded += allotment.RemoteArtifactBytesDownloaded
		summary.TotalArtifactBytesUploaded += allotment.RemoteArtifactBytesUploaded
		summary.TotalRemoteRetryCount += allotment.RemoteRetryCount
	}

	return summary
}

func durationMilliseconds(duration time.Duration) string {
	return strconv.FormatFloat(
		float64(duration)/float64(time.Millisecond),
		'f',
		3,
		64,
	)
}

func (m BuildMeasurement) ExportCSV(outputDir string) (string, error) {
	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create benchmark output directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("build-summary-%s.csv", timestamp)
	outputPath := filepath.Join(outputDir, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create benchmark CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	header := []string{
		"record_type",
		"allotment_key",
		"resolution",
		"remote_key_lookup",
		"remote_blob_lookup",
		"total_duration_ms",
		"remote_key_lookup_duration_ms",
		"remote_blob_restore_duration_ms",
		"build_duration_ms",
		"compression_duration_ms",
		"remote_blob_push_duration_ms",
		"remote_retry_count",
		"remote_artifact_bytes_downloaded",
		"remote_artifact_bytes_uploaded",
		"success",
		"total_allotments",
		"local_hit_count",
		"remote_restore_count",
		"rebuild_count",
		"remote_key_hit_count",
		"remote_key_miss_count",
		"remote_blob_hit_count",
		"remote_blob_miss_count",
		"total_artifact_bytes_downloaded",
		"total_artifact_bytes_uploaded",
		"total_remote_retry_count",
	}

	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write benchmark CSV header: %w", err)
	}

	summary := m.Summary()

	summaryRow := []string{
		"summary",
		"",
		"",
		"",
		"",
		durationMilliseconds(m.TotalDuration),
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		fmt.Sprint(m.Success),
		fmt.Sprint(summary.TotalAllotments),
		fmt.Sprint(summary.LocalHitCount),
		fmt.Sprint(summary.RemoteRestoreCount),
		fmt.Sprint(summary.RebuildCount),
		fmt.Sprint(summary.RemoteKeyHitCount),
		fmt.Sprint(summary.RemoteKeyMissCount),
		fmt.Sprint(summary.RemoteBlobHitCount),
		fmt.Sprint(summary.RemoteBlobMissCount),
		fmt.Sprint(summary.TotalArtifactBytesDownloaded),
		fmt.Sprint(summary.TotalArtifactBytesUploaded),
		fmt.Sprint(summary.TotalRemoteRetryCount),
	}

	if err := writer.Write(summaryRow); err != nil {
		return "", fmt.Errorf("failed to write benchmark summary row: %w", err)
	}

	for _, allotment := range m.Allotments {
		allotmentRow := []string{
			"allotment",
			fmt.Sprint(allotment.AllotmentKey),
			fmt.Sprint(allotment.Resolution),
			fmt.Sprint(allotment.RemoteKeyLookup),
			fmt.Sprint(allotment.RemoteBlobLookup),
			durationMilliseconds(allotment.TotalDuration),
			durationMilliseconds(allotment.RemoteKeyLookupDuration),
			durationMilliseconds(allotment.RemoteBlobRestoreDuration),
			durationMilliseconds(allotment.BuildDuration),
			durationMilliseconds(allotment.CompressionDuration),
			durationMilliseconds(allotment.RemoteBlobPushDuration),
			fmt.Sprint(allotment.RemoteRetryCount),
			fmt.Sprint(allotment.RemoteArtifactBytesDownloaded),
			fmt.Sprint(allotment.RemoteArtifactBytesUploaded),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		}

		if err := writer.Write(allotmentRow); err != nil {
			return "", fmt.Errorf(
				"failed to write benchmark allotment row for %s: %w",
				allotment.AllotmentKey,
				err,
			)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("failed to flush benchmark CSV: %w", err)
	}

	return outputPath, nil
}
