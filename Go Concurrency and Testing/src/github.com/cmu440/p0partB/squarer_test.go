// Tests for SquarerImpl. Students should write their code in this file.

package p0partB

import (
	"fmt"
	"testing"
	"time"
)

const (
	timeoutMillis = 5000
)

func TestBasicCorrectness(t *testing.T) {
	fmt.Println("Running TestBasicCorrectness.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	go func() {
		input <- 2
	}()
	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		t.Error("Test timed out.")
	case result := <-squares:
		if result != 4 {
			t.Error("Error, got result", result, ", expected 4 (=2^2).")
		}
	}
}

/*
	Test whether all resources are terminated after Close()
*/
func TestClose(t *testing.T) {
	// TODO: Write a test here.
	fmt.Println("Running TestClose.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)

	go func() {
		input <- 2
	}()

	sq.Close()

	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		fmt.Println("Test time's up.")
		return
	case <-squares:
		t.Error("Resources are not terminated")
	}
}

/*
	Test whether the squares are output in the right order
*/
func TestOutputOrder(t *testing.T) {
	// TODO: Write a test here.
	fmt.Println("Running TestOutputOrder.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)

	go func() {
		input <- 2
		input <- 1
		input <- 0
	}()

	result1 := <-squares
	result2 := <-squares
	result3 := <-squares
	if result1 == 4 {
		if result2 == 1 {
			if result3 == 0 {
				return
			}
		}
	}
	t.Error("The squares are not output in the right order")
}

/*
	Test zero input
*/
func TestZeroInput(t *testing.T) {
	// TODO: Write a test here.
	fmt.Println("Running TestZeroInput.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)

	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		fmt.Println("Test time's up.")
		return
	case <-squares:
		t.Error("Unexpected output")
	}
}
