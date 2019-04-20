package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"time"
)

//struct to send the opcode and its order
type channelTransaction struct {
	opcode, order, pipe int
}

const (
	//how long to loop for debug purposes
	opnum = 15
)

func main() {
	//create unidirectional quit and quit callback channels
	quitChannel := make(chan bool)
	quitCompleteChannel := make(chan bool)
	//starts the main dispatch thread
	go dispatch(quitChannel, quitCompleteChannel)

	fmt.Print("\nJosh Windsor (24008182) Opcode Processor")
	fmt.Print("\nPress q to quit")

	//creates a reader to wait for an input
	scanner := bufio.NewScanner(os.Stdin)
MainLoop:
	//waits for key entry
	for scanner.Scan() {
		//if the key is q then quit
		if scanner.Text() == "q" {
			fmt.Print("\nProcessing Status: Halting")
			quitChannel <- true
			//wait for quit to complete before breaking
			<-quitCompleteChannel
			break MainLoop
		}
	}

	fmt.Print("\nProcessing Status: Complete")

}

//main dispatch thread
//@Param quitChannel - input channel from main thread to stop running
//@Param quitCompleteChannel - output channel to main when running has stopped
func dispatch(quitChannel <-chan bool, quitCompleteChannel chan bool) {
	//creates an input & output channel for the pipelines
	inputChannel := make(chan channelTransaction)
	outputChannel := make(chan channelTransaction)
	//creates bool channel to exit out the threads
	threadExitChannel := make(chan bool)

	//spawns the number of threads for the processor minus main & this thread
	threadsAvailable := runtime.NumCPU() - 2
	fmt.Print("\nThreads Available: ")
	fmt.Print(threadsAvailable)
	for i := 0; i < threadsAvailable; i++ {
		go pipeline(inputChannel, outputChannel, threadExitChannel, i)
	}

	//creates some arrays for comparison
	inputData := []channelTransaction{}
	outputData := []channelTransaction{}
	unsortedOutputData := []channelTransaction{}
	retiredData := []channelTransaction{}
	//incrementors for processing
	opcodesSent := 0
	opcodesRetired := 0
	opcodesRecieved := 0
	//var to skip wait for q to shutdown
	hardQuit := false
	//tag for quit out break
OuterLoop:
	//loops for the max number for debug (this would be inf on actual system)
	for opcodesSent < opnum {
		//creates a random opcode to send
		randChannelData := channelTransaction{rand.Intn(4) + 1, opcodesSent, 0}
		select {
		//quits the thread and processes remaining opcodes
		case <-quitChannel:
			hardQuit = true
			break OuterLoop
		//sends the new opcode to a waiting thread
		case inputChannel <- randChannelData:
			//stores the sent opcode for checking later
			inputData = append(inputData, randChannelData)
			opcodesSent++
		//listens for return from a pipeline to retire
		case threadReturn := <-outputChannel:
			//stores the returned opcode for checking
			outputData = append(outputData, threadReturn)
			//updates the pipe on the input with the thread id
			inputData[threadReturn.order].pipe = threadReturn.pipe
			//stored to show execution order
			unsortedOutputData = append(unsortedOutputData, threadReturn)
			//sorts the array to retire in order
			sort.Slice(outputData, func(i, j int) bool {
				return outputData[i].order < outputData[j].order
			})
			opcodesRecieved++

			//if the next ordered opcode is finished then retire it and
			//loop through any above in the array to retire
			for opcodesRetired == outputData[opcodesRetired].order {
				retiredData = append(retiredData, outputData[opcodesRetired])
				opcodesRetired++
				if opcodesRetired == opcodesRecieved {
					break
				}
			}
			//displays the live output of threads, could be called as a goroutine
			formatOutput(inputData, unsortedOutputData, retiredData)
		default:
		}
	}

	//finishes off the remaining opcodes that have been processed (see comments above)
	for opcodesRecieved < opcodesSent {
		select {
		case threadReturn := <-outputChannel:
			outputData = append(outputData, threadReturn)
			inputData[threadReturn.order].pipe = threadReturn.pipe
			unsortedOutputData = append(unsortedOutputData, threadReturn)
			sort.Slice(outputData, func(i, j int) bool {
				return outputData[i].order < outputData[j].order
			})
			opcodesRecieved++
			for opcodesRetired == outputData[opcodesRetired].order {
				retiredData = append(retiredData, outputData[opcodesRetired])
				opcodesRetired++
				if opcodesRetired == opcodesRecieved {
					break
				}
			}
			formatOutput(inputData, unsortedOutputData, retiredData)
		default:
		}
	}

	//retires remaining opcodes
	for opcodesRetired < opcodesRecieved {
		retiredData = append(retiredData, outputData[opcodesRetired])
		//not calling as goroutine as needs to display final output
		formatOutput(inputData, unsortedOutputData, retiredData)
		opcodesRetired++
	}

	//compare arrays to check if correct output
	matching := true
	for i := 0; i < opcodesSent; i++ {
		if inputData[i] != retiredData[i] {
			matching = false
		}
	}
	fmt.Print("\n\n Arrays Matching?: ")
	fmt.Print(matching)
	fmt.Print("\n")
	fmt.Print("\nProcessing Status: Opcodes Complete")

	//stops all the threads
	threadExitChannel <- true
	//checks if user already pressed q
	if !hardQuit {
		fmt.Print("\nPress q to quit")
	QuitLoop:
		for {
			//waits for quit key to shutdown
			select {
			case <-quitChannel:
				break QuitLoop
			}
		}
	}
	//returns completion channel to shut down main thread
	quitCompleteChannel <- true

}

//main execution thread
//@Param inputChannel - input channel from dispatch with next opcode
//@Param outputChannel - output channel to dispatch with completed opcode
//@Param threadExitChannel - channel to halt the thread
//@Param pipeNum - the thread's Id for reporting
func pipeline(inputChannel <-chan channelTransaction, outputChannel chan<- channelTransaction, threadExitChannel <-chan bool, pipeNum int) {
	//Label to exit loop on completion
ThreadLoop:
	for {
		select {
		//if called to exit, jump out of label
		case <-threadExitChannel:
			break ThreadLoop
		//waits for a new piece of data
		case thread := <-inputChannel:
			//stores the id for display on completion
			thread.pipe = pipeNum
			//sleeps for the opcode duration
			time.Sleep(time.Second * time.Duration(thread.opcode))
			//returns the completed data
			outputChannel <- thread
		}
	}
}

//formatting output to console
//@Param inputData - the opcodes in order of their start
//@Param outputData - the opcodes in order of their processing completion
//@Param retiredData - the opcodes in order of their retirement
func formatOutput(inputData []channelTransaction, outputData []channelTransaction, retiredData []channelTransaction) {
	fmt.Print("\n\n\n    Input Opcodes: -")
	for index := 0; index < len(inputData); index++ {
		fmt.Print(inputData[index].opcode)
		fmt.Print("-")
	}
	fmt.Print("\n    Used Pipeline: -")
	for index := 0; index < len(inputData); index++ {
		fmt.Print(inputData[index].pipe)
		fmt.Print("-")
	}
	fmt.Print("\n\nProcessed Opcodes: -")
	for index := 0; index < len(outputData); index++ {
		fmt.Print(outputData[index].opcode)
		fmt.Print("-")
	}
	fmt.Print("\n    Used Pipeline: -")
	for index := 0; index < len(outputData); index++ {
		fmt.Print(outputData[index].pipe)
		fmt.Print("-")
	}
	fmt.Print("\n\n  Retired Opcodes: -")
	for index := 0; index < len(retiredData); index++ {
		fmt.Print(retiredData[index].opcode)
		fmt.Print("-")
	}

}
