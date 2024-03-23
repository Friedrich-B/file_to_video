package setup

import (
	"os"
)

const ImageWidth int = 100
const ImageHeight int = 100
const ColorChannelR uint8 = 0
const ColorChannelG uint8 = 100
const ColorChannelB uint8 = 200

func OutputDirectory(directory string) {
	_, err := os.Stat(directory)
	dirDoesNotExist := os.IsNotExist(err)

	if dirDoesNotExist {
		createOutputDirectory(directory)
	} else {
		err = os.RemoveAll(directory)

		if err != nil {
			panic(err)
		}

		createOutputDirectory(directory)
	}
}

func createOutputDirectory(directory string) {
	err := os.Mkdir(directory, os.ModeDir|os.ModePerm)

	if err != nil {
		panic(err)
	}
}
