package parser

import (
	"fmt"
	"log/slog"
	"os"
)

type Translations map[int]string

func ProcessD2iFile(d2iFilePath string) (Translations, error) {
	// See I18nFileAccessor.as
	translations := map[int]string{}
	slog.Debug("processing D2I file", "file", d2iFilePath)

	fileContentBytes, err := os.ReadFile(d2iFilePath)
	if err != nil {
		return translations, fmt.Errorf("error reading file: %w", err)
	}

	dataInput := NewDataInput(fileContentBytes)

	indexesPointer := dataInput.ReadInt()
	dataInput.SetPointer(indexesPointer)

	indexLen := dataInput.ReadInt()
	endIndexPointer := dataInput.IndexPointer + indexLen
	for dataInput.IndexPointer < endIndexPointer {
		id := dataInput.ReadInt()
		diacriticExists := dataInput.ReadBoolean()
		translations[id] = readString(dataInput, dataInput.ReadInt())
		if diacriticExists {
			// skip
			dataInput.ReadInt()
		}
	}

	return translations, nil
}

func readString(dataInput *DataInput, location int) string {
	startLocation := dataInput.IndexPointer
	dataInput.SetPointer(location)
	str := dataInput.ReadUTF()
	dataInput.SetPointer(startLocation)
	return str
}
