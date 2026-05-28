package parser

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

// Parser analyzes RAMoops / kernel crash logs
type Parser struct {
	patterns []*ErrorPattern
}

// NewParser creates a parser with default patterns
func NewParser() *Parser {
	return &Parser{
		patterns: DefaultPatterns(),
	}
}

// ParseFile reads a log file from disk and analyzes it
func (p *Parser) ParseFile(path string) (*AnalysisResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase max line size for kernel logs
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return p.ParseLines(lines), nil
}

// ParseString parses raw log text
func (p *Parser) ParseString(logContent string) *AnalysisResult {
	lines := strings.Split(logContent, "\n")
	return p.ParseLines(lines)
}

// ParseLines analyzes an array of log lines
func (p *Parser) ParseLines(lines []string) *AnalysisResult {
	result := &AnalysisResult{
		TotalLines:     len(lines),
		Matches:        make([]Match, 0),
		SeverityCounts: make(map[Severity]int),
	}

	// Track matches per pattern for TopIssues aggregation
	type patternStats struct {
		count     int
		firstLine int
		pattern   *ErrorPattern
	}
	stats := make(map[string]*patternStats)

	for i, line := range lines {
		lineNum := i + 1
		for _, pat := range p.patterns {
			if pat.Pattern.MatchString(line) {
				m := Match{
					LineNumber: lineNum,
					LineText:   line,
					Pattern:    pat,
				}
				result.Matches = append(result.Matches, m)
				result.SeverityCounts[pat.Severity]++

				key := pat.Diagnosis
				if s, ok := stats[key]; ok {
					s.count++
				} else {
					stats[key] = &patternStats{
						count:     1,
						firstLine: lineNum,
						pattern:   pat,
					}
				}
			}
		}
	}

	// Build TopIssues sorted by severity (CRITICAL first), then count
	var issues []TopIssue
	for _, s := range stats {
		issues = append(issues, TopIssue{
			Severity:   s.pattern.Severity,
			Category:   s.pattern.Category,
			Diagnosis:  s.pattern.Diagnosis,
			ActionHint: s.pattern.ActionHint,
			Count:      s.count,
			FirstLine:  s.firstLine,
		})
	}

	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Severity != issues[j].Severity {
			return issues[i].Severity > issues[j].Severity // CRITICAL (2) > WARNING (1) > INFO (0)
		}
		return issues[i].Count > issues[j].Count
	})

	// Limit to top 5 issues
	if len(issues) > 5 {
		issues = issues[:5]
	}
	result.TopIssues = issues

	logger.Info("Log analysis complete: %d lines, %d matches, %d unique issues",
		result.TotalLines, len(result.Matches), len(result.TopIssues))

	return result
}

// HighlightLine returns a color category for a log line for UI rendering
// Returns: "critical", "warning", "info", "ksu", "susfs", or "" (no highlight)
func (p *Parser) HighlightLine(line string) string {
	for _, pat := range p.patterns {
		if pat.Pattern.MatchString(line) {
			switch pat.Severity {
			case CRITICAL:
				return "critical"
			case WARNING:
				return "warning"
			default:
				if pat.Category == "KernelSU" {
					return "ksu"
				}
				if pat.Category == "SUSFS" {
					return "susfs"
				}
				return "info"
			}
		}
	}
	return ""
}
