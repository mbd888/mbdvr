package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"mbdvr/internal/types"
)

type Loader struct {
	Condition string
}

func (l *Loader) LoadFiles(pattern string) (*types.Dataset, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find files matching pattern %s: %v", pattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no files found matching pattern %s", pattern)
	}

	fmt.Printf("Found %d files matching pattern %s\n", len(matches), pattern)

	var allPoints []types.DataPoint
	var columns []string

	// Load each file and aggregate points
	for _, file := range matches {
		points, cols, err := l.loadSingleFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load file %s: %v", file, err)
		}

		// Set columns only once from the first file
		if len(columns) == 0 {
			columns = cols
		}

		allPoints = append(allPoints, points...)
	}

	dataset := &types.Dataset{
		Points:  allPoints,
		Columns: columns,
		Metadata: map[string]interface{}{
			"total_files":  len(matches),
			"total_points": len(allPoints),
		},
	}

	return dataset, nil
}

func (l *Loader) loadSingleFile(filePath string) ([]types.DataPoint, []string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV data: %v", err)
	}

	if len(records) < 2 {
		return nil, nil, fmt.Errorf("file %s has insufficient data", filePath)
	}

	headerRowIdx := 0
	dataStartIdx := 1

	// Extract headers
	headers := records[headerRowIdx]
	if len(headers) < 2 {
		return nil, nil, fmt.Errorf("file %s has insufficient columns", filePath)
	}

	// Assume first column is timestamp, rest are data columns
	dataCols := headers[1:]

	var points []types.DataPoint

	// Extract participant ID from filename (assuming format participantID_anything.csv)
	baseName := filepath.Base(filePath)
	participantID := strings.SplitN(baseName, "_", 2)[0]

	// Parse data rows
	for i, row := range records[dataStartIdx:] {
		if len(row) != len(headers) {
			return nil, nil, fmt.Errorf("row %d in file %s has incorrect number of columns", i+dataStartIdx+1, filePath)
		}

		timestamp, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid timestamp in row %d of file %s: %v", i+dataStartIdx+1, filePath, err)
		}

		point := types.DataPoint{
			Timestamp:     timestamp,
			Data:          make(map[string]float64),
			ParticipantID: participantID,
			Condition:     l.Condition,
		}

		//Convert all data columns to float64 if possible
		for j, col := range dataCols {
			if valStr := row[j+1]; valStr != "" {
				val, err := strconv.ParseFloat(valStr, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid data value in row %d, column %s of file %s: %v", i+dataStartIdx+1, col, filePath, err)
				}
				point.Data[col] = val
			}
		}

		points = append(points, point)
	}

	return points, headers, nil
}

func (l *Loader) SaveDatasetAsCSV(dataset *types.Dataset, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Write header
	header := append([]string{"timestamp", "participant_id", "condition"}, dataset.Columns...)

	//Skip first column from dataset.Columns if it's timestamp
	if len(dataset.Columns) > 0 && dataset.Columns[0] == "timestamp" {
		header = append([]string{"timestamp", "participant_id", "condition"}, dataset.Columns[1:]...)
	}

	w.Write(header)

	// Write data points
	for _, point := range dataset.Points {
		row := make([]string, len(header))
		row[0] = fmt.Sprintf("%f", point.Timestamp)
		row[1] = point.ParticipantID
		row[2] = point.Condition

		for i, col := range dataset.Columns[1:] { // Skip timestamp column
			if val, ok := point.Data[col]; ok {
				row[i+3] = fmt.Sprintf("%f", val)
			} else {
				row[i+3] = ""
			}
		}

		w.Write(row)
	}

	return nil
}
