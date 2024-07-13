package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/brequet/dofus-data-file-parser/internal/parser"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("Usage:", os.Args[0], "[--debug] dofusDataFolderPath outputFolderPath")
		os.Exit(1)
	}

	dofusDataFolderPath := flag.Arg(0)
	outputFolderPath := flag.Arg(1)

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Dofus Data File Parser started")
	slog.Debug("debug mode enabled")

	err := checkDofusDataFolder(dofusDataFolderPath)
	if err != nil {
		slog.Error("error with provided dofus data folder", "error", err)
		os.Exit(1)
	}

	err = prepareOutputFolder(outputFolderPath)
	if err != nil {
		slog.Error("error preparing output folder", "error", err)
		os.Exit(1)
	}

	err = processCommonFolder(filepath.Join(dofusDataFolderPath, "common"), outputFolderPath)
	if err != nil {
		slog.Error("error processing common folder", "error", err)
	}

	err = processI18nFolder(filepath.Join(dofusDataFolderPath, "i18n"), outputFolderPath)
	if err != nil {
		slog.Error("error processing i18n folder", "error", err)
	}
}

func checkDofusDataFolder(dofusDataFolderPath string) error {
	err := checkFolderExists(dofusDataFolderPath)
	if err != nil {
		return err
	}

	err = checkFolderExists(filepath.Join(dofusDataFolderPath, "common"))
	if err != nil {
		return err
	}

	err = checkFolderExists(filepath.Join(dofusDataFolderPath, "i18n"))
	if err != nil {
		return err
	}

	return nil
}

func checkFolderExists(folderPath string) error {
	folderInfo, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %w", err)
	}

	if !folderInfo.IsDir() {
		return fmt.Errorf("folder is not a directory: %w", err)
	}

	return nil
}

func prepareOutputFolder(outputFolderPath string) error {
	err := os.RemoveAll(outputFolderPath)
	if err != nil {
		return fmt.Errorf("error removing output folder: %w", err)
	}

	err = os.MkdirAll(outputFolderPath, 0755)
	if err != nil {
		return fmt.Errorf("error creating output folder: %w", err)
	}

	err = os.Mkdir(filepath.Join(outputFolderPath, "common"), 0755)
	if err != nil {
		return fmt.Errorf("error creating common folder: %w", err)
	}

	err = os.Mkdir(filepath.Join(outputFolderPath, "translation"), 0755)
	if err != nil {
		return fmt.Errorf("error creating translation folder: %w", err)
	}

	return nil
}

func processCommonFolder(commonFolderPath, outputFolderPath string) error {
	files, err := os.ReadDir(commonFolderPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	fileParsedCount := 0
	for _, file := range files {
		if file.IsDir() {
			slog.Debug("skipping directory", "directory", file.Name())
			continue
		}

		if filepath.Ext(file.Name()) != ".d2o" {
			slog.Debug("skipping file (wrong extension)", "file", file.Name())
			continue
		}

		d2oFilePath := filepath.Join(commonFolderPath, file.Name())
		data, err := parser.ProcessD2oFile(d2oFilePath)
		if err != nil {
			slog.Error("error parsing file", "error", err)
			continue
		}

		slog.Debug("file parsed", "file", file.Name(), "classes", len(data.Classes), "objects", len(data.Objects))

		jsonStr, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			slog.Error("error marshalling json", "error", err)
		}

		outputPath := filepath.Join(outputFolderPath, "common", file.Name()+".json")
		err = os.WriteFile(outputPath, jsonStr, 0644)
		if err != nil {
			slog.Error("error writing file", "error", err, "path", outputPath)
		}
		fileParsedCount++
	}
	slog.Info("d2o files parsed", "count", fileParsedCount)

	return nil
}

func processI18nFolder(i18nFolderPath, outputFolderPath string) error {
	files, err := os.ReadDir(i18nFolderPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	fileParsedCount := 0
	for _, file := range files {
		if file.IsDir() {
			slog.Debug("skipping directory", "directory", file.Name())
			continue
		}

		if filepath.Ext(file.Name()) != ".d2i" {
			slog.Debug("skipping file (wrong extension)", "file", file.Name())
			continue
		}

		d2iFilePath := filepath.Join(i18nFolderPath, file.Name())
		translations, err := parser.ProcessD2iFile(d2iFilePath)
		if err != nil {
			return fmt.Errorf("error processing i18n file: %w", err)
		}

		jsonStr, err := json.MarshalIndent(translations, "", "  ")
		if err != nil {
			slog.Error("error marshalling json", "error", err)
		}

		outputPath := filepath.Join(outputFolderPath, "translation", getLocalFromD2iFileName(file.Name())+".json")
		err = os.WriteFile(outputPath, jsonStr, 0644)
		if err != nil {
			slog.Error("error writing file", "error", err, "path", outputPath)
		}
		fileParsedCount++
	}
	slog.Info("d2i files parsed", "count", fileParsedCount)

	return nil
}

func getLocalFromD2iFileName(d2iFileName string) string {
	return d2iFileName[len("i18n_") : len(d2iFileName)-len(".d2i")]
}
