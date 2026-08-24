package report

import (
	"fmt"
	"sort"
	"strings"

	"courseworkledger/internal/domain"
)

type ResultLine struct {
	Student    string
	Assignment string
	Status     domain.SubmissionStatus
	Score      string
	Message    string
}

type SummaryDocument struct {
	Title   string
	Headers []string
	Lines   []ResultLine
	Footer  string
}

func NewResultLine(record *domain.SubmissionRecord) (ResultLine, error) {
	if record == nil || record.Student == nil || record.Assignment == nil || record.Submission == nil {
		return ResultLine{}, fmt.Errorf("complete submission record is required")
	}
	line := ResultLine{
		Student:    record.Student.DisplayName(),
		Assignment: record.Assignment.Name,
		Status:     record.Submission.Status,
		Score:      "-",
		Message:    "not submitted",
	}
	if record.Submission.Score != nil {
		line.Score = fmt.Sprintf("%d/%d", *record.Submission.Score, record.Assignment.MaximumScore)
		line.Message = "graded"
	} else if record.Submission.IsSubmitted() {
		line.Message = "submitted, awaiting grade"
	}
	return line, nil
}

func SortLines(lines []ResultLine) []ResultLine {
	result := append([]ResultLine(nil), lines...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Student != result[j].Student {
			return result[i].Student < result[j].Student
		}
		return result[i].Assignment < result[j].Assignment
	})
	return result
}

func BuildDocument(title string, lines []ResultLine, footer string) SummaryDocument {
	return SummaryDocument{
		Title:   strings.TrimSpace(title),
		Headers: []string{"student", "assignment", "status", "score", "message"},
		Lines:   SortLines(lines),
		Footer:  strings.TrimSpace(footer),
	}
}

func (d SummaryDocument) Render() string {
	parts := make([]string, 0, len(d.Lines)+2)
	if d.Title != "" {
		parts = append(parts, d.Title)
	}
	for _, line := range d.Lines {
		parts = append(parts, strings.Join([]string{line.Student, line.Assignment, string(line.Status), line.Score, line.Message}, " | "))
	}
	if d.Footer != "" {
		parts = append(parts, d.Footer)
	}
	return strings.Join(parts, "\n")
}

func StatusLabel(status domain.SubmissionStatus) string {
	switch status {
	case domain.StatusNotSubmitted:
		return "not submitted"
	case domain.StatusSubmitted:
		return "submitted"
	case domain.StatusLate:
		return "late"
	case domain.StatusGraded:
		return "graded"
	case domain.StatusReturned:
		return "returned"
	default:
		return "unknown"
	}
}
