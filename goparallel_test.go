// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package goparallel

import (
	"fmt"
	"math"
	"testing"
)

type dummy struct {
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

func TestRunBlocking(t *testing.T) {
	tasks := make([]Tasker, 1e2)
	// []*dummy does not convert []Tasker.
	// We need to iterate on []Tasker making an explicit cast.
	// http://golang.org/doc/faq#convert_slice_of_interface
	for i := range tasks {
		tasks[i] = Tasker(&dummy{false})
	}

	err := RunBlocking(tasks)
	if err != nil {
		t.Errorf("Test has failed")
	}
	for _, e := range tasks {
		if !e.(*dummy).done {
			fmt.Println("Error executig task")
			t.FailNow()
		}
	}
}

// TODO test what happen ia a receiver of Execute() in a concrete type.
// Is it modified into []Tasker?
//func TestRunBlocking_nopointer(t *testing.T) {}
