package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"sort"
)

type D2oData struct {
	Classes map[int]Class `json:"classes"`
	Objects []Object      `json:"objects"`
}

type Class struct {
	PackageName  string          `json:"packageName"`
	PackageClass string          `json:"packageClass"`
	Fields       []GameDataField `json:"fields"`
}

type Object = any

type GameDataField struct {
	Name    string         `json:"name"`
	Type    FieldType      `json:"type"`
	SubType *GameDataField `json:"subtype,omitempty"`
}

type FieldType int

// GameDataTypeEnum constants
const (
	Integer         FieldType = -1
	Boolean         FieldType = -2
	String          FieldType = -3
	Number          FieldType = -4
	I18n            FieldType = -5
	UnsignedInteger FieldType = -6
	Vector          FieldType = -99
)

func (f FieldType) String() string {
	switch f {
	case Integer:
		return "Integer"
	case Boolean:
		return "Boolean"
	case String:
		return "String"
	case Number:
		return "Number"
	case I18n:
		return "I18n"
	case UnsignedInteger:
		return "UnsignedInteger"
	case Vector:
		return "Vector"
	default:
		return fmt.Sprintf("%d", f)
	}
}

func (f FieldType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func ProcessD2oFile(d2oFilePath string) (D2oData, error) {
	// See GameDataFileAccessor.as
	slog.Debug("processing D2O file", "file", d2oFilePath)

	fileContentBytes, err := os.ReadFile(d2oFilePath)
	if err != nil {
		return D2oData{}, fmt.Errorf("error reading file: %w", err)
	}

	dataInput := NewDataInput(fileContentBytes)
	header := string(dataInput.Read(3))
	if header != "D2O" {
		return D2oData{}, fmt.Errorf("invalid header: %s", header)
	}

	indexesPointer := dataInput.ReadInt()
	dataInput.SetPointer(indexesPointer)
	slog.Debug("indexes pointer", "pointer", indexesPointer)

	indexTable := make(map[int]int)
	indexesLength := dataInput.ReadInt() / 8
	slog.Debug("indexes length", "length", indexesLength)
	for i := 0; i < indexesLength; i++ {
		key := dataInput.ReadInt()
		pointer := dataInput.ReadInt()
		indexTable[key] = pointer
	}

	classTable := make(map[int]Class)
	classCount := dataInput.ReadInt()
	slog.Debug("class count", "count", classCount)
	for i := 0; i < classCount; i++ {
		classIdentifier := dataInput.ReadInt()
		class := readClassDefinition(dataInput)
		classTable[classIdentifier] = class
	}

	objects := make([]Object, 0)
	indexValues := getSortedValues(indexTable)
	slog.Debug("index values", "count", len(indexValues))
	for _, index := range indexValues {
		dataInput.SetPointer(index)
		slog.Debug("reading object", "index", dataInput.OffsetStr())
		classId := dataInput.ReadInt()
		object := readObject(dataInput, classTable, classTable[classId])
		objects = append(objects, object)
	}

	return D2oData{
		Classes: classTable,
		Objects: objects,
	}, nil
}

func readClassDefinition(dataInput *DataInput) Class {
	className := dataInput.ReadUTF()
	packageName := dataInput.ReadUTF()

	slog.Debug("reading class", "package", packageName, "class", className)

	fields := make([]GameDataField, 0)
	fieldsCount := dataInput.ReadInt()
	for i := 0; i < fieldsCount; i++ {
		fields = append(fields, readField(dataInput))
	}

	return Class{
		PackageName:  packageName,
		PackageClass: className,
		Fields:       fields,
	}
}

func readField(dataInput *DataInput) GameDataField {
	fieldName := dataInput.ReadUTF()
	var fieldType FieldType
	var subType *GameDataField

	fieldTypeId := dataInput.ReadInt()
	switch FieldType(fieldTypeId) {
	case Vector:
		fieldType = Vector
		subTypeObj := readField(dataInput)
		subType = &subTypeObj
	default:
		if fieldTypeId < 0 { // GameDataTypeEnum cases
			fieldType = FieldType(fieldTypeId)
		} else if fieldTypeId > 0 { // Custom Object cases
			fieldType = FieldType(fieldTypeId)
		} else {
			log.Fatal("unknown type ", dataInput.IndexPointer, fieldTypeId)
		}
	}

	return GameDataField{
		Name:    fieldName,
		Type:    fieldType,
		SubType: subType,
	}
}

func readObject(dataInput *DataInput, classeTable map[int]Class, class Class) Object {
	object := map[string]any{}
	object["ClassType_"] = class.PackageClass

	slog.Debug("reading object", "class", fmt.Sprintf("%s.%s", class.PackageName, class.PackageClass), "field count", len(class.Fields), "offset", dataInput.OffsetStr())
	for _, field := range class.Fields {
		fieldObject := interface{}(nil)
		fieldType := field.Type
		slog.Debug("reading field", "name", field.Name, "type", fieldType, "offset", dataInput.OffsetStr())
		switch fieldType {
		case Integer:
			fieldObject = dataInput.ReadInt()
		case Boolean:
			fieldObject = dataInput.ReadBoolean()
		case String:
			fieldObject = dataInput.ReadUTF()
		case Number:
			number := dataInput.ReadDouble()
			if math.IsNaN(number) {
				fieldObject = nil
			} else {
				fieldObject = number
			}
		case I18n:
			fieldObject = dataInput.ReadInt()
		case UnsignedInteger:
			fieldObject = dataInput.ReadUint()
		case Vector:
			fieldObject = readVector(dataInput, classeTable, field)
		default:
			classId := dataInput.ReadInt()
			if _, ok := classeTable[classId]; !ok {
				classId = int(field.Type)
			}
			fieldObject = readObject(dataInput, classeTable, classeTable[classId])
		}
		object[field.Name] = fieldObject
	}

	return object
}

func readVector(dataInput *DataInput, classeTable map[int]Class, field GameDataField) Object {
	vector := []any{}

	vectorLength := dataInput.ReadInt()
	slog.Debug("reading vector", "size", vectorLength, slog.Group("field", "name", field.Name, "type", field.Type), "offset", dataInput.OffsetStr())
	for i := 0; i < vectorLength; i++ {
		// slog.Debug("reading vector element", "index", i, "type", field.SubType.Type, "offset", dataInput.OffsetStr())
		switch field.SubType.Type {
		case Integer:
			vector = append(vector, dataInput.ReadInt())
		case Boolean:
			vector = append(vector, dataInput.ReadBoolean())
		case String:
			vector = append(vector, dataInput.ReadUTF())
		case Number:
			vector = append(vector, dataInput.ReadDouble())
		case I18n:
			vector = append(vector, dataInput.ReadInt())
		case UnsignedInteger:
			vector = append(vector, dataInput.ReadUint())
		case Vector:
			vector = append(vector, readVector(dataInput, classeTable, *field.SubType))
		default:
			classId := dataInput.ReadInt()
			if _, ok := classeTable[classId]; ok {
				vector = append(vector, readObject(dataInput, classeTable, classeTable[classId]))
			} else {
				vector = append(vector, nil)
			}
		}
	}

	return vector
}

func getSortedValues(m map[int]int) []int {
	values := make([]int, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}
