package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

var numProcesses int

func init() {
	flag.IntVar(&numProcesses, "j", runtime.NumCPU(), "Max number of coroutines to use")
	flag.Parse()
}

func main() {

	var command string
	if len(flag.Args()) > 0 {
		command = flag.Args()[0]
	} else {
		command = "/usr/bin/md5sum"
	}

	runner := NewWorkGroup(numProcesses, command)
	defer close(runner.inputPipeline)

	scanner := bufio.NewScanner(os.Stdin)

	runner.wg.Add(1)
	go func() {
		for scanner.Scan() {
			runner.wg.Add(1)
			runner.Run(scanner.Text())
		}
		runner.wg.Done()
	}()

	runner.wg.Wait()
}

// WorkGroup is a container for possible
type WorkGroup struct {
	wg            *sync.WaitGroup
	inputPipeline chan string
	size          int
}

// NewWorkGroup returns a workgroup and
func NewWorkGroup(maxWorkers int, command string) WorkGroup {

	inputChannel := make(chan string)

	runner := WorkGroup{
		wg:            &sync.WaitGroup{},
		size:          maxWorkers,
		inputPipeline: inputChannel,
	}

	for i := 1; i <= maxWorkers; i++ {
		go Worker(runner.wg, command, inputChannel)
	}

	return runner
}

// Run starts the input for a single command
func (runner *WorkGroup) Run(input string) {
	runner.inputPipeline <- input
}

// Worker is a single process running the command
func Worker(wg *sync.WaitGroup, command string, inputChannel <-chan string) {
	for input := range inputChannel {
		res, err := exec.Command(command, input).Output()
		if err != nil {
			log.Println(err)
		} else {
			fmt.Print(string(res))
		}
		wg.Done()
	}
}
