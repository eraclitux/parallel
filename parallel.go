// Copyright (c) 2015 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

// Package parallel simplifies use of parallel
// (as not concurrent) workers that run on their own core.
// Number of workers is adjusted at runtime in base of numbers of cores.
// This paradigm is particularly useful in presence of heavy,
// independent tasks.
package parallel

// NOTE Useful for debugging on Linux: pidstat -tu  -C '<pid-name>'  1

import (
	"errors"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/eraclitux/trace"
)

// Tasker interface models an heavy task that have to be
// executed from a worker.
type Tasker interface {
	Execute()
}

// ErrTasksNotCompleted says that not all tasks where completed.
var ErrTasksNotCompleted = errors.New("SIGINT received, not all tasks have been completed")

var workersNumber int = runtime.NumCPU()

// Run starts the goroutines that will execute Taskers.
// It is intended to run blocking in the main goroutine.
func Run(jobs []Tasker) (err error) {
	// []T does not convert to []Tasker implicitly even is T implements
	// Tasker. We need to iterate on []Tasker making an explicit cast.
	// http://golang.org/doc/faq#convert_slice_of_interface
	prematureEnd := make(chan struct{})
	jobsQueue := make(chan Tasker, workersNumber)
	done := make(chan struct{}, workersNumber)
	var totalDone int
	go populateQueue(jobsQueue, jobs, prematureEnd)
	go parallelizeWorkers(jobsQueue, done)
	// TODO add a case timeout that returns error.
	for {
		select {
		case <-done:
			totalDone++
		case <-prematureEnd:
			err = ErrTasksNotCompleted
		}
		if totalDone == workersNumber {
			// We can assume that jobsQueue is closed and
			// that no goroutine is operating on []Tasker.
			break
		}
	}
	return
}

func populateQueue(jobsQueue chan<- Tasker, jobs []Tasker, prematureEnd chan<- struct{}) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	for _, t := range jobs {
		select {
		default:
			jobsQueue <- t
		case <-signalChan:
			// Abort jobs queue evaluation.
			// Taskers already sended will be finished
			// and an error will be returned.
			trace.Println("parallel: received SIGINT")
			prematureEnd <- struct{}{}
			close(jobsQueue)
			return
		}
	}
	trace.Println("close jobsQueue")
	close(jobsQueue)
}

// parallelizeWorkers creates a goroutine for every worker
// which will call Execute() method.
func parallelizeWorkers(jobsQueue <-chan Tasker, doneChan chan<- struct{}) {
	for i := 0; i < workersNumber; i++ {
		go evaluateQueue(jobsQueue, doneChan)
	}
}

// evaluateQueue does jobs in sequence on its own goroutine
// on a single core.
func evaluateQueue(jobsQueue <-chan Tasker, doneChan chan<- struct{}) {
	for j := range jobsQueue {
		j.Execute()
	}
	doneChan <- struct{}{}
}

// runSync is used to compare benchmark of parallelism
// implemented with channels.
func runSync(jobs []Tasker) (err error) {
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			j.Execute()
		}()
	}
	wg.Wait()
	return nil
}

// TODO has a non blocking version a sense (API semplification, performance etc.)? Es:
// When using Run one must wait that all tasks are done
// and put separate results togherther in the end. RunNonBlocking avoids that.
// func RunNonBlocking(jobs <-chan Tasker) (results chan<- Resulter) {
//code
//code
// Comunicate to callers that we are done.
// close(results)
//}
