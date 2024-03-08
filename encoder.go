package main

import (
	"fmt"
	"os"
)

/*
TODO:
- read in a file (first hard coded, later by passing path as argument in the cli)
- use the binary stream (or hex if dorectly available idk) and use it to set the pixels of an image
- create multiple images like this
- make video from these images
- video name should be the name of the file read in at the beginnings

- make a decoder that reads a video
- gets each frame (frame rate will be defined in my encoder I guess)
- reads each "pixel's" rgb hex values (3 Bytes, one per color channel)
(a "pixel" might become bigger than an actual pixel e.g. 2*2 pixels, evaluate later how much youtube video compression affects it)
- writes the read out values to a new file, in the end I should have a copy of the file I encoded previously
*/

func main() {
	fmt.Println("converting file to video...")

	const FILE_TO_READ string = "test.txt"

	// reads out the content of the file as a char array, can be used if you just want to get the file content itself
	//data, err := os.ReadFile(FILE_TO_READ)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println(string(data))

	file, err := os.Open(FILE_TO_READ)

	// close filestream if function ends or error occurs
	defer file.Close()

	check(err)

	fileInfo, err := file.Stat()

	check(err)
	fmt.Printf("file size: %d bytes\n", fileInfo.Size())

	// todo could iterate until read bytes == 0
	bytesToRead := 10
	byteData := make([]byte, bytesToRead)

	bytesActuallyRead, err := file.Read(byteData)

	check(err)
	fmt.Printf("%d bytes read\n", bytesActuallyRead)

	fmt.Printf(
		"%d bytes: %s\n",
		bytesActuallyRead,
		string(byteData[:bytesActuallyRead]),
	)

	// todo how to create an image / a frame of a video?
	// im curious what the images will look like

	// todo how will I tell the decoder where the files data ends?
	//  -> encode somewhere how many bytes are encoded in a single frame
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
