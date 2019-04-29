package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"time"
)

//struct to send the opcode, its order & the thread id that it was processed on
type channelTransaction struct {
	opcode, order, pipe int
}

const (
	//how long to loop for debug purposes
	opnum = 10000
)

//creates some arrays for comparison
var inputData = []channelTransaction{}
var outputData = []channelTransaction{}
var unsortedOutputData = []channelTransaction{}
var retiredData = []channelTransaction{}

//incrementors for processing
var opcodesSent = 0
var opcodesRetired = 0
var opcodesRecieved = 0

func main() {
	//create unidirectional quit and quit callback channels
	quitChannel := make(chan bool)
	quitCompleteChannel := make(chan bool)
	//starts the main dispatch thread
	go dispatch(quitChannel, quitCompleteChannel)

	fmt.Print("\nJosh Windsor (24008182) Opcode Processor")
	fmt.Print("\n  Press q to quit: ")

	//creates a reader to wait for an input
	scanner := bufio.NewScanner(os.Stdin)
MainLoop:
	//waits for key entry
	for scanner.Scan() {
		//if the key is q then quit
		if scanner.Text() == "q" {
			quitChannel <- true
			//wait for quit to complete before breaking
			<-quitCompleteChannel
			break MainLoop
		}
	}
	time.Sleep(time.Millisecond * 100)
	fmt.Print("\nProcessing Status: Complete")
}

//main dispatch thread
//@Param quitChannel - input channel from main thread to stop running
//@Param quitCompleteChannel - output channel to main when running has stopped
func dispatch(quitChannel <-chan bool, quitCompleteChannel chan<- bool) {
	//creates an input & output channel arrays for the pipelines
	inputChannels := []chan channelTransaction{}
	outputChannels := []chan channelTransaction{}
	//creates bool channel to exit out the threads
	threadExitChannel := make(chan bool)

	//spawns the number of threads for the processor minus main & this thread
	threadsAvailable := runtime.NumCPU() - 2
	fmt.Print("\nThreads Available: " + strconv.Itoa(threadsAvailable))
	for i := 0; i < threadsAvailable; i++ {
		//creates new channels for each thread
		newInputChan := make(chan channelTransaction)
		newOutputChan := make(chan channelTransaction)
		inputChannels = append(inputChannels, newInputChan)
		outputChannels = append(outputChannels, newOutputChan)
		go pipeline(newInputChan, newOutputChan, threadExitChannel, i)
	}

	retireExitChannel := make(chan bool)
	go retire(outputChannels, retireExitChannel, threadsAvailable)

	var threadIterator = 0
	//var to skip wait for q to shutdown
	hardQuit := false
	//bool to generate new instruction
	generateOpcode := true
	//next opcode to use
	randChannelData := channelTransaction{}
	//tag for quit out break
OuterLoop:
	//loops for the max number for debug (this would be inf on actual system)
	//Seperate go routines could be used for the generation & retiring of opcodes
	for opcodesSent < opnum {
		if generateOpcode {
			//creates a random opcode to send
			randChannelData = channelTransaction{rand.Intn(5) + 1, opcodesSent, 0}
			generateOpcode = false
		}
		select {
		//quits the thread and processes remaining opcodes
		case <-quitChannel:
			retireExitChannel <- true
			hardQuit = true
			break OuterLoop
		//sends the new opcode to a waiting thread
		case inputChannels[threadIterator] <- randChannelData:
			//stores the sent opcode for checking later
			inputData = append(inputData, randChannelData)
			opcodesSent++
			generateOpcode = true
		default:
			threadIterator++
			if threadIterator == threadsAvailable {
				threadIterator = 0
			}
		}
	}

	//stops all the threads
	for index := 0; index < threadsAvailable; index++ {
		threadExitChannel <- true
	}

	//checks if user already pressed q
	if !hardQuit {
		//wait to display q for threads to finish
		time.Sleep(time.Millisecond * 100)
		fmt.Print("\n  Press q to quit: ")
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
	fmt.Print("\n  Thread " + strconv.Itoa(pipeNum) + " Status: Started")
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
	fmt.Print("\n  Thread " + strconv.Itoa(pipeNum) + " Status: Ended")
}

//retire & display thread
//@Param outputChannel - output channel from pipeline with completed opcode
//@Param retireExitChannel - channel to halt the thread
//@Param threadsAvailable - num threads to increment thread iterator
func retire(outputChannel []chan channelTransaction, retireExitChannel <-chan bool, threadsAvailable int) {
	threadIterator := 0
	exitBool := false
RetireExit:
	for {
		select {
		//if called to exit, jump out of label
		case <-retireExitChannel:
			exitBool = true
		//waits for a new piece of data
		case threadReturn := <-outputChannel[threadIterator]:
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
			//displays the live output of threads, could be in go routine but due to console clearing
			//needs to be concurrent or might clear output data
			formatOutput(inputData, unsortedOutputData, retiredData, exitBool)

			if exitBool && opcodesRetired == opcodesRecieved {
				break RetireExit
			}
		default:
			threadIterator++
			if threadIterator == threadsAvailable {
				threadIterator = 0
			}
		}
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

}

//formatting output to console
//@Param inputData - the opcodes in order of their start
//@Param outputData - the opcodes in order of their processing completion
//@Param retiredData - the opcodes in order of their retirement
func formatOutput(inputData []channelTransaction, outputData []channelTransaction, retiredData []channelTransaction, halting bool) {
	//Uses console commands to clear the console
	cmd := exec.Command("cmd", "/C", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()

	//displays all the lists in the correct format
	fmt.Print("    Input Opcodes: -")
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

	//displays to the user that the remaining codes are being finished
	if halting {
		fmt.Print("\nProcessing Status: Halting")
	}
}
