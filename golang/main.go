package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}

func main() {
	const workerCount = 5
	const taskCount = 50

	taskChannel := make(chan Ttype, taskCount)
	doneTasks := make(chan Ttype, taskCount)
	undoneTasks := make(chan error, taskCount)

	// создание задач. в README.md представлен вариант без счетчика
	go func() {
		for i := 0; i < taskCount; i++ {
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 {
				ft = "Some error occured"
			}
			// taskChannel <- Ttype{cT: ft, id: int(time.Now().Unix())}
			taskChannel <- Ttype{cT: ft, id: i} // id = i
		}
		close(taskChannel)
	}()

	// пул воркеров
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChannel {
				processTask(task, doneTasks, undoneTasks)
			}
		}()
	}

	wg.Wait()
	close(doneTasks)
	close(undoneTasks)

	// Обработка результатов
	result := map[int]Ttype{}
	var errors []error

	for r := range doneTasks {
		result[r.id] = r
	}

	for r := range undoneTasks {
		errors = append(errors, r)
	}

	fmt.Println("Done tasks:")
	for _, r := range result {
		fmt.Println(r.id, r.fT)
	}

	fmt.Println("Errors:")
	for _, r := range errors {
		fmt.Println(r)
	}
}

func processTask(t Ttype, doneTasks chan<- Ttype, undoneTasks chan<- error) {
	createdTime, err := time.Parse(time.RFC3339, t.cT)
	if err != nil {
		t.taskRESULT = []byte("something went wrong")
	} else if createdTime.After(time.Now().Add(-20 * time.Second)) {
		t.taskRESULT = []byte("task has been successed")
	} else {
		t.taskRESULT = []byte("something went wrong")
	}

	t.fT = time.Now().Format(time.RFC3339Nano)
	time.Sleep(time.Millisecond * 150)

	if string(t.taskRESULT[14:]) == "successed" {
		doneTasks <- t
	} else {
		undoneTasks <- fmt.Errorf("task id %d time %s, error %s", t.id, t.cT, t.taskRESULT)
	}
}
