package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type CovReportSlice []CovReport

func (c CovReportSlice) Len() int {
	return len(c)
}

func (c CovReportSlice) Less(i, j int) bool {
	cmp := strings.Compare(c[i].Filename, c[j].Filename)
	if cmp < 0 {
		return true
	}
	if cmp > 0 {
		return false
	}

	return c[i].StartLine < c[j].StartLine
}

func (c CovReportSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type CovReport struct {
	Filename  string
	StartLine int
	Raw       []byte
}

func main() {
	err := start()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func start() error {
	r := []CovReport{}
	for _, covfile := range os.Args[1:] {
		file, err := os.ReadFile(covfile)
		if err != nil {
			return fmt.Errorf("error reading file '%s': %w", covfile, err)
		}

		lines := bytes.Split(file, []byte("\n"))
		for _, line := range lines[1:] {
			if len(line) == 0 {
				continue
			}

			filename, rest, found := bytes.Cut(line, []byte(":"))
			if !found {
				return fmt.Errorf("found line with no ':' separator: '%s'", string(line))
			}
			startLineBytes, _, found := bytes.Cut(rest, []byte("."))
			if !found {
				return fmt.Errorf("found line with no '.' separator: '%s'", string(line))
			}

			startLine, err := strconv.Atoi(string(startLineBytes))
			if err != nil {
				return fmt.Errorf("found line with no '.' separator: '%s'", string(line))
			}

			r = append(r, CovReport{
				Filename:  string(filename),
				StartLine: startLine,
				Raw:       line,
			})
		}
	}

	sort.Sort(CovReportSlice(r))

	out, err := os.OpenFile("merged.coverage.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer out.Close()

	_, err = out.Write([]byte("mode: set\n"))
	if err != nil {
		return fmt.Errorf("error writing starting line to file: %w", err)
	}

	numLinesIgnored := 0
	lastReport := CovReport{}
	for _, report := range r {
		if report.Filename == lastReport.Filename &&
			report.StartLine == lastReport.StartLine {
			numLinesIgnored++
			continue
		}

		_, err := out.Write(append(report.Raw, '\n'))
		if err != nil {
			return fmt.Errorf("error writing line to file: %w", err)
		}

		lastReport = report
	}

	fmt.Println("number of duplicated lines ignored:", numLinesIgnored)

	return nil
}
