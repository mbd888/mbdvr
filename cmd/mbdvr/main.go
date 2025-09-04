package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"mbdvr/internal/cleaner"
	"mbdvr/internal/clipper"
	"mbdvr/internal/loader"
	"mbdvr/internal/replay"
	"mbdvr/internal/stats"
	"mbdvr/internal/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mbdvr <command> [options]")
		fmt.Println("Commands: load | stats | replay | clean | clip")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "load":
		loadCommand()
	case "stats":
		statsCommand()
	case "replay":
		replayCommand()
	case "clean":
		cleanCommand()
	case "clip":
		clipCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func loadCommand() {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	pattern := fs.String("pattern", "", "File pattern to load (e.g. 'Boring*.csv' for 'Boring', '*.csv' for all CSVs) (required)")
	output := fs.String("output", "", "Name your output CSV file (required)")
	condition := fs.String("condition", "", "Condition name for the dataset (default: null)")

	fs.Parse(os.Args[2:])

	if *pattern == "" || *output == "" {
		fs.Usage()
		fmt.Printf("Pattern and output are required fields.\n")
		fmt.Printf("Sample usage: mbdvr load --pattern 'Test1*.csv' --output 'output.csv' --condition 'Uninterested'\n")
		os.Exit(1)
	}

	fmt.Printf("Loading files: %s\n", *pattern)
	fmt.Printf("Output: %s\n", *output)
	fmt.Printf("Condition: %s\n", *condition)

	loader := &loader.Loader{
		Condition: *condition,
	}

	dataset, err := loader.LoadFiles(*pattern)
	if err != nil {
		fmt.Printf("Error loading files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d data points with %d columns\n",
		len(dataset.Points), len(dataset.Columns))

	err = loader.SaveDatasetAsCSV(dataset, *output)
	if err != nil {
		fmt.Printf("Error saving dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Dataset saved to %s\n", *output)
}

func replayCommand() {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	input := fs.String("input", "", "Input CSV file to replay (required)")

	fs.Parse(os.Args[2:])

	if *input == "" {
		fs.Usage()
		os.Exit(1)
	}

	loader := &loader.Loader{}
	dataset, err := loader.LoadFiles(*input)
	if err != nil {
		fmt.Printf("Error loading input file: %v\n", err)
		os.Exit(1)
	}

	replay.StartUI(dataset, 1.0)
}

func cleanCommand() {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	input := fs.String("input", "", "Input CSV file to clean (required)")
	output := fs.String("output", "", "Output cleaned CSV file (required)")
	requiredCols := fs.String("required", "", "Comma-separated list of required columns")
	removeOutliers := fs.Bool("remove-outliers", false, "Whether to remove outliers")
	outlierMethod := fs.String("outlier-method", "iqr", "Outlier detection method: 'iqr' or 'zscore'")
	maxMissing := fs.Float64("max-missing", 0.0, "Max % of missing data per row (0-100)")
	zThreshold := fs.Float64("z-threshold", 3.0, "Z-score threshold for outlier detection")

	fs.Parse(os.Args[2:])

	if *input == "" || *output == "" {
		fs.Usage()
		fmt.Printf("Input and output are required fields.\n")
		fmt.Printf("Sample usage: mbdvr clean --input 'data.csv' --output 'cleaned.csv' --required 'X_Gaze,Y_Gaze' --remove-outliers --outlier-method 'zscore' --max-missing 10 --z-threshold 3.0\n")
		os.Exit(1)
	}

	fmt.Printf("Cleaning data: %s → %s\n", *input, *output)

	loader := &loader.Loader{}
	dataset, err := loader.LoadFiles(*input)
	if err != nil {
		fmt.Printf("Error loading input file: %v\n", err)
		os.Exit(1)
	}

	var reqCols []string
	if *requiredCols != "" {
		reqCols = strings.Split(*requiredCols, ",")

		for i := range reqCols {
			reqCols[i] = strings.TrimSpace(reqCols[i])
		}
	}

	cleanConfig := cleaner.CleanConfig{
		RequiredColumns:   reqCols,
		RemoveOutliers:    *removeOutliers,
		OutlierMethod:     *outlierMethod,
		MaxMissingPercent: *maxMissing,
		ZScoreThreshold:   *zThreshold,
	}

	//Clean the data
	cleanedDataset, stats, err := cleaner.CleanDataset(dataset, cleanConfig)
	if err != nil {
		fmt.Printf("Error cleaning dataset: %v\n", err)
		os.Exit(1)
	}

	//Save cleaned dataset
	err = loader.SaveDatasetAsCSV(cleanedDataset, *output)
	if err != nil {
		fmt.Printf("Error saving cleaned dataset: %v\n", err)
		os.Exit(1)
	}

	//Print cleaning summary
	fmt.Printf("Cleaning complete. Original points: %d, Removed missing: %d, Removed outliers: %d, Final points: %d\n",
		stats.OriginalPoints, stats.RemovedMissing, stats.RemovedOutliers, stats.FinalPoints)
	fmt.Printf("Cleaned dataset saved to %s\n", *output)
}

func clipCommand() {
	fs := flag.NewFlagSet("clip", flag.ExitOnError)
	input := fs.String("input", "", "Input CSV file to clip")
	output := fs.String("output", "", "Output clipped CSV file")
	startTime := fs.Float64("start", -1.0, "Start time in seconds")
	endTime := fs.Float64("end", -1.0, "End time in seconds")

	fs.Parse(os.Args[2:])

	if *input == "" || *output == "" || *startTime < 0 || *endTime < 0 {
		fs.Usage()
		fmt.Printf("Input, output, start, and end are required fields.\n")
		fmt.Printf("Sample usage: mbdvr clip --input 'data.csv' --output 'clipped.csv' --start 10.0 --end 20.0\n")
		os.Exit(1)
	}

	fmt.Printf("Clipping data: %s → %s (%.2f to %.2f seconds)\n", *input, *output, *startTime, *endTime)

	loader := &loader.Loader{}
	dataset, err := loader.LoadFiles(*input)
	if err != nil {
		fmt.Printf("Error loading input file: %v\n", err)
		os.Exit(1)
	}

	clipConfig := clipper.ClipConfig{}

	if !math.IsNaN(*startTime) {
		clipConfig.StartTime = startTime
	}
	if !math.IsNaN(*endTime) {
		clipConfig.EndTime = endTime
	}

	// Perform clipping
	clippedDataset, info, err := clipper.ClipDataset(dataset, clipConfig)
	if err != nil {
		fmt.Printf("Error clipping data: %v\n", err)
		os.Exit(1)
	}

	// Save clipped dataset
	err = loader.SaveDatasetAsCSV(clippedDataset, *output)
	if err != nil {
		fmt.Printf("Error saving clipped dataset: %v\n", err)
		os.Exit(1)
	}

	// Print clipping summary
	fmt.Printf("Data clipped successfully!\n")
	fmt.Printf("Original: %d points (%.3fs to %.3fs, %s)\n",
		info.OriginalPoints,
		info.MinTimestamp,
		info.MaxTimestamp,
		clipper.FormatDuration(info.TotalDuration))

	fmt.Printf("Clipped: %d points (%.3fs to %.3fs, %s)\n",
		info.ClippedPoints,
		info.ActualStartTime,
		info.ActualEndTime,
		clipper.FormatDuration(info.ActualEndTime-info.ActualStartTime))

	if clipConfig.StartTime != nil || clipConfig.EndTime != nil {
		fmt.Printf("Requested range: %.3fs to %.3fs\n",
			getFloat64OrDefault(clipConfig.StartTime, info.MinTimestamp),
			getFloat64OrDefault(clipConfig.EndTime, info.MaxTimestamp))

		if clipConfig.StartTime != nil {
			diff := math.Abs(info.ActualStartTime - *clipConfig.StartTime)
			fmt.Printf("Start frame difference: %.3fs\n", diff)
		}
		if clipConfig.EndTime != nil {
			diff := math.Abs(info.ActualEndTime - *clipConfig.EndTime)
			fmt.Printf("End frame difference: %.3fs\n", diff)
		}
	}

	retentionPercent := float64(info.ClippedPoints) / float64(info.OriginalPoints) * 100
	fmt.Printf("Retained: %.1f%% of original data\n", retentionPercent)
	fmt.Printf("Saved to: %s\n", *output)
}

func getFloat64OrDefault(val *float64, def float64) float64 {
	if val != nil {
		return *val
	}
	return def
}

func statsCommand() {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	inputs := fs.String("inputs", "", "Comma-separated input CSV files (required)")
	analyzeColumns := fs.String("analyze", "", "Comma-separated columns to analyze (required)")
	byCondition := fs.Bool("by-condition", true, "Group statistics by condition")
	byParticipant := fs.Bool("by-participant", false, "Group statistics by participant")
	output := fs.String("output", "", "Output file for detailed results (optional)")

	fs.Parse(os.Args[2:])

	if *inputs == "" || *analyzeColumns == "" {
		fmt.Println("Error: --inputs and --analyze are required")
		fmt.Println("\nExample:")
		fmt.Println("  mbdvr stats --inputs \"boring.csv,interesting.csv\" --analyze \"gaze_x,gaze_y,pupil_size\"")
		fs.Usage()
		os.Exit(1)
	}

	inputFiles := strings.Split(*inputs, ",")
	for i := range inputFiles {
		inputFiles[i] = strings.TrimSpace(inputFiles[i])
	}

	columns := strings.Split(*analyzeColumns, ",")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	loader := &loader.Loader{}
	var allPoints []types.DataPoint
	var allColumns []string
	for _, file := range inputFiles {
		dataset, err := loader.LoadFiles(file)
		if err != nil {
			fmt.Printf("Error loading file %s: %v\n", file, err)
			os.Exit(1)
		}
		allPoints = append(allPoints, dataset.Points...)
		allColumns = append(allColumns, dataset.Columns...)
	}

	// Remove duplicate columns
	columnSet := make(map[string]struct{})
	for _, col := range allColumns {
		columnSet[col] = struct{}{}
	}
	uniqueColumns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		uniqueColumns = append(uniqueColumns, col)
	}

	dataset := &types.Dataset{
		Points:  allPoints,
		Columns: uniqueColumns,
	}

	statsConfig := stats.StatsConfig{
		ByCondition:    *byCondition,
		ByParticipant:  *byParticipant,
		AnalyzeColumns: columns,
	}

	report, err := stats.ComputeStats(dataset, statsConfig)
	if err != nil {
		fmt.Printf("Error computing statistics: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	if report.OverallStats != nil {
		fmt.Println("Overall Statistics:")
		for _, colStats := range report.OverallStats {
			fmt.Printf("Column: %s | Count: %d | Min: %.3f | Max: %.3f | Mean: %.3f | Median: %.3f | StdDev: %.3f\n",
				colStats.Column, colStats.Count, colStats.Min, colStats.Max, colStats.Mean, colStats.Median, colStats.StdDev)
		}
	}

	if len(report.ConditionStats) > 0 {
		fmt.Println("\nStatistics by Condition:")
		for condition, stats := range report.ConditionStats {
			fmt.Printf("Condition: %s\n", condition)
			for _, colStats := range stats {
				fmt.Printf("  Column: %s | Count: %d | Min: %.3f | Max: %.3f | Mean: %.3f | Median: %.3f | StdDev: %.3f\n",
					colStats.Column, colStats.Count, colStats.Min, colStats.Max, colStats.Mean, colStats.Median, colStats.StdDev)
			}
		}
	}

	if len(report.ParticipantStats) > 0 {
		fmt.Println("\nStatistics by Participant:")
		for participant, stats := range report.ParticipantStats {
			fmt.Printf("Participant: %s\n", participant)
			for _, colStats := range stats {
				fmt.Printf("  Column: %s | Count: %d | Min: %.3f | Max: %.3f | Mean: %.3f | Median: %.3f | StdDev: %.3f\n",
					colStats.Column, colStats.Count, colStats.Min, colStats.Max, colStats.Mean, colStats.Median, colStats.StdDev)
			}
		}
	}

	// Optionally save detailed report
	if *output != "" {
		err := stats.SaveReport(report, *output)
		if err != nil {
			fmt.Printf("Error saving report to %s: %v\n", *output, err)
			os.Exit(1)
		}
		fmt.Printf("\nDetailed report saved to %s\n", *output)
	}
}
