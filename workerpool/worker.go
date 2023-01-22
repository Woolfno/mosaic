package workerpool

import (
	"fmt"
	"sync"
)

type Worker struct {
	ID       int
	taskChan chan *Task
	quit     chan bool
}

func NewWorker(channel chan *Task, ID int) *Worker {
	return &Worker{
		ID:       ID,
		taskChan: channel,
		quit:     make(chan bool),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range w.taskChan {
			process(w.ID, task)
		}
	}()
}

func (w *Worker) StartBackground() {
	fmt.Printf("Starting worker %d\n", w.ID)

	for {
		select {
		case task := <-w.taskChan:
			process(w.ID, task)
		case <-w.quit:
			return
		}
	}
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
