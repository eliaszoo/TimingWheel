package timing_wheel_test

import (
	"github.com/eliaszoo/TimingWheel"

	"fmt"
	"testing"
	"time"
)

func TestTimingWheel(t *testing.T) {
	tw, _ := timing_wheel.NewTimingWheel(time.Millisecond, 20)
	tw.Run()
	defer tw.Stop()

	durations := []int {1, 1, 5, 5, 6, 10, 20}
	for _, d := range durations {
		t.Run("", func(t *testing.T) {
			tw.AfterFunc(time.Duration(d) * time.Millisecond, func() {
				fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
			})
		})
	}
}