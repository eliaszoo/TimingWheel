package timing_wheel

import (
	"time"
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

type TimingWheel struct {
	ticker 		time.Ticker
	tickMs 		int
	wheelSize	int
	slots 		[]*slot
	slotIndex 	int
	tick 		*time.Ticker
	prevWheel 	*TimingWheel
	nextWheel 	unsafe.Pointer
	level 		int
	exit 		chan struct{}
	sync.WaitGroup
}

func NewTimingWheel(tick time.Duration, wheelSize int) (*TimingWheel, error) {
	if wheelSize <= 0 {
		return nil, errors.New("wheel size should > 0")
	}

	tickMs := int(tick / time.Millisecond)
	if tickMs <= 0 {
		return nil, errors.New("tick should > 0")
	}

	return newTimingWheel(tickMs, wheelSize, 0, nil), nil
}

func newTimingWheel(tick, wheelSize, level int, prev *TimingWheel) *TimingWheel {
	slots := make([]*slot, wheelSize)
	for i := 0; i < wheelSize; i ++ {
		slots[i] = newSlot()
	}

	return &TimingWheel {
		tickMs: tick,
		wheelSize: wheelSize,
		slots: slots,
		prevWheel: prev,
		nextWheel: nil,
		level: level,
	}
}

func (tw *TimingWheel) Run() {
	tw.tick = time.NewTicker(time.Millisecond * time.Duration(tw.tickMs))

	tw.Add(1)
	go func() {
		for {
			select {
			case <- tw.tick.C:
				tw.advance()
			case <- tw.exit:
				break
			}
		}
		tw.Done()
	}()
}

func (tw *TimingWheel) advance() {
	slot := tw.slots[tw.slotIndex]
	if 0 == tw.level {
		slot.trigger()
	} else {
		timers := slot.getClear()
		curms := time.Now().UnixNano() / int64(time.Millisecond)
		for t := timers.Front(); t != nil; t = t.Next() {
			t := t.Value.(*timer)
			tw.prevWheel.addTimer(int(t.expiredTime - curms), t)
		}
	}

	tw.slotIndex = (tw.slotIndex + 1) % tw.wheelSize
	if 0 == tw.slotIndex {
		nextTw := atomic.LoadPointer(&tw.nextWheel)
		if nil != nextTw {
			(*TimingWheel)(nextTw).advance()
		}
	}
}

func (tw *TimingWheel) addTimer(duration int, t *timer) {
	index := (int(duration / tw.tickMs) + tw.slotIndex) % tw.wheelSize
	tw.slots[index].add(t)
}

func (tw *TimingWheel) AfterFunc(duration time.Duration, callback func()) {
	durationMs := int(duration / time.Millisecond)
	if durationMs > tw.wheelSize * tw.tickMs {
		nextWheel := atomic.LoadPointer(&tw.nextWheel)
		if nil == nextWheel {
			newTw := newTimingWheel(tw.tickMs * tw.wheelSize, tw.wheelSize, tw.level + 1, tw)
			atomic.CompareAndSwapPointer(&tw.nextWheel, nil, unsafe.Pointer(newTw))
			nextWheel = atomic.LoadPointer(&tw.nextWheel)
		}
		(*TimingWheel)(nextWheel).AfterFunc(duration, callback)
	} else {
		expriedTime := time.Now().UnixNano() / int64(time.Millisecond) + int64(durationMs)
		tw.addTimer(durationMs, &timer{expriedTime, callback})
	}
}

func (tw *TimingWheel) Stop() {
	close(tw.exit)
	tw.Wait()
}