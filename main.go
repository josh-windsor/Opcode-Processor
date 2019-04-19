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

type channelTransaction struct {
	opcode, order int
}

func main() {
	quitChannel := make(chan bool)
	quitCompleteChannel := make(chan bool)
	go dispatch(quitChannel, quitCompleteChannel)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nJosh Windsor (24008182) Opcode Processor")
	fmt.Print("Press Q to quit")

	reader.ReadString('\n')
	fmt.Print("\nHalting Processor")
	quitChannel <- true
	<-quitCompleteChannel
	fmt.Println("\nProcessing Complete")
}

const (
	opnum = 15
)

func dispatch(quitChannel <-chan bool, quitCompleteChannel chan<- bool) {

	inputChannel := make(chan channelTransaction)
	outputChannel := make(chan channelTransaction)

	threadsAvailable := runtime.NumCPU() - 1
	fmt.Print("\nThreads Available: ")
	fmt.Print(threadsAvailable)
	for i := 0; i < threadsAvailable; i++ {
		go pipeline(inputChannel, outputChannel)
	}

	inputData := []channelTransaction{}
	outputData := []channelTransaction{}
	outputDataUnsorted := []channelTransaction{}
	opcodesSent := 0
	opcodesRetired := 0
	opcodesRecieved := 0
	for opcodesSent < opnum {
		randChannelData := channelTransaction{rand.Intn(9), opcodesSent}
		select {
		case <-quitChannel:
			break
		case inputChannel <- randChannelData:
			inputData = append(inputData, randChannelData)
			opcodesSent++
		case threadReturn := <-outputChannel:
			outputData = append(outputData, threadReturn)
			outputDataUnsorted = append(outputDataUnsorted, threadReturn)
			sort.Slice(outputData, func(i, j int) bool {
				return outputData[i].order < outputData[j].order
			})
			if opcodesRetired == outputData[opcodesRetired].order {
				fmt.Print("\nRetired: ")
				fmt.Print(outputData[opcodesRetired])
				opcodesRetired++
			}
			opcodesRecieved++
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
			if opcodesRetired == outputData[opcodesRetired].order {
				fmt.Print("\nRetired: ")
				fmt.Print(outputData[opcodesRetired])
				opcodesRetired++
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
		time.Sleep(time.Second * time.Duration(thread.opcode))
		outputChannel <- thread
	}
}

func retire(inputChannel <-chan channelTransaction, outputChannel chan<- channelTransaction) {

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
