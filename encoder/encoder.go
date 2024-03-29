package encoder

import (
	"bufio"
	"fileToVideo/setup"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
)

const OutputDirectory string = "out-encoder"

func Encode(fileToRead string) {
	setup.OutputDirectory(OutputDirectory)

	file, err := os.Open(fileToRead)

	if err != nil {
		fmt.Println("error when opening file")
		panic(err)
	}

	defer file.Close()

	encodeFileToPNGs(file)

	ffmpegCommand := exec.Command(
		"ffmpeg",
		"-framerate",
		"1",
		"-i",
		fmt.Sprintf("%s/%s", OutputDirectory, "%d.png"),
		"-c:v",
		"libx264rgb", // ensure rgb colors are preserved, don't convert to color space
		"-crf",       // ensure lossless compression
		"0",
		"-r",
		"1",
		fmt.Sprintf("%s/%s.mp4", OutputDirectory, fileToRead),
	)

	err = ffmpegCommand.Run()

	if err != nil {
		fmt.Println("error when converting images to video,", err)
		panic(err)
	}

	fmt.Printf("finished encoding %s\n", fileToRead)
}

func encodeFileToPNGs(file *os.File) {
	reader := bufio.NewReader(file)
	fileStats, _ := file.Stat()

	// calc bytes writable to image by multiplying image dimensions
	// and by subtracting the bytes (1 pixel = 3 bytes) (used to instruct the decoder)
	bytesPerImage := float64((setup.ImageWidth * setup.ImageHeight * 3) - 3)

	// this could cause issues with very big files which makes it more preferable
	// to create new images "on the fly" rather than calculating the image count beforehand IMO
	fileSize := float64(fileStats.Size())

	resultImagesCount := int(math.Ceil(fileSize / bytesPerImage))

	for i := 1; i <= resultImagesCount; i++ {
		encodeSingleImage(
			reader,
			int(bytesPerImage),
			resultImagesCount,
			i,
		)
	}
}

func encodeSingleImage(reader *bufio.Reader, bytesToRead int, totalImageCount int, n int) {
	bytes := make([]byte, bytesToRead)

	bytesRead, err := reader.Read(bytes)

	if err != nil {
		fmt.Printf("error while trying to read %d bytes for image %d", bytesToRead, n)
	}

	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: setup.ImageWidth, Y: setup.ImageHeight}
	rectangle := image.Rectangle{Min: upLeft, Max: lowRight}
	imageData := image.NewRGBA(rectangle)

	isLastImage := n == totalImageCount
	decoderInstructionNotLastImage := color.RGBA{R: 255, G: 255, B: 255, A: 0xff}

	offset := 0

	for y := 0; y < setup.ImageHeight; y++ {
		for x := 0; x < setup.ImageWidth; x++ {
			if x == 0 && y == 0 {
				// continue needed to not overwrite the first pixel and also to not increase the offset

				if isLastImage {
					imageData.Set(
						x,
						y,
						encodeLastBytePosition(bytesRead),
					)
				} else {
					imageData.Set(0, 0, decoderInstructionNotLastImage)
				}

				continue
			}

			encodeSinglePixel(
				imageData,
				x,
				y,
				bytes[offset:offset+3],
			)

			offset += 3
		}
	}

	fileName := fmt.Sprintf("%s/%d.png", OutputDirectory, n)
	imageFile, _ := os.Create(fileName)
	defer imageFile.Close()

	png.Encode(imageFile, imageData)
}

func encodeSinglePixel(imageData *image.RGBA, x int, y int, data []byte) {
	encodedBytes := color.RGBA{
		R: data[0],
		G: data[1],
		B: data[2],
		A: 0xff,
	}

	imageData.Set(x, y, encodedBytes)
}

/*
This function encodes the position of the last byte read from the input file.
The current image dimensions are set to 100 * 100.
The current implementation would limit it to have a maximum size of 256 * 256.
The posisition is encoded as:
R = xPositionIndex
G = yPositionIndex
B = colorChannelOfTheLastByte (0=R, 100=G, 200=B)
*/
func encodeLastBytePosition(bytesToEncode int) color.RGBA {
	// defines in which color channel the last byte is written
	// channels are written in the order: R -> G -> B
	// the value is one of [0, 100, 200] -> 0=R, 100=G, 200=B
	var channel uint8

	bytesLeft := bytesToEncode % 3

	switch bytesLeft {
	case 0:
		// all bytes can be stored within an amount of pixels that is dividable by 3
		channel = setup.ColorChannelB
	case 1:
		// one byte "too much", one more pixel needed, last byte in the R channel
		channel = setup.ColorChannelR
	case 2:
		// two bytes "too much", one more pixel needed, last byte in the G channel
		channel = setup.ColorChannelG
	}

	// defines the coordinates of the pixel containing the last written byte
	// range 0..99 since current image dimensions are 100 * 100
	var y uint8 = 0
	x := bytesToEncode / 3

	if channel != setup.ColorChannelR {
		x++
	}

	// the x and y indexes are calculated by the following:
	// y: how often ImageWidth can fit inside the amount of pixels that will contain file data
	// x: by subtracting the ImageWidth from x until x would be negative
	for {
		if (x - setup.ImageWidth) < 0 {
			break
		}

		x -= setup.ImageWidth
		y++
	}

	return color.RGBA{
		R: uint8(x),
		G: y,
		B: channel,
		A: 0xff,
	}
}
