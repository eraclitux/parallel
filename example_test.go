// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package parallel_test

import (
	"fmt"
	"math"
	"runtime"

	"github.com/eraclitux/parallel"
)

type job struct {
	start   int
	stop    int
	results map[int]bool
}

func (j *job) Execute() {
	j.results = make(map[int]bool)
	for i := j.start; i <= j.stop; i++ {
		j.results[i] = isPrime(uint64(i))
	}
}

func isPrime(n uint64) bool {
	if n <= 2 {
		return true
	}
	var i uint64
	i = 2
	num := uint64(math.Sqrt(float64(n)))
	for i <= num {
		if n%i == 0 {
			return false
		}
		i++
	}
	return true
}

// Example shows example usage of the package.
func Example() {
	cores := runtime.NumCPU()
	// Creates the slice of tasks that we want to execute in parallel.
	tasks := make([]parallel.Tasker, 0, cores)
	prev := 1
	// Bigger number to check.
	var limit int = 1e5
	// Create as much tasks as number of cores.
	d := int(limit / cores)
	for i := 1; i < limit; i++ {
		// This is not the best way to distribute load
		// as complexity is not the same in different
		// intervals (bigger numbers are more difficult to verify),
		// so some cores remains idle sooner.
		// We could increase efficiency making different interval lengths.
		if (i % d) == 0 {
			j := &job{start: prev, stop: i}
			prev = i + 1
			tasks = append(tasks, parallel.Tasker(j))
		}
	}
	// Do not forget last interval.
	j := &job{start: prev, stop: limit}
	tasks = append(tasks, parallel.Tasker(j))

	// Run tasks in parallel using all cores.
	err := parallel.RunBlocking(tasks)
	if err == nil {
		fmt.Println("Example OK")
	}

	// Output:
	// Example OK
}
