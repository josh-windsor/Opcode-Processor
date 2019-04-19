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
	opcode, order int
}

func main() {
	//create unidirectional quit and quit callback channels
	quitChannel := make(chan bool)
	quitCompleteChannel := make(chan bool)
	//starts the main dispatch thread
	go dispatch(quitChannel, quitCompleteChannel)

	fmt.Print("\nJosh Windsor (24008182) Opcode Processor")
	fmt.Print("\nPress any key to quit")

	//creates a reader to wait for an input
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	//stops the program
	fmt.Print("\nHalting Processor")
	//sends quit to main thread to process & retire remaining opcodes
	quitChannel <- true
	//waits for the quit to complete before ending the main thread
	<-quitCompleteChannel
	fmt.Println("\nProcessing Complete")
}

const (
	//how long to loop for debug purposes
	opnum = 10000
)

func dispatch(quitChannel <-chan bool, quitCompleteChannel chan<- bool) {
	//creates an input & output channel for the pipelines
	inputChannel := make(chan channelTransaction)
	outputChannel := make(chan channelTransaction)

	//spawns the number of threads for the processor minus main & this thread
	threadsAvailable := runtime.NumCPU() - 2
	fmt.Print("\nThreads Available: ")
	fmt.Print(threadsAvailable)
	for i := 0; i < threadsAvailable; i++ {
		go pipeline(inputChannel, outputChannel)
	}

	//creates some arrays for comparison
	inputData := []channelTransaction{}
	outputData := []channelTransaction{}
	outputDataUnsorted := []channelTransaction{}
	//incrementors for processing
	opcodesSent := 0
	opcodesRetired := 0
	opcodesRecieved := 0

	//tag for quit out break
OuterLoop:
	//loops for the max number for debug (this would be inf on actual system)
	for opcodesSent < opnum {
		//creates a random opcode to send
		randChannelData := channelTransaction{rand.Intn(9), opcodesSent}
		select {
		//quits the thread and processes remaining opcodes
		case <-quitChannel:
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
			outputDataUnsorted = append(outputDataUnsorted, threadReturn)
			//sorts the array to retire in order
			sort.Slice(outputData, func(i, j int) bool {
				return outputData[i].order < outputData[j].order
			})
			//if the next ordered opcode is finished then retire it
			for opcodesRetired == outputData[opcodesRetired].order {
				fmt.Print("\nRetired: ")
				fmt.Print(outputData[opcodesRetired])
				opcodesRetired++
				if opcodesRetired == opcodesRecieved {
					break
				}
			}
			opcodesRecieved++
		default:
		}
	}

	//finishes off the remaining opcodes that have been processed
	for opcodesRecieved < opcodesSent {
		select {
		case threadReturn := <-outputChannel:
			outputData = append(outputData, threadReturn)
			outputDataUnsorted = append(outputDataUnsorted, threadReturn)
			sort.Slice(outputData, func(i, j int) bool {
				return outputData[i].order < outputData[j].order
			})
			for opcodesRetired == outputData[opcodesRetired].order {
				fmt.Print("\nRetired: ")
				fmt.Print(outputData[opcodesRetired])
				opcodesRetired++
				if opcodesRetired == opcodesRecieved {
					break
				}
			}
			opcodesRecieved++
		default:
		}
	}
	//retires remaining opcodes
	for opcodesRetired < opcodesRecieved {
		fmt.Print("\nRetired: ")
		fmt.Print(outputData[opcodesRetired])
		opcodesRetired++

	}
	//compare arrays to check if correct output
	matching := true
	for i := 0; i < opcodesSent; i++ {
		if inputData[i] != outputData[i] {
			matching = false
		}
	}
	fmt.Print("\nArrays Matching? : ")
	fmt.Print(matching)
	fmt.Print("\n")
	formatOutput(inputData, outputData, outputDataUnsorted)
	quitCompleteChannel <- true
}

func pipeline(inputChannel <-chan channelTransaction, outputChannel chan<- channelTransaction) {
	for {
		thread := <-inputChannel
		fmt.Print("\nStarted: ")
		fmt.Print(thread)
		time.Sleep(time.Second * time.Duration(thread.opcode))
		fmt.Print("\n  Ended: ")
		fmt.Print(thread)
		outputChannel <- thread
	}
}

func formatOutput(inputData []channelTransaction, outputData []channelTransaction, outputDataUnsorted []channelTransaction) {

	fmt.Print("\n Input Opcodes: -")
	for index := 0; index < len(outputData); index++ {
		fmt.Print(inputData[index].opcode)
		fmt.Print("-")
	}
	fmt.Print("\n Process Order: -")
	for index := 0; index < len(outputData); index++ {
		fmt.Print(outputDataUnsorted[index].opcode)
		fmt.Print("-")
	}
	fmt.Print("\nOutput Opcodes: -")
	for index := 0; index < len(outputData); index++ {
		fmt.Print(outputData[index].opcode)
		fmt.Print("-")
	}

}
