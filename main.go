package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("test_data/ubuntu.metadata")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(os.Stdout)

	decoder := Decoder{reader}
	encoder := Encoder{writer}
	
	
	output, err := decoder.Decode()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)

	err = encoder.Encode(output)

	if err != nil {
		log.Fatal(err)
	}

	encoder.buffer.Flush()
}