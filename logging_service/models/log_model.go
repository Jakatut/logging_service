package models

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"logging_service/core"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Log Types
const (
	ALL         = 0
	DEBUG       = 1
	INFO        = 2
	WARN        = 3
	ERROR       = 4
	FATAL       = 5
	MinLogLevel = 1
	MaxLogLevel = 5
	MinSeverity = 1
	MaxSeverity = 7
)

var logLevels = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

//json:"created_date,omitempty" form:"created_date,omitempty"

// LogModel defines the contents of a log
type LogModel struct {
	CreatedDate time.Time `schema:"created_date"`
	Severity    int       `schema:"severity,omitempty"` // Severity levels are 1-7 (lowest to highest)
	Type        int       `schema:"type,omitempty"`     // DEBUG, INFO, WARN, ERROR, FATAL, ALL
	Message     string    `schema:"message,omitempty"`
	Location    string    `schema:"location,omitempty"` // Ideally filename or file location from the software using the logging service.
}

const logDateFormat = "2006-01-02T15-04-05Z07"
const resourceFileNameDateFormat = "2006-01-02"

// Writing

// WriteLog writes a log to a logfile.
func (logModel *LogModel) WriteLog(*core.FileMutexPool) error {
	logLocation, err := getLogWriteLocation(logModel)

	createLogLevelDirectory(logModel.Type)

	file, err := os.OpenFile(logLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return err
	}

	defer file.Close()

	bufferedWriter := bufio.NewWriter(file)
	bytesWritten, err := bufferedWriter.Write(
		buildLogMessage(logModel),
	)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("Bytes written: %d\n", bytesWritten)
	bufferedWriter.Flush()

	return nil
}

// Reading

// ReadLog reads a log from the log file.
func (logModel *LogModel) ReadLog(mutexPool *core.FileMutexPool) ([]LogModel, error) {

	var logs []LogModel

	if !logModel.CreatedDate.IsZero() {
		logLocations := logModel.getFileSearchLocations()
		for _, location := range logLocations {
			mutexPool.LockReadFileMutex(location)
			logs = append(logs, searchLog(location, logModel)...)
			mutexPool.UnlockReadFileMutex(location)
		}
	}

	return logs, nil
}

func (logModel *LogModel) getFileSearchLocations() []string {
	var logLevel, err = getLogLevelAsString(logModel.Type)
	if err != nil {
		return nil
	}

	var paths []string
	if logLevel != "ALL" {
		paths = walkLogLevelPaths([]string{logLevel})
	} else {
		paths = walkLogLevelPaths(logLevels)
	}

	return paths
}

func walkLogLevelPaths(logLevels []string) []string {
	var paths []string
	for _, level := range logLevels {
		err := filepath.Walk(""+level+"/", func(path string, info os.FileInfo, err error) error {
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			log.Fatal(err)
			return nil
		}
	}

	return paths
}

func searchLog(location string, logModel *LogModel) []LogModel {
	var foundLogs []LogModel
	file, _ := os.Open(location)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		logLine := rawLogToModel(scanner.Text())
		if logModel.Compare(logLine) {
			foundLogs = append(foundLogs, *logLine)
		}
	}

	return nil
}

func rawLogToModel(rawLog string) *LogModel {

	logTextIndicator := strings.Index(":", rawLog)
	// Remove leading and trailing braces, removes the content of the log, and splits the details.
	logProperties := strings.Split("]-[", rawLog[1:logTextIndicator-1])

	var logModel *LogModel
	logModel.CreatedDate, _ = time.Parse(logDateFormat, logProperties[0])
	logModel.Location, _ = url.QueryUnescape(logProperties[1])
	logModel.Severity, _ = strconv.Atoi(logProperties[2])
	logModel.Message, _ = url.QueryUnescape(rawLog[logTextIndicator:])

	return logModel
}

func createLogLevelDirectory(logLevel int) {
	path, err := getLogLevelAsString(logLevel)
	if err != nil {

	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}
}

func getLogLevelAsString(logLevel int) (string, error) {
	var logLevels = make(map[int]string)
	logLevels[0] = "ALL"
	logLevels[1] = "DEBUG"
	logLevels[2] = "INFO"
	logLevels[3] = "WARN"
	logLevels[4] = "ERROR"
	logLevels[5] = "FATAL"

	if logLevel <= 0 || logLevel >= 6 {
		return "", errors.New("logLevel must be 0 to 5")
	}

	return logLevels[logLevel], nil
}

func getLogWriteLocation(logModel *LogModel) (string, error) {
	logLevel, err := getLogLevelAsString(logModel.Type)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s-%d.txt", logLevel, time.Now().Format(resourceFileNameDateFormat), logModel.Severity), nil
}

func buildLogMessage(logModel *LogModel) []byte {
	location := url.QueryEscape(logModel.Location)
	messageText := url.QueryEscape(logModel.Message)
	return []byte(fmt.Sprintf("[%s]-[%s]-[%d]:%s\n", time.Now().Format(logDateFormat), location, logModel.Severity, messageText))
}

// Compare compares the values between two log models: The receiver and the comparison.
// If the two models are the same, true is returned. Otherwise, false.
func (logModel *LogModel) Compare(comparison *LogModel) bool {

	if !logModel.CreatedDate.IsZero() && logModel.CreatedDate != comparison.CreatedDate {
		return false
	}
	if (logModel.Severity >= MinSeverity && logModel.Severity <= MaxSeverity) && logModel.Severity != comparison.Severity {
		return false
	}
	if (logModel.Type >= MinLogLevel && logModel.Type <= MaxLogLevel) && logModel.Type != comparison.Type {
		return false
	}
	if logModel.Message != "" && logModel.Message != comparison.Message {
		return false
	}
	if logModel.Location != "" && logModel.Location != comparison.Location {
		return false
	}

	return true
}
