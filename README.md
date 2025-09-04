# MBDVR - Multi-condition Behavioral Data from Virtual Reality CLI

A command-line tool for processing and analyzing eye-tracking data collected from VR headset experiments across different experimental conditions.

## Quick Start

```bash
# Load your VR eye-tracking data from different conditions
mbdvr load --pattern "Boring*.csv" --condition boring --output boring.csv
mbdvr load --pattern "Interesting*.csv" --condition interesting --output interesting.csv

# Clean the data (remove outliers, handle missing values)
mbdvr clean --input boring.csv --output boring_clean.csv --remove-outliers --required "gaze_x,gaze_y"

# Clip specific time segments 
mbdvr clip --input session.csv --output intro.csv --start 0 --end 30

# Analyze and compare conditions
mbdvr stats --inputs "boring.csv,interesting.csv,clinical.csv" --analyze "gaze_x,gaze_y,pupil_size"

# Visual replay of gaze patterns
mbdvr replay --input cleaned_data.csv
```

## Philosophy

MBDVR is designed for **VR eye-tracking research** where you need to compare behavioral data across different experimental conditions. Whether you're studying attention in virtual environments, presence effects, or cognitive load during VR experiences, MBDVR handles the data processing so you can focus on the research.

The tool is intended to be **experiment-agnostic** - instead of being hardcoded for specific conditions, it lets you specify:
- **Any condition names** you want (`boring`, `engaging`, `stressful`, `relaxing`, `bright`, `dark`, etc.)
- **Any column names** from your CSV files  
- **Any experimental design** (2 conditions, 10 conditions, whatever you need)

The tool handles the data processing and analysis - you focus on your research.

## Installation

### Prerequisites
- Go 1.19+ (for data processing)

### Build from Source
```bash
git clone https://github.com/mbd888/mbdvr.git
cd mbdvr
go build -o mbdvr cmd/mbdvr/main.go
```

## Commands

### `load` - Load and Combine CSV Files

Loads CSV files matching a pattern and combines them into a single dataset.

```bash
mbdvr load --pattern "Boring*.csv" --condition boring --output boring.csv
```

**Options:**
- `--pattern` (required): File glob pattern (e.g., `"Happy*.csv"`, `"*.csv"`)
- `--condition` (required): Condition name to assign to all loaded data  
- `--output` (required): Output CSV file path

**Auto-Detection Features:**
- **Smart header detection**: Automatically finds where your data starts (assumes row 0 = headers, row 1+ = data)
- **Participant ID extraction**: Pulls participant IDs from filenames
- **Flexible column handling**: Works with any CSV column structure

### `clean` - Data Cleaning and Quality Control

Remove outliers, handle missing data, and filter low-quality tracking points.

```bash
mbdvr clean --input data.csv --output clean_data.csv --remove-outliers --required "gaze_x,gaze_y"
```

**Options:**
- `--input` (required): Input CSV file
- `--output` (required): Output cleaned CSV file
- `--required`: Comma-separated required columns
- `--remove-outliers`: Enable outlier detection and removal
- `--outlier-method`: Method for outlier detection (`iqr` or `zscore`)
- `--max-missing`: Maximum percentage of missing data per row (0-100)
- `--z-threshold`: Z-score threshold for outlier detection (default: 3.0)

### `clip` - Temporal Data Segmentation

Extract specific time segments from your VR sessions.

```bash
mbdvr clip --input session.csv --output intro_phase.csv --start 0 --end 30
```

**Options:**
- `--input` (required): Input CSV file
- `--output` (required): Output clipped CSV file  
- `--start` (required): Start time in seconds
- `--end` (required): End time in seconds

**Features:**
- **Closest frame matching**: Finds actual data points nearest to requested times
- **Duration reporting**: Shows actual vs requested time ranges
- **Retention statistics**: Reports how much data was kept

### `stats` - Statistical Analysis

Compute descriptive statistics and compare conditions.

```bash
mbdvr stats --inputs "boring.csv,interesting.csv,clinical.csv" --analyze "gaze_x,gaze_y,pupil_size"
```

**Options:**
- `--inputs` (required): Comma-separated input CSV files  
- `--analyze` (required): Comma-separated columns to analyze
- `--by-condition`: Group statistics by experimental condition (default: true)
- `--by-participant`: Group statistics by participant (default: false)
- `--output`: Save detailed results to file

**Statistical Measures:**
- Descriptive statistics (mean, median, std dev, min/max, quartiles)
- Missing data counts
- Outlier detection and counts
- Condition-wise and participant-wise breakdowns

### `replay` - Visual Data Replay

Interactive GUI for visualizing gaze patterns over time.

```bash
mbdvr replay --input cleaned_data.csv
```

**Features:**
- **Interactive controls**: Start/stop, speed adjustment
- **Column selection**: Choose X/Y gaze columns from dropdown
- **Real-time visualization**: See gaze positions as they occurred
- **Speed control**: Replay at different speeds (0.1x to 5x)

## Data Format

MBDVR works with CSV files containing eye-tracking data. After processing, your data will have this structure:

```csv
timestamp,participant_id,condition,gaze_x,gaze_y,pupil_size,...
0.0,P001,boring,123.45,67.89,3.2,...
0.016,P001,boring,124.12,68.23,3.1,...
0.032,P002,interesting,145.67,45.32,2.9,...
```

The tool preserves all your original columns while adding:
- `timestamp`: Time from start of recording
- `participant_id`: Extracted from filename  
- `condition`: As specified in the load command

## Workflow Examples

### Basic VR Comparison Study
```bash
# 1. Load different VR conditions
mbdvr load --pattern "Boring*.csv" --condition boring --output boring.csv
mbdvr load --pattern "Interesting*.csv" --condition interesting --output interesting.csv

# 2. Clean the data
mbdvr clean --input boring.csv --output boring_clean.csv --remove-outliers --required "gaze_x,gaze_y"
mbdvr clean --input interesting.csv --output interesting_clean.csv --remove-outliers --required "gaze_x,gaze_y"

# 3. Analyze specific time segments
mbdvr clip --input boring_clean.csv --output boring_task.csv --start 30 --end 120
mbdvr clip --input interesting_clean.csv --output interesting_task.csv --start 30 --end 120

# 4. Compare conditions statistically  
mbdvr stats --inputs "boring_task.csv,interesting_task.csv" --analyze "gaze_x,gaze_y,pupil_size"

# 5. Visual inspection
mbdvr replay --input boring_task.csv
```

### Multi-phase VR Session Analysis
```bash
# Extract different phases of VR experience
mbdvr clip --input full_session.csv --output intro.csv --start 0 --end 45
mbdvr clip --input full_session.csv --output main_task.csv --start 45 --end 195  
mbdvr clip --input full_session.csv --output ending.csv --start 195 --end 240

# Clean each phase
mbdvr clean --input main_task.csv --output main_clean.csv --remove-outliers

# Analyze the main task phase
mbdvr stats --inputs "main_clean.csv" --analyze "gaze_x,gaze_y,pupil_size" --by-participant
```

## Troubleshooting

### Common Issues

**"No files found for pattern"**
```bash
# Make sure to use quotes around patterns with wildcards
mbdvr load --pattern "Boring*.csv" --condition boring --output boring.csv  # Correct
```

**"Column not found in data"**
```bash
# Check your exact column names (they're case-sensitive)
head -1 your_file.csv  # See actual column names
# Then use the exact names:
mbdvr stats --inputs "data.csv" --analyze "exact_column_name"
```

**"Row has incorrect number of columns"**
- Your CSV file may have inconsistent formatting
- Check for missing commas or extra delimiters in the raw data
- Consider cleaning the CSV file before processing

## Development Status

**Current Status**: Core functionality complete
- ✅ CSV loading with auto header detection
- ✅ Data cleaning and quality filtering
- ✅ Temporal clipping and segmentation  
- ✅ Descriptive statistical analysis
- ✅ Interactive visual replay (Fyne GUI)
- 🔄 Advanced statistical tests (planned)
- 🔄 Data visualization plots (planned)

## License

MIT License - see LICENSE file for details.

## Citation

If you use MBDVR in your research, please cite:

```bibtex
@software{mbdvr,
  title={MBDVR: Multi-condition Behavioral Data from Virtual Reality CLI},
  author={mbd888},
  year={2025},
  url={https://github.com/mbd888/mbdvr}
}
```
