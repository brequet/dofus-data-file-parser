package generator

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/brequet/dofus-data-file-parser/pkg/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateGoFromClasses(classes []parser.Class) ([]byte, error) {
	fileContent, err := buildFileContent(classes)
	if err != nil {
		return nil, fmt.Errorf("build file content: %w", err)
	}

	protocolGoFileContent, err := formatGolangFile([]byte(fileContent))
	if err != nil {
		return nil, fmt.Errorf("format file to golang: %w", err)
	}

	return protocolGoFileContent, nil

}

func formatGolangFile(fileContent []byte) ([]byte, error) {
	formattedSrc, err := format.Source(fileContent)
	if err != nil {
		return nil, fmt.Errorf("format file: %w", err)
	}

	return formattedSrc, nil
}

func buildFileContent(classList []parser.Class) ([]byte, error) {
	var fileContent bytes.Buffer

	fileContent.WriteString("package types\n\n")

	for _, class := range classList {
		fileContent.WriteString(buildClassStruct(class))
	}

	return fileContent.Bytes(), nil
}

func buildClassStruct(class parser.Class) string {
	var fileContent bytes.Buffer

	fileContent.WriteString(fmt.Sprintf("type %s struct {\n", class.PackageClass))
	for _, field := range class.Fields {
		fileContent.WriteString(buildField(field))
	}
	fileContent.WriteString("}\n\n")

	return fileContent.String()
}

func buildField(field parser.GameDataField) string {
	var fileContent bytes.Buffer

	if field.Type == parser.Vector {
		fileContent.WriteString(handleVectorFieldType(field))
	} else if field.Type < 0 {
		fileContent.WriteString(fmt.Sprintf("%s %s `json:\"%s\"`\n", toTitledString(field.Name), mapSimpleFieldTypeToGolangType(field.Type), field.Name))
	} else {
		// custom type
		fileContent.WriteString(fmt.Sprintf("// %s custom type not implemented (%s)\n", field.Name, field.Type))
		// TODO
	}

	return fileContent.String()
}

func handleVectorFieldType(field parser.GameDataField) string {
	var fileContent bytes.Buffer

	if field.SubType.Type == parser.Vector {
		// TODO
		fileContent.WriteString(fmt.Sprintf("// %s vector subtype not implemented\n", field.Name))
	} else if field.SubType.Type < 0 {
		fileContent.WriteString(fmt.Sprintf("%s []%s `json:\"%s\"`\n", toTitledString(field.Name), mapSimpleFieldTypeToGolangType(field.SubType.Type), field.Name))
	} else {
		// TODO
		fileContent.WriteString(fmt.Sprintf("// %s vector custom subtype not implemented (%s)\n", field.Name, field.SubType.Type))
	}

	return fileContent.String()
}

func mapSimpleFieldTypeToGolangType(fieldType parser.FieldType) string {
	switch fieldType {
	case parser.Integer:
		return "int"
	case parser.Boolean:
		return "bool"
	case parser.String:
		return "string"
	case parser.Number:
		return "float64"
	case parser.I18n:
		return "int"
	case parser.UnsignedInteger:
		return "uint"
	default:
		panic("unknown field type: " + fieldType.String())
	}
}

func toTitledString(str string) string {
	return cases.Title(language.Und, cases.NoLower).String(str)
}
