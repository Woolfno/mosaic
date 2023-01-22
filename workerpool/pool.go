package workerpool

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
	Tasks   []*Task
	Workers []*Worker

	concurrency   int
	collector     chan *Task
	runBackgroung chan bool
	wg            sync.WaitGroup
}

func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		collector:   make(chan *Task, 1000),
	}
}

func (p *Pool) AddTask(task *Task) {
	p.collector <- task
}

func (p *Pool) Run() {
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		worker.Start(&p.wg)
	}

	for i := range p.Tasks {
		p.collector <- p.Tasks[i]
	}
	close(p.collector)
	p.wg.Wait()
}

func (p *Pool) RunBackground() {
	go func() {
		for {
			fmt.Println("Waiting for tasks to come in ...")
			time.Sleep(10 * time.Second)
		}
	}()

	for i := 1; i < p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		p.Workers = append(p.Workers, worker)
		go worker.StartBackground()
	}

	for i := range p.Tasks {
		p.collector <- p.Tasks[i]
	}

	p.runBackgroung = make(chan bool)
	<-p.runBackgroung
}

func (p *Pool) Stop() {
	for i := range p.Workers {
		p.Workers[i].Stop()
	}
	p.runBackgroung <- true
}
