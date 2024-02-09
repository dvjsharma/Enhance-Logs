package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

type LogEntry struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Keyword string `json:"keyword"`
	Message string `json:"message"`
}

type LogFilter struct {
	LogLevel        string
	CustomKeyword   string
	LogFilePath     string
	ServerAPI       bool
	APIResponsePath string
	JSONFilePath    string
}

func NewLogFilter() *LogFilter {
	return &LogFilter{
		LogFilePath:     "sample.log",
		APIResponsePath: "apiresponse.log",
		JSONFilePath:    "apiresponse.json",
	}
}

func (lf *LogFilter) ParseCommandLineFlags() {
	flag.StringVar(&lf.LogLevel, "level", "", "Log level to filter (INFO, ERROR, WARN, etc.)")
	flag.StringVar(&lf.CustomKeyword, "keyword", "", "Custom keyword to filter logs")
	flag.StringVar(&lf.LogFilePath, "file", "sample.log", "Path to the log file")
	flag.BoolVar(&lf.ServerAPI, "serverapi", false, "Save API response to a log file")
	flag.StringVar(&lf.JSONFilePath, "jsonfile", "apiresponse.json", "Path to the JSON log file")
	flag.Parse()
}

func (lf *LogFilter) PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("\nCommands:")
	fmt.Println("  -level=INFO     : Filter logs by log level (e.g., INFO, ERROR, WARN)")
	fmt.Println("  -keyword=string : Filter logs by a custom keyword")
	fmt.Println("  -file=path      : Specify the path to the log file")
	fmt.Println("  -serverapi      : Save API response to a log file")
	fmt.Println("  -jsonfile=path  : Specify the path to the JSON log file")
}

func (lf *LogFilter) Run() error {
	f, err := os.Open(lf.LogFilePath)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	if lf.ServerAPI {
		apiResponseFile, err := os.OpenFile(lf.APIResponsePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening API response log file: %v", err)
		}
		defer apiResponseFile.Close()

		apiResponseWriter := bufio.NewWriter(apiResponseFile)

		jsonFile, err := os.Create(lf.JSONFilePath)
		if err != nil {
			return fmt.Errorf("error creating JSON log file: %v", err)
		}
		defer jsonFile.Close()

		jsonWriter := bufio.NewWriter(jsonFile)
		jsonWriter.WriteString("[\n")

		firstEntry := true
		for {
			s, err := r.ReadString('\n')
			if err != nil {
				break
			}

			if (lf.LogLevel == "" || strings.Contains(s, lf.LogLevel)) &&
				(lf.CustomKeyword == "" || strings.Contains(s, lf.CustomKeyword)) {
				printColoredLog(s)

				apiResponseWriter.WriteString(s)

				logEntry := parseLogEntry(s)

				if !firstEntry {
					jsonWriter.WriteString(",\n")
				}
				firstEntry = false

				jsonEntry, _ := json.Marshal(logEntry)
				jsonWriter.WriteString("  " + string(jsonEntry))
			}
		}

		jsonWriter.WriteString("\n]\n")
		apiResponseWriter.Flush()
		jsonWriter.Flush()
	} else {
		for {
			s, err := r.ReadString('\n')
			if err != nil {
				break
			}

			if (lf.LogLevel == "" || strings.Contains(s, lf.LogLevel)) &&
				(lf.CustomKeyword == "" || strings.Contains(s, lf.CustomKeyword)) {
				printColoredLog(s)
			}
		}
	}

	return nil
}

func parseLogEntry(logEntry string) LogEntry {
	parts := strings.Fields(logEntry)
	dateTime := time.Now().Format("2006-01-02 15:04:05")

	return LogEntry{
		Date:    dateTime[:10],
		Time:    dateTime[11:19],
		Keyword: parts[2],
		Message: strings.Join(parts[3:], " "),
	}
}

func printColoredLog(logEntry string) {
	parts := strings.Fields(logEntry)

	switch strings.ToUpper(parts[2]) {
	case "WARNING":
		color.New(color.FgHiRed).Print(parts[0])
		fmt.Print(" ")
		color.New(color.FgHiBlack).Print(parts[1])
		fmt.Print(" ")
		color.New(color.FgHiRed).Print(parts[2])
		fmt.Print(" ")
		color.New(color.FgWhite).Print(strings.Join(parts[3:], " "))
	case "TRACE":
		color.New(color.FgHiRed).Print(parts[0])
		fmt.Print(" ")
		color.New(color.FgHiBlack).Print(parts[1])
		fmt.Print(" ")
		color.New(color.FgHiWhite).Print(parts[2])
		fmt.Print(" ")
		color.New(color.FgWhite).Print(strings.Join(parts[3:], " "))
	default:
		color.New(color.FgHiRed).Print(parts[0])
		fmt.Print(" ")
		color.New(color.FgHiBlack).Print(parts[1])
		fmt.Print(" ")
		color.New(color.FgHiGreen).Print(parts[2])
		fmt.Print(" ")
		color.New(color.FgWhite).Print(strings.Join(parts[3:], " "))
	}

	fmt.Println()
}
func main() {
	logFilter := NewLogFilter()
	logFilter.ParseCommandLineFlags()

	if flag.NFlag() == 0 || flag.Lookup("help") != nil {
		logFilter.PrintUsage()
		return
	}

	if err := logFilter.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
