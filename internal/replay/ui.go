package replay

// Use Fyne to create a simple UI for replaying eye gaze data

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mbdvr/internal/types"
)

func StartUI(dataset *types.Dataset, speed float64) {
	a := app.New()
	w := a.NewWindow("Eye Gaze Data Replay")

	//Dropdowns for selecting the x and y gaze.
	xGazeSelect := widget.NewSelect(dataset.Columns, func(selected string) {})
	xGazeSelect.PlaceHolder = "Select X Gaze Column"
	yGazeSelect := widget.NewSelect(dataset.Columns, func(selected string) {})
	yGazeSelect.PlaceHolder = "Select Y Gaze Column"

	//Slider for speed control.
	speedSlider := widget.NewSlider(0.1, 5.0)
	speedSlider.Value = speed
	speedLabel := widget.NewLabel("Speed: 1.0x")
	speedSlider.OnChanged = func(value float64) {
		speedLabel.SetText("Speed: " + strconv.FormatFloat(value, 'f', 1, 64) + "x")
	}

	//Canvas for displaying the eye gaze position.
	canvas := widget.NewLabel("Eye Gaze Position")
	startButton := widget.NewButton("Start", func() {
		if xGazeSelect.Selected == "" || yGazeSelect.Selected == "" {
			canvas.SetText("Please select both X and Y gaze columns.")
			return
		}
		go replayData(dataset, xGazeSelect.Selected, yGazeSelect.Selected, speedSlider.Value, canvas)
	})
	stopButton := widget.NewButton("Stop", func() {
		// Implement stop functionality if needed.
	})

	w.SetContent(container.NewVBox(
		xGazeSelect,
		yGazeSelect,
		speedLabel,
		speedSlider,
		startButton,
		stopButton,
		canvas,
	))

	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}

func replayData(dataset *types.Dataset, xCol, yCol string, speed float64, canvas *widget.Label) {
	if dataset == nil || len(dataset.Points) == 0 {
		canvas.SetText("No data to replay.")
		return
	}

	startTime := dataset.Points[0].Timestamp
	for i, point := range dataset.Points {
		// Calculate the time to wait before showing the next point
		var waitTime float64
		if i == 0 {
			waitTime = 0
		} else {
			timeDiff := point.Timestamp - dataset.Points[i-1].Timestamp
			waitTime = timeDiff / speed
		}

		time.Sleep(time.Duration(waitTime*1000) * time.Millisecond)

		xGaze, xOk := point.Data[xCol]
		yGaze, yOk := point.Data[yCol]

		if !xOk || !yOk || xGaze == -1 || yGaze == -1 {
			canvas.SetText("No valid gaze data at time: " + strconv.FormatFloat(point.Timestamp-startTime, 'f', 2, 64))
		} else {
			canvas.SetText("Time: " + strconv.FormatFloat(point.Timestamp-startTime, 'f', 2, 64) +
				"\nX Gaze: " + strconv.FormatFloat(xGaze, 'f', 2, 64) +
				"\nY Gaze: " + strconv.FormatFloat(yGaze, 'f', 2, 64))
		}
	}

	canvas.SetText("Replay finished.")
}
