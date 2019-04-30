## Golang实现的层级时间轮定时器

---

## TODO
* 使用小根堆优化ticker, 降低CPU唤醒频率

## install
```
go get -u github.com/eliaszoo/TimingWheel
```

## eg
``` Golang
package main

import (
	"fmt"
	"time"
	"github.com/eliaszoo/TimingWheel"
)

func main() {
	// 构造一个层级时间轮，第一层以1ms为一tick，共20个槽，第二层以1ms * 20为一tick，以此类推
	tw, err := timing_wheel.NewTimingWheel(time.Millisecond, 20) 
	
	// 启动定时器
	tw.Run() 

	// 停止定时器
	defer tw.Stop()

	durations := []int {50, 100, 111, 111, 112, 113, 200, 1000}
	for _, d := range durations {
		// 添加定时器， 第一个参数为延迟时间， 第二个参数为回调函数
		tw.AfterFunc(time.Duration(d) * time.Millisecond, func() {
			fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
		})
	}

	select {}
}
```