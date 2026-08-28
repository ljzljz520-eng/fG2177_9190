package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"courseworkledger/internal/domain"
)

type AssignmentImportIssue struct {
	Row     int
	Field   string
	Message string
}

type AssignmentImportResult struct {
	Assignments []domain.Assignment
	Issues      []AssignmentImportIssue
}

func ParseAssignments(reader io.Reader) (AssignmentImportResult, error) {
	if reader == nil {
		return AssignmentImportResult{}, fmt.Errorf("assignment CSV reader is required")
	}
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1
	rows, err := csvReader.ReadAll()
	if err != nil {
		return AssignmentImportResult{}, fmt.Errorf("read assignment CSV: %w", err)
	}
	if len(rows) == 0 {
		return AssignmentImportResult{}, fmt.Errorf("assignment CSV is empty")
	}
	header, err := mapHeader(rows[0], []string{"id", "name", "maximum_score", "required"})
	if err != nil {
		return AssignmentImportResult{}, fmt.Errorf("assignment CSV header: %w", err)
	}
	result := AssignmentImportResult{Assignments: make([]domain.Assignment, 0, len(rows)-1)}
	seenIDs := make(map[string]int)
	for index, row := range rows[1:] {
		rowNumber := index + 2
		values, issue := valuesForHeader(row, header, rowNumber)
		if issue != nil {
			result.Issues = append(result.Issues, AssignmentImportIssue{Row: rowNumber, Field: issue.Field, Message: issue.Message})
			continue
		}
		maximumScore, parseErr := strconv.Atoi(values["maximum_score"])
		if parseErr != nil {
			result.Issues = append(result.Issues, AssignmentImportIssue{Row: rowNumber, Field: "maximum_score", Message: "must be an integer"})
			continue
		}
		required, parseErr := strconv.ParseBool(values["required"])
		if parseErr != nil {
			result.Issues = append(result.Issues, AssignmentImportIssue{Row: rowNumber, Field: "required", Message: "must be true or false"})
			continue
		}
		description := values["description"]
		tags := splitTags(values["tags"])
		assignment, createErr := domain.NewAssignment(values["id"], values["name"], description, maximumScore, required, tags)
		if createErr != nil {
			result.Issues = append(result.Issues, AssignmentImportIssue{Row: rowNumber, Message: createErr.Error()})
			continue
		}
		if firstRow, exists := seenIDs[assignment.ID]; exists {
			result.Issues = append(result.Issues, AssignmentImportIssue{Row: rowNumber, Field: "id", Message: fmt.Sprintf("duplicates row %d", firstRow)})
			continue
		}
		seenIDs[assignment.ID] = rowNumber
		result.Assignments = append(result.Assignments, assignment)
	}
	result.Assignments = domain.SortAssignments(result.Assignments)
	sort.SliceStable(result.Issues, func(i, j int) bool {
		if result.Issues[i].Row != result.Issues[j].Row {
			return result.Issues[i].Row < result.Issues[j].Row
		}
		return result.Issues[i].Field < result.Issues[j].Field
	})
	return result, nil
}

func splitTags(value string) []string {
	parts := strings.Split(value, "|")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		key := strings.ToLower(cleaned)
		if cleaned == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, cleaned)
	}
	sort.Strings(result)
	return result
}

func FormatStudentIssues(issues []StudentImportIssue) string {
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		location := fmt.Sprintf("row %d", issue.Row)
		if issue.Field != "" {
			location += ", " + issue.Field
		}
		parts = append(parts, location+": "+issue.Message)
	}
	return strings.Join(parts, "; ")
}

func FormatAssignmentIssues(issues []AssignmentImportIssue) string {
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		location := fmt.Sprintf("row %d", issue.Row)
		if issue.Field != "" {
			location += ", " + issue.Field
		}
		parts = append(parts, location+": "+issue.Message)
	}
	return strings.Join(parts, "; ")
}
