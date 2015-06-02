// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package goparallel_test

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/eraclitux/goparallel"
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

// FIXME remove time comparison here and use testing's test/benchmark.

// Example shows example usage of the package.
// It compares time consuption using single core.
func Example() {
	// Creates the slice of tasks that we want to execute in parallel.
	tasks := make([]goparallel.Tasker, 0, 1e3)
	prev := 1
	// Limit is the bigger number to check.
	var limit int = 1e5
	pre := time.Now()
	// Create as much tasks as number of cores.
	d := limit / runtime.NumCPU()
	for i := 1; i < limit; i++ {
		// This is not the best way to disbrubute load
		// as complexity is not the same in different
		// intervals (bigger numbers are more difficult to verify),
		// so some cores remains idle sooner.
		// We could increase efficency making different interval lenghts.
		if (i % d) == 0 {
			j := &job{start: prev, stop: i}
			prev = i + 1
			tasks = append(tasks, goparallel.Tasker(j))
		}
	}
	// Do not forget last interval.
	j := &job{start: prev, stop: limit}
	tasks = append(tasks, goparallel.Tasker(j))

	// Run tasks in parallel using all cores.
	goparallel.RunBlocking(tasks)

	after := time.Now()
	Δt1 := after.Sub(pre)

	// Lets compare execution time using single core.
	pre = time.Now()
	results := make(map[int]bool)
	for i := 1; i <= limit; i++ {
		results[i] = isPrime(uint64(i))
	}
	after = time.Now()
	Δt2 := after.Sub(pre)
	if Δt2 > Δt1 {
		fmt.Println("Using goparallel takes less time.")
	} else {
		fmt.Println("Using goparallel takes more time.")
	}
	// We use stderr as stdout is checked to pass test.
	fmt.Fprintf(os.Stderr, "%30s %9dns\n", "Time with goworker:", Δt1)
	fmt.Fprintf(os.Stderr, "%30s %9dns\n", "Time without goworker:", Δt2)

	// Output:
	// Using goparallel takes less time.
}
