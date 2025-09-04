package stats

//THIS PACKAGE ASSUMES A FILENAME PATTERN TO LOAD CONDITIONS (E.G. USER1_BORING.CSV, USER1_INTERESTED.CSV) --- IGNORE ---

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"mbdvr/internal/types"
)

const MAX_DATASETS = 10

type StatsConfig struct {
	AnalyzeColumns []string
	ByCondition    bool
	ByParticipant  bool
}

type ColumnStats struct {
	Column          string
	Mean            float64
	Median          float64
	StdDev          float64
	Min             float64
	Max             float64
	Count           int
	MissingCount    int
	OutlierCount    int
	OutlierMethod   string
	ZScoreThreshold float64
}

type StatsReport struct {
	OverallStats     []ColumnStats
	ConditionStats   map[string][]ColumnStats
	ParticipantStats map[string][]ColumnStats
}

func ComputeStats(dataset *types.Dataset, config StatsConfig) (*StatsReport, error) {
	if dataset == nil || len(dataset.Points) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}

	report := &StatsReport{
		ConditionStats:   make(map[string][]ColumnStats),
		ParticipantStats: make(map[string][]ColumnStats),
	}

	if len(config.AnalyzeColumns) == 0 {
		config.AnalyzeColumns = dataset.Columns
	}

	if config.ByCondition {
		conditionMap := make(map[string][]types.DataPoint)
		for _, point := range dataset.Points {
			condition := point.Condition
			if condition == "" {
				condition = "unknown"
			}
			conditionMap[condition] = append(conditionMap[condition], point)
		}

		for condition, points := range conditionMap {
			subDataset := &types.Dataset{
				Points:  points,
				Columns: dataset.Columns,
			}
			stats, err := computeColumnStats(subDataset, config.AnalyzeColumns, config)
			if err != nil {
				return nil, fmt.Errorf("failed to compute stats for condition %s: %v", condition, err)
			}
			report.ConditionStats[condition] = stats
		}
	}

	if config.ByParticipant {
		participantMap := make(map[string][]types.DataPoint)
		for _, point := range dataset.Points {
			participantID := point.ParticipantID
			if participantID == "" {
				participantID = "unknown"
			}
			participantMap[participantID] = append(participantMap[participantID], point)
		}

		for participant, points := range participantMap {
			subDataset := &types.Dataset{
				Points:  points,
				Columns: dataset.Columns,
			}
			stats, err := computeColumnStats(subDataset, config.AnalyzeColumns, config)
			if err != nil {
				return nil, fmt.Errorf("failed to compute stats for participant %s: %v", participant, err)
			}
			report.ParticipantStats[participant] = stats
		}
	}

	if !config.ByCondition && !config.ByParticipant {
		stats, err := computeColumnStats(dataset, config.AnalyzeColumns, config)
		if err != nil {
			return nil, fmt.Errorf("failed to compute overall stats: %v", err)
		}
		report.OverallStats = stats
	}

	return report, nil
}

func computeColumnStats(dataset *types.Dataset, columns []string, config StatsConfig) ([]ColumnStats, error) {
	var statsList []ColumnStats

	for _, col := range columns {
		values := extractColumnValues(dataset.Points, col)
		if len(values) == 0 {
			continue
		}

		stats := ColumnStats{
			Column: col,
			Count:  len(values),
			Min:    math.Inf(1),
			Max:    math.Inf(-1),
		}

		var sum, sumSq float64
		for _, v := range values {
			if math.IsNaN(v) {
				stats.MissingCount++
				continue
			}
			sum += v
			sumSq += v * v
			if v < stats.Min {
				stats.Min = v
			}
			if v > stats.Max {
				stats.Max = v
			}
		}

		stats.Mean = sum / float64(stats.Count-stats.MissingCount)

		sortedValues := make([]float64, 0, len(values)-stats.MissingCount)
		for _, v := range values {
			if !math.IsNaN(v) {
				sortedValues = append(sortedValues, v)
			}
		}
		sort.Float64s(sortedValues)
		mid := len(sortedValues) / 2
		if len(sortedValues)%2 == 0 {
			stats.Median = (sortedValues[mid-1] + sortedValues[mid]) / 2
		} else {
			stats.Median = sortedValues[mid]
		}

		variance := (sumSq / float64(stats.Count-stats.MissingCount)) - (stats.Mean * stats.Mean)
		stats.StdDev = math.Sqrt(variance)

		// Outlier detection using Z-score method
		if stats.StdDev > 0 {
			zThreshold := 3.0 // Common threshold
			stats.OutlierMethod = "z-score"
			stats.ZScoreThreshold = zThreshold

			for _, v := range values {
				if !math.IsNaN(v) {
					zScore := math.Abs((v - stats.Mean) / stats.StdDev)
					if zScore > zThreshold {
						stats.OutlierCount++
					}
				}
			}
		}

		statsList = append(statsList, stats)
	}

	return statsList, nil
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

func (r *StatsReport) String() string {
	var sb strings.Builder

	if len(r.OverallStats) > 0 {
		sb.WriteString("Overall Statistics:\n")
		for _, stats := range r.OverallStats {
			sb.WriteString(fmt.Sprintf("Column: %s\n", stats.Column))
			sb.WriteString(fmt.Sprintf("  Mean: %.4f\n", stats.Mean))
			sb.WriteString(fmt.Sprintf("  Median: %.4f\n", stats.Median))
			sb.WriteString(fmt.Sprintf("  StdDev: %.4f\n", stats.StdDev))
			sb.WriteString(fmt.Sprintf("  Min: %.4f\n", stats.Min))
			sb.WriteString(fmt.Sprintf("  Max: %.4f\n", stats.Max))
			sb.WriteString(fmt.Sprintf("  Count: %d\n", stats.Count))
			sb.WriteString(fmt.Sprintf("  MissingCount: %d\n", stats.MissingCount))
			sb.WriteString(fmt.Sprintf("  OutlierCount: %d\n", stats.OutlierCount))
			sb.WriteString(fmt.Sprintf("  OutlierMethod: %s\n", stats.OutlierMethod))
			sb.WriteString(fmt.Sprintf("  ZScoreThreshold: %.2f\n", stats.ZScoreThreshold))
		}
		sb.WriteString("\n")
	}

	if len(r.ConditionStats) > 0 {
		sb.WriteString("Statistics by Condition:\n")
		// Sort conditions for consistent output
		conditions := make([]string, 0, len(r.ConditionStats))
		for condition := range r.ConditionStats {
			conditions = append(conditions, condition)
		}
		sort.Strings(conditions)

		for _, condition := range conditions {
			stats := r.ConditionStats[condition]
			sb.WriteString(fmt.Sprintf("Condition: %s\n", condition))
			for _, colStats := range stats {
				sb.WriteString(fmt.Sprintf("  Column: %s\n", colStats.Column))
				sb.WriteString(fmt.Sprintf("    Mean: %.4f\n", colStats.Mean))
				sb.WriteString(fmt.Sprintf("    Median: %.4f\n", colStats.Median))
				sb.WriteString(fmt.Sprintf("    StdDev: %.4f\n", colStats.StdDev))
				sb.WriteString(fmt.Sprintf("    Min: %.4f\n", colStats.Min))
				sb.WriteString(fmt.Sprintf("    Max: %.4f\n", colStats.Max))
				sb.WriteString(fmt.Sprintf("    Count: %d\n", colStats.Count))
				sb.WriteString(fmt.Sprintf("    MissingCount: %d\n", colStats.MissingCount))
				sb.WriteString(fmt.Sprintf("    OutlierCount: %d\n", colStats.OutlierCount))
				sb.WriteString(fmt.Sprintf("    OutlierMethod: %s\n", colStats.OutlierMethod))
				sb.WriteString(fmt.Sprintf("    ZScoreThreshold: %.2f\n", colStats.ZScoreThreshold))
			}
			sb.WriteString("\n")
		}
	}

	if len(r.ParticipantStats) > 0 {
		sb.WriteString("Statistics by Participant:\n")
		// Sort participants for consistent output
		participants := make([]string, 0, len(r.ParticipantStats))
		for participant := range r.ParticipantStats {
			participants = append(participants, participant)
		}
		sort.Strings(participants)

		for _, participant := range participants {
			stats := r.ParticipantStats[participant]
			sb.WriteString(fmt.Sprintf("Participant: %s\n", participant))
			for _, colStats := range stats {
				sb.WriteString(fmt.Sprintf("  Column: %s\n", colStats.Column))
				sb.WriteString(fmt.Sprintf("    Mean: %.4f\n", colStats.Mean))
				sb.WriteString(fmt.Sprintf("    Median: %.4f\n", colStats.Median))
				sb.WriteString(fmt.Sprintf("    StdDev: %.4f\n", colStats.StdDev))
				sb.WriteString(fmt.Sprintf("    Min: %.4f\n", colStats.Min))
				sb.WriteString(fmt.Sprintf("    Max: %.4f\n", colStats.Max))
				sb.WriteString(fmt.Sprintf("    Count: %d\n", colStats.Count))
				sb.WriteString(fmt.Sprintf("    MissingCount: %d\n", colStats.MissingCount))
				sb.WriteString(fmt.Sprintf("    OutlierCount: %d\n", colStats.OutlierCount))
				sb.WriteString(fmt.Sprintf("    OutlierMethod: %s\n", colStats.OutlierMethod))
				sb.WriteString(fmt.Sprintf("    ZScoreThreshold: %.2f\n", colStats.ZScoreThreshold))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func SaveReport(report *StatsReport, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(report.String())
	if err != nil {
		return fmt.Errorf("failed to write report to file: %v", err)
	}

	return nil
}
