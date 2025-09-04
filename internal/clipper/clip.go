package clipper

import (
	"fmt"
	"math"

	"mbdvr/internal/types"
)

type ClipConfig struct {
	StartTime *float64 // nil = from beginning
	EndTime   *float64 // nil = to end
}

type ClipInfo struct {
	MinTimestamp    float64
	MaxTimestamp    float64
	TotalDuration   float64
	OriginalPoints  int
	ClippedPoints   int
	StartFrame      int // Index of first clipped point
	EndFrame        int // Index of last clipped point
	ActualStartTime float64
	ActualEndTime   float64
}

func ClipDataset(dataset *types.Dataset, config ClipConfig) (*types.Dataset, ClipInfo, error) {
	if dataset == nil || len(dataset.Points) == 0 {
		return nil, ClipInfo{}, fmt.Errorf("dataset is empty")
	}

	info := ClipInfo{
		OriginalPoints: len(dataset.Points),
		MinTimestamp:   math.Inf(1),
		MaxTimestamp:   math.Inf(-1),
	}

	for _, point := range dataset.Points {
		if point.Timestamp < info.MinTimestamp {
			info.MinTimestamp = point.Timestamp
		}
		if point.Timestamp > info.MaxTimestamp {
			info.MaxTimestamp = point.Timestamp
		}
	}

	info.TotalDuration = info.MaxTimestamp - info.MinTimestamp

	startTime := info.MinTimestamp
	endTime := info.MaxTimestamp

	if config.StartTime != nil {
		if *config.StartTime < info.MinTimestamp || *config.StartTime > info.MaxTimestamp {
			return nil, info, fmt.Errorf("start time %.2f is out of bounds (%.2f - %.2f)", *config.StartTime, info.MinTimestamp, info.MaxTimestamp)
		}
		startTime = *config.StartTime
	}

	if config.EndTime != nil {
		if *config.EndTime < info.MinTimestamp || *config.EndTime > info.MaxTimestamp {
			return nil, info, fmt.Errorf("end time %.2f is out of bounds (%.2f - %.2f)", *config.EndTime, info.MinTimestamp, info.MaxTimestamp)
		}
		endTime = *config.EndTime
	}

	if endTime <= startTime {
		return nil, info, fmt.Errorf("end time %.2f must be greater than start time %.2f", endTime, startTime)
	}

	//Find closest frames to start and end times
	startFrame := -1
	endFrame := -1
	for i, point := range dataset.Points {
		if startFrame == -1 && point.Timestamp >= startTime {
			startFrame = i
		}
		if point.Timestamp <= endTime {
			endFrame = i
		}
	}

	if startFrame == -1 || endFrame == -1 || startFrame > endFrame {
		return nil, info, fmt.Errorf("no data points found in the specified time range")
	}

	clippedPoints := dataset.Points[startFrame : endFrame+1]

	info.ClippedPoints = len(clippedPoints)
	info.StartFrame = startFrame
	info.EndFrame = endFrame
	info.ActualStartTime = clippedPoints[0].Timestamp
	info.ActualEndTime = clippedPoints[len(clippedPoints)-1].Timestamp

	clippedDataset := &types.Dataset{
		Points:  clippedPoints,
		Columns: dataset.Columns,
		Metadata: map[string]interface{}{
			"original_points":   info.OriginalPoints,
			"clipped_points":    info.ClippedPoints,
			"original_duration": info.TotalDuration,
			"clipped_duration":  info.ActualEndTime - info.ActualStartTime,
			"start_time":        info.ActualStartTime,
			"end_time":          info.ActualEndTime,
			"requested_start":   startTime,
			"requested_end":     endTime,
		},
	}

	return clippedDataset, info, nil
}

func FormatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	} else if seconds < 3600 {
		minutes := int(seconds / 60)
		remaining := seconds - float64(minutes*60)
		return fmt.Sprintf("%dm %.1fs", minutes, remaining)
	} else {
		hours := int(seconds / 3600)
		remaining := seconds - float64(hours*3600)
		minutes := int(remaining / 60)
		secs := remaining - float64(minutes*60)
		return fmt.Sprintf("%dh %dm %.1fs", hours, minutes, secs)
	}
}
