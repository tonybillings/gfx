package gfx

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultSignalDataFilename = "signal_data_%d.csv"
)

func getSignalDataCsvFilename(filenameTemplate ...string) string {
	template := defaultSignalDataFilename
	if len(filenameTemplate) > 0 {
		template = filenameTemplate[0]
	}
	return fmt.Sprintf(template, time.Now().UnixMilli())
}

func ExportSignalDataToCsv(line *SignalLine, filenameTemplate ...string) error {
	line.Signal.Lock()
	data := make([]float64, line.Signal.dataSize)
	copy(data, line.Signal.dataTransformed)
	line.Signal.Unlock()

	filename := getSignalDataCsvFilename(filenameTemplate...)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating signal data file: %w", err)
	}
	defer func(file *os.File) {
		if e := file.Close(); e != nil {
			panic(fmt.Errorf("error closing signal data file: %w", e))
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err = writer.Write([]string{line.Name()}); err != nil {
		return fmt.Errorf("error writing to signal data file: %w", err)
	}

	for _, d := range data {
		dStr := strconv.FormatFloat(d, 'f', -1, 64)
		if err = writer.Write([]string{dStr}); err != nil {
			return fmt.Errorf("error writing to signal data file: %w", err)
		}
	}

	return nil
}

func ExportSignalGroupDataToCsv(group *SignalGroup, filenameTemplate ...string) error {
	signals := group.Signals()
	if len(signals) == 0 {
		return fmt.Errorf("no signals in the group")
	}

	filename := getSignalDataCsvFilename(filenameTemplate...)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating signal group data file: %w", err)
	}
	defer func(file *os.File) {
		if e := file.Close(); e != nil {
			panic(fmt.Errorf("error closing signal group data file: %w", e))
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := make([]string, len(signals))
	for i, line := range signals {
		header[i] = line.Name()
	}
	if err = writer.Write(header); err != nil {
		return fmt.Errorf("error writing header to signal group data file: %w", err)
	}

	// Warning: assumes all signals in the group have the same data size!
	dataSize := signals[0].Signal.dataSize

	for i := 0; i < dataSize; i++ {
		row := make([]string, len(signals))
		for j, line := range signals {
			line.Signal.Lock()
			if i < line.dataSize {
				row[j] = strconv.FormatFloat(line.Signal.dataTransformed[i], 'f', -1, 64)
			} else {
				row[j] = ""
			}
			line.Signal.Unlock()
		}
		if err = writer.Write(row); err != nil {
			return fmt.Errorf("error writing data to signal group data file: %w", err)
		}
	}

	return nil
}
