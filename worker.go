package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, offset [2]int, works *Works) {
	ws := PlaceWs{Url: wsUrl, AutoReconnect: true}
	ws.Connect()

	fmt.Printf("Worker %d connected to websocket\n", id)

	for {
		time.Sleep(500 * time.Microsecond)

		work := works.Get()

		if work == nil {
			continue
		}

		fmt.Println("Worker", id, "working on", work.x, work.y)

		x, y := work.x, work.y
		r, g, b := work.r, work.g, work.b

		if err := ws.PlacePixel(x, y, r, g, b); err != nil {
			fmt.Println(err)
		}

	}

}

type Work struct {
	x, y    int
	r, g, b uint8
}

type Works struct {
	Queue []*Work

	Mutex *sync.Mutex
}

func (works *Works) Add(work *Work) {
	works.Mutex.Lock()
	defer works.Mutex.Unlock()

	works.Queue = append(works.Queue, work)
}

func (works *Works) Get() *Work {
	works.Mutex.Lock()
	defer works.Mutex.Unlock()

	if len(works.Queue) == 0 {
		return nil
	}

	work := works.Queue[0]
	works.Queue = works.Queue[1:]

	return work
}

func (works *Works) Compact() {
	works.Mutex.Lock()
	defer works.Mutex.Unlock()

	newQueue := make([]*Work, 0)

	for _, work := range works.Queue {
		if work == nil {
			continue
		}
		newQueue = append(newQueue, work)
	}

	works.Queue = newQueue
}
