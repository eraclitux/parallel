// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package parallel

import (
	"math"
	"runtime"
	"testing"
	"time"
)

type dummy struct {
	done bool
}

type dummyNop struct {
	done bool
}

// isPrime returns true if a given int is prime.
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

func (d *dummy) Execute() {
	for i := 0; i < 1e4; i++ {
		isPrime(uint64(i))
	}
	d.done = true
}

func (d dummyNop) Execute() {
	for i := 0; i < 1e4; i++ {
		isPrime(uint64(i))
	}
	d.done = true
}

var testCases = make([]Tasker, 1e2)

func initTests() {
	// []*dummy does not convert []Tasker.
	// We need to iterate on []Tasker making an explicit cast.
	// http://golang.org/doc/faq#convert_slice_of_interface
	for i := range testCases {
		testCases[i] = Tasker(&dummy{})
	}

}

func TestRun(t *testing.T) {
	initTests()
	err := Run(testCases)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range testCases {
		if !e.(*dummy).done {
			t.Fatal("task not executed")
		}
	}
}
func TestRunSync(t *testing.T) {
	initTests()
	err := runSync(testCases)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range testCases {
		if !e.(*dummy).done {
			t.Error("error executing task")
		}
	}
}

func BenchmarkChannels(b *testing.B) {
	initTests()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Run(testCases)
	}
}
func BenchmarkSync(b *testing.B) {
	initTests()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runSync(testCases)
	}
}

// TestRun_nopointer shows that Execute() method
// must be implemented on a pointer receiver or computed values
// will be lost.
func TestRun_nopointer(t *testing.T) {
	tasks := make([]Tasker, 1e1)
	for i := range tasks {
		tasks[i] = Tasker(dummyNop{})
	}
	err := Run(tasks)
	if err != nil {
		t.Fatal("Test has failed", err)
	}
	for _, e := range tasks {
		if e.(dummyNop).done {
			t.Fatal("Error, receiver modified!")
		}
	}
}

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

// Verify that using parallel is faster than a serial execution
// considering also setup time.
func TestGain(t *testing.T) {
	cores := runtime.NumCPU()
	tasks := make([]Tasker, 0, cores)
	prev := 1
	// Bigger number to check.
	var limit int = 1e6
	pre := time.Now()
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
			tasks = append(tasks, Tasker(j))
		}
	}
	// Do not forget last interval.
	j := &job{start: prev, stop: limit}
	tasks = append(tasks, Tasker(j))
	// Run tasks in parallel using all cores.
	err := Run(tasks)
	if err != nil {
		t.Fatal(err)
	}
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
	if Δt2 < Δt1 {
		t.Error("using parallel takes more time")
	}
	t.Logf("%30s %9dns\n", "time with \"parallel\":", Δt1)
	t.Logf("%30s %9dns\n", "time without \"parallel\":", Δt2)
	diff := Δt2 - Δt1
	p := (100 * diff) / Δt2
	t.Logf("gain: %9d%%\n", p)
}
