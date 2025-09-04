package cleaner

import (
	"fmt"
	"math"
	"sort"

	"mbdvr/internal/types"
)

type CleanConfig struct {
	RequiredColumns   []string
	RemoveOutliers    bool
	OutlierMethod     string  // "iqr" or "zscore"
	MaxMissingPercent float64 // 0-100, max % of missing data per row
	ZScoreThreshold   float64 // for zscore outlier detection
}

type CleanStats struct {
	OriginalPoints  int
	RemovedMissing  int
	RemovedOutliers int
	FinalPoints     int
}

func CleanDataset(dataset *types.Dataset, config CleanConfig) (*types.Dataset, CleanStats, error) {
	stats := CleanStats{
		OriginalPoints: len(dataset.Points),
	}

	cleanedPoints := dataset.Points

	if config.MaxMissingPercent > 0 {
		cleanedPoints, stats.RemovedMissing = filterMissingData(cleanedPoints, config.RequiredColumns, config.MaxMissingPercent)
		fmt.Printf("Removed %d points due to missing data\n", stats.RemovedMissing)
	}

	if config.RemoveOutliers {
		cleanedPoints, stats.RemovedOutliers = filterOutliers(cleanedPoints, config.RequiredColumns, config.OutlierMethod, config.ZScoreThreshold)
		fmt.Printf("Removed %d points as outliers\n", stats.RemovedOutliers)
	}

	stats.FinalPoints = len(cleanedPoints)

	cleanedDataset := &types.Dataset{
		Points:  cleanedPoints,
		Columns: dataset.Columns,
		Metadata: map[string]interface{}{
			"original_points":    stats.OriginalPoints,
			"cleaned_points":     stats.FinalPoints,
			"cleaning_config":    config,
			"points_removed":     stats.OriginalPoints - stats.FinalPoints,
			"removal_percentage": float64(stats.OriginalPoints-stats.FinalPoints) / float64(stats.OriginalPoints) * 100,
		},
	}

	return cleanedDataset, stats, nil
}

func filterMissingData(points []types.DataPoint, requiredCols []string, maxMissingPercent float64) ([]types.DataPoint, int) {
	var filtered []types.DataPoint
	removedCount := 0
	maxMissing := int(math.Floor(float64(len(requiredCols)) * maxMissingPercent / 100.0))

	for _, p := range points {
		missing := 0
		for _, col := range requiredCols {
			if val, ok := p.Data[col]; !ok || math.IsNaN(val) {
				missing++
			}
		}
		if missing <= maxMissing {
			filtered = append(filtered, p)
		} else {
			removedCount++
		}
	}

	return filtered, removedCount
}

func filterOutliers(points []types.DataPoint, cols []string, method string, zThreshold float64) ([]types.DataPoint, int) {
	if len(cols) == 0 {
		return points, 0
	}

	var filtered []types.DataPoint
	removedCount := 0

	outlierBounds := make(map[string][2]float64) // col -> (min, max)

	for _, col := range cols {
		values := extractColumnValues(points, col)
		if len(values) == 0 {
			continue
		}

		var lowerBound, upperBound float64

		switch method {
		case "iqr":
			lowerBound, upperBound = calculateIQRBounds(values)
		case "zscore":
			lowerBound, upperBound = calculateZScoreBounds(values, zThreshold)
		default:
			lowerBound, upperBound = calculateIQRBounds(values) // Default to IQR
		}

		outlierBounds[col] = [2]float64{lowerBound, upperBound}
	}

	for _, p := range points {
		isOutlier := false

		for _, col := range cols {
			if bounds, ok := outlierBounds[col]; ok {
				if val, ok := p.Data[col]; ok {
					if val < bounds[0] || val > bounds[1] {
						isOutlier = true
						break
					}
				}
			}
		}

		if !isOutlier {
			filtered = append(filtered, p)
		} else {
			removedCount++
		}
	}

	return filtered, removedCount
}

func extractColumnValues(points []types.DataPoint, col string) []float64 {
	var values []float64
	for _, p := range points {
		if val, ok := p.Data[col]; ok && !math.IsNaN(val) {
			values = append(values, val)
		}
	}
	return values
}

func calculateIQRBounds(values []float64) (float64, float64) {
	if len(values) == 0 {
		return math.NaN(), math.NaN()
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	q1 := percentile(sorted, 25)
	q3 := percentile(sorted, 75)
	iqr := q3 - q1

	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	return lowerBound, upperBound
}

func calculateZScoreBounds(values []float64, threshold float64) (float64, float64) {
	if len(values) == 0 {
		return math.NaN(), math.NaN()
	}

	mean := mean(values)
	stdDev := stdDev(values, mean)

	lowerBound := mean - threshold*stdDev
	upperBound := mean + threshold*stdDev

	return lowerBound, upperBound
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	k := (p / 100) * float64(len(sorted)-1)
	f := math.Floor(k)
	c := math.Ceil(k)
	if f == c {
		return sorted[int(k)]
	}
	d0 := sorted[int(f)] * (c - k)
	d1 := sorted[int(c)] * (k - f)
	return d0 + d1
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	varianceSum := 0.0
	for _, v := range values {
		varianceSum += (v - mean) * (v - mean)
	}
	variance := varianceSum / float64(len(values))
	return math.Sqrt(variance)
}
