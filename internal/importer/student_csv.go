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

type StudentImportIssue struct {
	Row     int
	Field   string
	Message string
}

type StudentImportResult struct {
	Students []domain.Student
	Issues   []StudentImportIssue
}

func ParseStudents(reader io.Reader) (StudentImportResult, error) {
	if reader == nil {
		return StudentImportResult{}, fmt.Errorf("student CSV reader is required")
	}
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1
	rows, err := csvReader.ReadAll()
	if err != nil {
		return StudentImportResult{}, fmt.Errorf("read student CSV: %w", err)
	}
	if len(rows) == 0 {
		return StudentImportResult{}, fmt.Errorf("student CSV is empty")
	}
	header, err := mapHeader(rows[0], []string{"id", "student_number", "name", "email", "cohort"})
	if err != nil {
		return StudentImportResult{}, fmt.Errorf("student CSV header: %w", err)
	}
	result := StudentImportResult{Students: make([]domain.Student, 0, len(rows)-1)}
	seenIDs := make(map[string]int)
	seenNumbers := make(map[string]int)
	for index, row := range rows[1:] {
		rowNumber := index + 2
		values, issue := valuesForHeader(row, header, rowNumber)
		if issue != nil {
			result.Issues = append(result.Issues, StudentImportIssue{Row: rowNumber, Field: issue.Field, Message: issue.Message})
			continue
		}
		student, createErr := domain.NewStudent(values["id"], values["student_number"], values["name"], values["email"], values["cohort"])
		if createErr != nil {
			result.Issues = append(result.Issues, StudentImportIssue{Row: rowNumber, Message: createErr.Error()})
			continue
		}
		if activeText, exists := values["active"]; exists && activeText != "" {
			active, parseErr := strconv.ParseBool(activeText)
			if parseErr != nil {
				result.Issues = append(result.Issues, StudentImportIssue{Row: rowNumber, Field: "active", Message: "must be true or false"})
				continue
			}
			student.Active = active
		}
		if firstRow, exists := seenIDs[student.ID]; exists {
			result.Issues = append(result.Issues, StudentImportIssue{Row: rowNumber, Field: "id", Message: fmt.Sprintf("duplicates row %d", firstRow)})
			continue
		}
		if firstRow, exists := seenNumbers[student.StudentNumber]; exists {
			result.Issues = append(result.Issues, StudentImportIssue{Row: rowNumber, Field: "student_number", Message: fmt.Sprintf("duplicates row %d", firstRow)})
			continue
		}
		seenIDs[student.ID] = rowNumber
		seenNumbers[student.StudentNumber] = rowNumber
		result.Students = append(result.Students, student)
	}
	result.Students = domain.SortStudents(result.Students)
	sort.SliceStable(result.Issues, func(i, j int) bool {
		if result.Issues[i].Row != result.Issues[j].Row {
			return result.Issues[i].Row < result.Issues[j].Row
		}
		return result.Issues[i].Field < result.Issues[j].Field
	})
	return result, nil
}

type rowIssue struct {
	Field   string
	Message string
}

func mapHeader(row []string, required []string) (map[string]int, error) {
	header := make(map[string]int, len(row))
	for index, value := range row {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			return nil, fmt.Errorf("column %d has an empty name", index+1)
		}
		if _, exists := header[name]; exists {
			return nil, fmt.Errorf("duplicate column %q", name)
		}
		header[name] = index
	}
	for _, name := range required {
		if _, exists := header[name]; !exists {
			return nil, fmt.Errorf("missing required column %q", name)
		}
	}
	return header, nil
}

func valuesForHeader(row []string, header map[string]int, rowNumber int) (map[string]string, *rowIssue) {
	values := make(map[string]string, len(header))
	for name, index := range header {
		if index >= len(row) {
			return nil, &rowIssue{Field: name, Message: fmt.Sprintf("row %d is missing column value", rowNumber)}
		}
		values[name] = strings.TrimSpace(row[index])
	}
	return values, nil
}
