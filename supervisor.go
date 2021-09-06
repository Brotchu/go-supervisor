package gosupervisor

import (
	"fmt"
	"reflect"
	"sync"
)

type Supervisor struct {
	Func reflect.Value
	wg   sync.WaitGroup
}

func NewSupervisor(f interface{}) (*Supervisor, error) {
	fval := reflect.ValueOf(f)
	if fval.Kind().String() != "func" {
		return nil, fmt.Errorf("expected func as argument")
	}
	return &Supervisor{
		Func: fval,
		wg:   sync.WaitGroup{},
	}, nil
}

func (s *Supervisor) Wait() {
	s.wg.Wait()
}

func (s *Supervisor) Run(args ...interface{}) error {

	taskChan := make(chan int)
	s.wg.Add(len(args))

	status := make(map[int]int)

	go func() {
		for {
			taskID := <-taskChan
			go func(idx int) {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println(r)
						if _, ok := status[idx]; !ok {
							status[idx] = 1
							taskChan <- idx
						} else {
							if status[idx] < 3 {
								status[idx] += 1
								taskChan <- idx
							} else {
								fmt.Println("terminated ", idx)
								s.wg.Done()
							}
						}
					}
				}()
				param := []reflect.Value{}
				param = append(param, reflect.ValueOf(args[idx]))
				s.Func.Call(param)
				s.wg.Done()
			}(taskID)
		}
	}()

	for i := range args {
		taskChan <- i
	}
	s.wg.Wait()
	return nil
}
