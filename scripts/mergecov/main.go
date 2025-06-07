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

type CovSection struct {
	Filename  string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmts  int
	IsCovered int
	CovFile   string
}

func NewCovSection(covfile string, line []byte) (*CovSection, error) {
	filename, rest, found := bytes.Cut(line, []byte(":"))
	if !found {
		return nil, fmt.Errorf("found line with no ':' separator: '%s'", string(line))
	}

	startLine, rest, err := parseInt(rest, ".")
	if err != nil {
		return nil, fmt.Errorf("error parsing startLine: %w", err)
	}

	startCol, rest, err := parseInt(rest, ",")
	if err != nil {
		return nil, fmt.Errorf("error parsing startCol: %w", err)
	}

	endLine, rest, err := parseInt(rest, ".")
	if err != nil {
		return nil, fmt.Errorf("error parsing endLine: %w", err)
	}

	endCol, rest, err := parseInt(rest, " ")
	if err != nil {
		return nil, fmt.Errorf("error parsing endCol: %w", err)
	}

	numStmts, rest, err := parseInt(rest, " ")
	if err != nil {
		return nil, fmt.Errorf("error parsing numStmts: %w", err)
	}

	isCovered, err := strconv.Atoi(string(rest))
	if err != nil {
		return nil, fmt.Errorf("error parsing isCovered: %w", err)
	}

	return &CovSection{
		Filename:  string(filename),
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmts:  numStmts,
		IsCovered: isCovered,
		CovFile:   covfile,
	}, nil
}

func parseInt(line []byte, delim string) (i int, rest []byte, err error) {
	b, rest, found := bytes.Cut(line, []byte(delim))
	if !found {
		return 0, nil, fmt.Errorf("missing '%s' separator on line: '%s'", delim, string(line))
	}

	i, err = strconv.Atoi(string(b))
	if err != nil {
		return 0, nil, fmt.Errorf("expected to parse integer value for line '%s': %w", string(line), err)
	}

	return i, rest, nil
}

func (c CovSection) Bytes() []byte {
	return []byte(
		fmt.Sprintf("%s:%d.%d,%d.%d %d %d",
			c.Filename,
			c.StartLine,
			c.StartCol,
			c.EndLine,
			c.EndCol,
			c.NumStmts,
			c.IsCovered,
		),
	)
}

type CovSectionSlice []*CovSection

func (c CovSectionSlice) Len() int {
	return len(c)
}

func (c CovSectionSlice) Less(i, j int) bool {
	cmp := strings.Compare(c[i].Filename, c[j].Filename)
	if cmp < 0 {
		return true
	}
	if cmp > 0 {
		return false
	}

	return c[i].StartLine < c[j].StartLine
}

func (c CovSectionSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func main() {
	err := start()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func start() error {
	r := []*CovSection{}
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

			covSection, err := NewCovSection(covfile, line)
			if err != nil {
				return err
			}

			r = append(r, covSection)
		}
	}

	sort.Sort(CovSectionSlice(r))

	// Remove duplicate lines:
	numLinesMerged, r := removeDuplicateLines(r)

	fmt.Println("number of duplicated lines merged:", numLinesMerged)

	// Manually add coverage for gaps in order to fix problem on CodeCov
	numGapsCovered, r := fillGapsInCoverage(r)

	fmt.Println("number of gaps covered:", numGapsCovered)

	// Write coverage report to file:

	out, err := os.OpenFile("merged.coverage.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer out.Close()

	_, err = out.Write([]byte("mode: set\n"))
	if err != nil {
		return fmt.Errorf("error writing starting line to file: %w", err)
	}

	for _, report := range r {
		_, err := out.Write(append(report.Bytes(), '\n'))
		if err != nil {
			return fmt.Errorf("error writing line to file: %w", err)
		}
	}

	return nil
}

func removeDuplicateLines(s []*CovSection) (numLinesMerged int, _ []*CovSection) {
	numLinesMerged = 0
	lastReport := s[0]
	out := s[:0] // Reuse the same slice for memory efficiency
	for _, report := range s {
		if report.Filename == lastReport.Filename &&
			report.StartLine == lastReport.StartLine {

			// When finding duplicates keep the ones that say it was tested:
			if report.IsCovered == 1 {
				lastReport = report
			}
			numLinesMerged++
			continue
		}

		out = append(out, lastReport)

		lastReport = report
	}

	out = append(out, lastReport)

	return numLinesMerged, out
}

// This function is only required to fix an error on codecov, normally lines
// with no testable code like comments, struct declarations etc are omitted
// from code coverage profiles, but codecov is randomly interpreting some of
// them as if they were untested code, we'll fix it by making them tested code
// with number of statements = 0, so they don't affect the overall average.
func fillGapsInCoverage(ss []*CovSection) (numGapsCovered int, out []*CovSection) {
	prevSec := &CovSection{}
	for _, s := range ss {
		if prevSec.Filename != s.Filename {
			prevSec = &CovSection{Filename: s.Filename, EndLine: 0, EndCol: 0}
		}

		if prevSec.EndLine < s.StartLine || prevSec.EndCol < s.EndCol {
			out = append(out, &CovSection{
				Filename:  s.Filename,
				StartLine: prevSec.EndLine,
				StartCol:  prevSec.EndCol,
				EndLine:   s.StartLine,
				EndCol:    s.StartCol,
				NumStmts:  0,
				IsCovered: 1,
			})

			numGapsCovered++
		}

		out = append(out, s)
	}

	return numGapsCovered, out
}
