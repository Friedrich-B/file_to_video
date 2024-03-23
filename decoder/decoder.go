package decoder

import (
	"fileToVideo/setup"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strings"
)

const OutputDirectory string = "out-decoder"

func Decode() {
	setup.OutputDirectory(OutputDirectory)

	// TODO: read from console input
	const FileToOpen = "out-encoder/large.jpg.mp4"

	inputVideo, err := os.Open(FileToOpen)

	if err != nil {
		fmt.Printf("error when opening file %s\n", FileToOpen)
		panic(err)
	}

	defer inputVideo.Close()

	createImagesCommand := exec.Command(
		"ffmpeg",
		"-i",
		FileToOpen,
		"-vf",
		"fps=1",
		fmt.Sprintf("%s/%s", OutputDirectory, "%d.png"),
	)

	err = createImagesCommand.Run()

	if err != nil {
		fmt.Println("error when creating pngs from video")
		panic(err)
	}

	directoryContent, err := os.ReadDir(OutputDirectory)

	if err != nil {
		fmt.Println("error when opening output directory")
		panic(err)
	}

	imageCount := len(directoryContent)

	replacedString := strings.ReplaceAll(
		inputVideo.Name(),
		"/",
		"_",
	)

	outputFileName := fmt.Sprintf(
		"%s/%s",
		OutputDirectory,
		replacedString[0:len(replacedString)-4],
	)

	outputFile, err := os.Create(outputFileName)

	if err != nil {
		fmt.Println("error when creating output file")
	}

	defer outputFile.Close()

	for i := 1; i <= imageCount; i++ {
		if i == imageCount {
			decodeLastImage(i, outputFile)
		} else {
			decodeSingleImage(i, outputFile)
		}
	}

	fmt.Printf("finished decoding %s\n", inputVideo.Name())
}

func decodeSingleImage(n int, outputFile *os.File) {
	fileToDecode, err := os.Open(fmt.Sprintf("%s/%d.png", OutputDirectory, n))

	if err != nil {
		fmt.Printf("error opening file %d.png", n)
		panic(err)
	}

	defer fileToDecode.Close()

	decodedBytes := make([]byte, 0)

	decodedImage, _, err := image.Decode(fileToDecode)

	if err != nil {
		fmt.Printf("error when decoding image %d.png\n", n)
		panic(err)
	}

	for y := 0; y < setup.ImageHeight; y++ {
		for x := 0; x < setup.ImageWidth; x++ {
			if x == 0 && y == 0 {
				continue
			}

			pixel := decodedImage.At(x, y)
			r, g, b, _ := pixel.RGBA()

			decodedBytes = append(
				decodedBytes,
				byte(r),
				byte(g),
				byte(b),
			)
		}
	}

	_, err = outputFile.Write(decodedBytes)

	if err != nil {
		fmt.Println("error when writing to output file")
		panic(err)
	}
}

func decodeLastImage(n int, outputFile *os.File) {
	fileToDecode, err := os.Open(fmt.Sprintf("%s/%d.png", OutputDirectory, n))

	if err != nil {
		fmt.Printf("error opening file %d.png", n)
		panic(err)
	}

	defer fileToDecode.Close()

	decodedBytes := make([]byte, 0)

	decodedImage, _, err := image.Decode(fileToDecode)

	if err != nil {
		fmt.Printf("error when decoding image %d.png\n", n)
		panic(err)
	}

	lastByteX, lastByteY, lastByteChannel, _ := decodedImage.At(0, 0).RGBA()

readBytes:
	for y := 0; y < setup.ImageHeight; y++ {
		for x := 0; x < setup.ImageWidth; x++ {
			if x == 0 && y == 0 {
				continue
			}

			if x == int(lastByteX) && y == int(lastByteY) {
				pixel := decodedImage.At(x, y)
				r, g, b, _ := pixel.RGBA()

				switch uint8(lastByteChannel) {
				case setup.ColorChannelR:
					decodedBytes = append(decodedBytes, byte(r))
				case setup.ColorChannelG:
					decodedBytes = append(decodedBytes, byte(r), byte(g))
				case setup.ColorChannelB:
					decodedBytes = append(decodedBytes, byte(r), byte(g), byte(b))
				}

				break readBytes
			}

			pixel := decodedImage.At(x, y)
			r, g, b, _ := pixel.RGBA()

			decodedBytes = append(
				decodedBytes,
				byte(r),
				byte(g),
				byte(b),
			)
		}
	}

	_, err = outputFile.Write(decodedBytes)

	if err != nil {
		fmt.Println("error when writing to output file")
		panic(err)
	}
}
