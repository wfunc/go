package util

import (
	"fmt"
	"testing"
	"time"
)

func TestNamedRunnerWithSeconds(t *testing.T) {
	var F = func() (err error) {
		fmt.Println("TestNamedRunnerWithSeconds done")
		return
	}
	running := true
	go NamedRunnerWithSeconds("TestNamedRunnerWithSeconds", 1, &running, F)
	time.Sleep(time.Second * 5)
	running = false
	time.Sleep(time.Second * 2)
}
