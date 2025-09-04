package replay

import (
	"fmt"
	"time"

	"mbdvr/internal/types"
)

type Replay struct {
	Dataset *types.Dataset
	Speed   float64 // Speed multiplier for replay
}

func (r *Replay) Start() error {
	if r.Dataset == nil || len(r.Dataset.Points) == 0 {
		return fmt.Errorf("no data to replay")
	}

	fmt.Println("Starting replay...")

	startTime := r.Dataset.Points[0].Timestamp
	for i, point := range r.Dataset.Points {
		// Calculate the time to wait before showing the next point
		var waitTime time.Duration
		if i == 0 {
			waitTime = 0
		} else {
			timeDiff := point.Timestamp - r.Dataset.Points[i-1].Timestamp
			waitTime = time.Duration(timeDiff/r.Speed*1000) * time.Millisecond
		}

		time.Sleep(waitTime)

		// Display the data point (for simplicity, just print it)
		fmt.Printf("Time: %.2f, Data: %v\n", point.Timestamp-startTime, point.Data)
	}

	fmt.Println("Replay finished.")
	return nil
}

func (r *Replay) SetSpeed(speed float64) {
	if speed <= 0 {
		speed = 1.0 // Default to normal speed if invalid
	}
	r.Speed = speed
}

func NewReplay(dataset *types.Dataset, speed float64) *Replay {
	return &Replay{
		Dataset: dataset,
		Speed:   speed,
	}
}
