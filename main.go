package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/trace"
)

func main() {
	out, _ := os.Create("trace.out")
	trace.Start(out)
	defer trace.Stop()

	runner := NewWorkGroup(8)
	defer close(runner.inputPipeline)

	command := "/usr/bin/md5sum"

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		runner.Run(command, scanner.Text())
	}
}

// WorkGroup is a container for possible
type WorkGroup struct {
	inputPipeline chan string
	size          int
}

// NewWorkGroup returns a workgroup and
func NewWorkGroup(maxWorkers int) WorkGroup {

	inputChannel := make(chan string)
	for i := 1; i <= maxWorkers; i++ {
		go Worker(inputChannel)
	}
	runner := WorkGroup{
		size:          maxWorkers,
		inputPipeline: inputChannel,
	}

	return runner
}

// Run starts the input for a single command
func (runner *WorkGroup) Run(command string, input string) {
	// fullCommand := fmt.Sprintf("%s '%s'", command, input)
	runner.inputPipeline <- input
}

// Worker is a single process running the command
func Worker(inputChannel <-chan string) {
	for input := range inputChannel {
		res, err := exec.Command("md5sum", input).Output()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Print(string(res))
	}
}
