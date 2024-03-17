package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const OutputDirectory string = "out"
const ImageWidth int = 100
const ImageHeight int = 100
const ColorChannelR uint8 = 0
const ColorChannelG uint8 = 100
const ColorChannelB uint8 = 200

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
	/*
		TODO: fix following error when converting inet_exp.png
		panic: runtime error: slice bounds out of range [:30000] with capacity 29997

		goroutine 1 [running]:
		main.encodeSingleImage(0x14000074e60, 0x752d, 0x2, 0x1)
			/Users/friedrich.burmeister/projects/file_to_video/encoder.go:158 +0x350
		main.encode(0x14000050020)
			/Users/friedrich.burmeister/projects/file_to_video/encoder.go:109 +0x110
		main.main()
			/Users/friedrich.burmeister/projects/file_to_video/encoder.go:50 +0x6c
		exit status 2
	*/

	// TODO: read file from console input
	const FileToRead string = "test2.txt"

	setup()

	file, err := os.Open(FileToRead)

	if err != nil {
		fmt.Println("error when opening file")
		panic(err)
	}

	defer file.Close()

	encode(file)
}

func createSampleImage() {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 0xff}

	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: ImageWidth, Y: ImageHeight}
	rectangle := image.Rectangle{Min: upLeft, Max: lowRight}
	imageData := image.NewRGBA(rectangle)

	for x := 0; x < ImageWidth; x++ {
		for y := 0; y < ImageHeight; y++ {
			colorToSet := black

			isUpLeft := x < 50 && y < 50
			isBottemRight := x > 50 && y > 50

			if isUpLeft || isBottemRight {
				colorToSet = white
			}

			imageData.Set(x, y, colorToSet)
		}
	}

	imageFile, _ := os.Create(OutputDirectory + "/" + "sample.png")
	defer imageFile.Close()

	png.Encode(imageFile, imageData)
}

func encode(file *os.File) {
	reader := bufio.NewReader(file)
	fileStats, _ := file.Stat()

	// calc bytes writable to image by multiplying image dimensions
	// and by subtracting the bytes (1 pixel = 3 bytes) (used to instruct the decoder)
	bytesPerImage := float64((ImageWidth * ImageHeight * 3) - 3)

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

	// todo use for case when its the last image to set all remaining pixels to 0,0,0
	bytesRead, err := reader.Read(bytes)

	if err != nil {
		fmt.Printf("error while trying to read %d bytes for image %d", bytesToRead, n)
	}

	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: ImageWidth, Y: ImageHeight}
	rectangle := image.Rectangle{Min: upLeft, Max: lowRight}
	imageData := image.NewRGBA(rectangle) // todo what's the diff between rgba and rgba64 ???

	isLastImage := n == totalImageCount
	decoderInstructionNotLastImage := color.RGBA{R: 255, G: 255, B: 255, A: 0xff}

	offset := 0

	for y := 0; y < ImageHeight; y++ {
		for x := 0; x < ImageWidth; x++ {
			if x == 0 && y == 0 {
				if isLastImage {
					imageData.Set(
						x,
						y,
						encodeLastBytePosition(bytesRead),
					)

					continue
				}

				imageData.Set(0, 0, decoderInstructionNotLastImage)
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
		channel = ColorChannelB
	case 1:
		// one byte "too much", one more pixel needed, last byte in the R channel
		channel = ColorChannelR
	case 2:
		// two bytes "too much", one more pixel needed, last byte in the G channel
		channel = ColorChannelG
	}

	// defines the coordinates of the pixel containing the last written byte
	// range 0..99 since current image dimensions are 100 * 100
	var y uint8 = 0
	x := bytesToEncode / 3

	if channel != ColorChannelR {
		x++
	}

	// the x and y indexes are calculated by the following:
	// y: how often ImageWidth can fit inside the amount of pixels that will contain file data
	// x: by subtracting the ImageWidth from x until x would be negative
	for {
		if (x - ImageWidth) < 0 {
			break
		}

		x -= ImageWidth
		y++
	}

	return color.RGBA{
		R: uint8(x),
		G: y,
		B: channel,
		A: 0xff,
	}
}

// todo remove
func keepStuffToNotLooseImports(file *os.File) {
	fileInfo, _ := file.Stat()
	fmt.Printf("file size: %d bytes\n", fileInfo.Size())

	// todo could iterate until read bytes == 0
	bytesToRead := 10
	byteData := make([]byte, bytesToRead)

	bytesActuallyRead, _ := file.Read(byteData)

	fmt.Printf("%d bytes read\n", bytesActuallyRead)

	fmt.Printf(
		"%d bytes: %s\n",
		bytesActuallyRead,
		string(byteData[:bytesActuallyRead]),
	)

	// todo how to create an image / a frame of a video?
	// - create image, store in out dir
	// - read images from out dir, create video

	// todo how will I tell the decoder where the files data ends?
	//  -> encode somewhere how many bytes are encoded in a single frame

	// bytes storable per image = height * width * color_channels = 200 * 200 * 3 = 120_000
	width := 200
	height := 200
	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: width, Y: height}
	rectangle := image.Rectangle{Min: upLeft, Max: lowRight}

	imageData := image.NewRGBA(rectangle)
	// now can set colors like: imageData.Set(123, 123, color.RGBA{123, 123, 123, 0xff})
	//color.RGBA{} // image/color
	//png.Encode() // image/png

	color1 := color.RGBA{R: 255, G: 127, B: 127, A: 0xff}
	color2 := color.RGBA{R: 255, G: 255, B: 255, A: 0xff}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			colorToSet := color1

			if x > 100 && y > 100 {
				colorToSet = color2
			}

			// if no color is set the pixel automatically becomes invisible

			imageData.Set(x, y, colorToSet)
		}
	}

	imageFile, _ := os.Create("image.png")
	defer imageFile.Close()

	png.Encode(imageFile, imageData)

}

func setup() {
	_, err := os.Stat(OutputDirectory)
	dirDoesNotExist := os.IsNotExist(err)

	if dirDoesNotExist {
		fmt.Println("output directory not found - creating...")

		// create directory with permissions rwx
		err = os.Mkdir(OutputDirectory, os.ModeDir|os.ModePerm)

		if err != nil {
			panic(err)
		}

		fmt.Println("output directory created")
	}
	// todo: else: delete it's content
}

/*fmt.Println("converting file to video...")

  const FileToRead string = "test1.txt"

  file, err := os.Open(FileToRead)

  if err != nil {
  	fmt.Println(err)
  }

  defer file.Close()

  reader := bufio.NewReader(file)
  byteRead, err := reader.ReadByte()

  if err != nil {
  	fmt.Println(err)
  }

  fmt.Println(byteRead)

  byteRead, err = reader.ReadByte()
  reader.Size()

  if err != nil {
  	fmt.Println(err)
  }

  fmt.Println(byteRead)*/
