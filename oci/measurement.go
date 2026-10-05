package oci

import "time"

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
	}

	return summary
}
