package main

import (
	"fileToVideo/decoder"
	"fileToVideo/encoder"
	"fmt"
	"os"
)

const Encode string = "encode"

const Decode string = "decode"

func main() {
	arguments := os.Args

	if len(arguments) < 3 {
		exit("not enough arguments")
	} else if len(arguments) > 3 {
		exit("too many arguments")
	}

	operation := arguments[1]
	isValidOperation := operation == Encode || operation == Decode

	if !isValidOperation {
		exit(fmt.Sprintf("got invalid operation '%s' as first argument", operation))
	}

	if operation == Encode {
		encoder.Encode(arguments[2])
	} else {
		decoder.Decode(arguments[2])
	}
}

func exit(reason string) {
	fmt.Println(reason)
	os.Exit(1)
}
