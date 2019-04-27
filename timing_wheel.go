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
	nextWheel 	unsafe.Pointer
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

	slots := make([]*slot, wheelSize)
	for i := 0; i < wheelSize; i ++ {
		slots[i] = newSlot()
	}

	return &TimingWheel {
		tickMs: tickMs,
		wheelSize: wheelSize,
		slots: slots,
		nextWheel: nil,
	}, nil
}

func (tw *TimingWheel) Start() {
	tw.tick = time.NewTicker(time.Millisecond * time.Duration(tw.tickMs))
	
}

func (tw *TimingWheel) advance() {
	slot := tw.slots[tw.slotIndex]
	slot.trigger()

	tw.slotIndex = (tw.slotIndex + 1) % tw.wheelSize
	if 0 == tw.slotIndex {

	}
}

func (tw *TimingWheel) AfterFunc(duration time.Duration, callback func()) {
	durationMs := int(duration / time.Millisecond)
	if durationMs > tw.wheelSize * tw.tickMs {
		nextWheel := atomic.LoadPointer(&tw.nextWheel)
		if nil == nextWheel {
			newTw, _ := NewTimingWheel(time.Duration(tw.tickMs * tw.wheelSize) * time.Millisecond, tw.wheelSize)
			atomic.CompareAndSwapPointer(&tw.nextWheel, nil, unsafe.Pointer(newTw))
			nextWheel = atomic.LoadPointer(&tw.nextWheel)
		}
		(*TimingWheel)(nextWheel).AfterFunc(duration, callback)
	} else {
		index := (int(durationMs / tw.tickMs) + tw.slotIndex) % tw.wheelSize
		tw.slots[index].add(timer{callback})
	}
}

