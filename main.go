package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

type channelTransaction struct {
	opcode, order int
}

const numOpcodes = 10

func main() {
	dispatch()
}

func dispatch() {
	inputChannel := make(chan channelTransaction)
	outputChannel := make(chan channelTransaction)
	threadsAvailable := runtime.NumCPU() - 1
	fmt.Println("Threads Available: " + strconv.Itoa(threadsAvailable))
	for i := 0; i < threadsAvailable; i++ {
		go pipeline(inputChannel, outputChannel)
	}

	var inputArr [numOpcodes]int
	for index := 0; index < numOpcodes; index++ {
		inputArr[index] = rand.Intn(9)
	}
	var channelData [numOpcodes]channelTransaction
	var outputArr [numOpcodes]int
	for i := 0; i < numOpcodes; i++ {
		channelData[i] = channelTransaction{inputArr[i], i}
	}
	j := 0
	for i := 0; i < numOpcodes; {
		select {
		case inputChannel <- channelData[i]:
			i++
		case threadReturn := <-outputChannel:
			fmt.Println("End1: ID:" + strconv.Itoa(threadReturn.order) + " OpCode:" + strconv.Itoa(threadReturn.opcode))
			outputArr[threadReturn.order] = threadReturn.opcode
			j++
		default:
		}
	}
	for j < numOpcodes {
		select {
		case threadReturn := <-outputChannel:
			fmt.Println("End2: ID:" + strconv.Itoa(threadReturn.order) + " OpCode:" + strconv.Itoa(threadReturn.opcode))
			outputArr[threadReturn.order] = threadReturn.opcode
			j++
		default:
		}
	}
	fmt.Println(outputArr)
	fmt.Println(inputArr)
	fmt.Println(inputArr == outputArr)
}

func pipeline(inputChannel <-chan channelTransaction, outputChannel chan<- channelTransaction) {
	for {
		thread := <-inputChannel
		fmt.Println("Start: ID:" + strconv.Itoa(thread.order) + " OpCode:" + strconv.Itoa(thread.opcode))
		time.Sleep(time.Second * time.Duration(thread.opcode))
		outputChannel <- thread
	}
}
