package workerpool

type Task struct {
	Err    error
	Data   interface{}
	Result interface{}
	f      func(interface{}) (interface{}, error)
}

func NewTask(f func(interface{}) (interface{}, error), data interface{}) *Task {
	return &Task{
		Data: data,
		f:    f,
	}
}

func process(workerID int, task *Task) {
	task.Result, task.Err = task.f(task.Data)
}
