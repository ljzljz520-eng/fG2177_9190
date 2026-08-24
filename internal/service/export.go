package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"courseworkledger/internal/domain"
)

type ExportOptions struct {
	IncludeMissing bool
	OnlyGraded     bool
	Cohort         string
}

func (c *Catalog) ExportRows(options ExportOptions) ([]string, error) {
	dashboard, err := c.BuildDashboard()
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(dashboard.Rows)+1)
	lines = append(lines, "student_number,student_name,assignment,status,score,percent,message")
	for _, row := range dashboard.Rows {
		if !options.IncludeMissing && row.Status == domain.StatusNotSubmitted {
			continue
		}
		if options.OnlyGraded && row.Status != domain.StatusGraded && row.Status != domain.StatusReturned {
			continue
		}
		if options.Cohort != "" {
			student, findErr := c.FindStudentByNumber(row.StudentNumber)
			if findErr != nil || student.Cohort != options.Cohort {
				continue
			}
		}
		lines = append(lines, strings.Join([]string{row.StudentNumber, row.StudentName, row.Assignment, string(row.Status), row.Score, row.Percent, row.Message}, ","))
	}
	return lines, nil
}

func (c *Catalog) ExportJSON() ([]byte, error) {
	state, err := c.repository.ExportState()
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode export: %w", err)
	}
	return data, nil
}

func ParseExport(data []byte) (map[string]any, error) {
	var value map[string]any
	if len(data) == 0 {
		return nil, fmt.Errorf("export data is empty")
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("decode export: %w", err)
	}
	return value, nil
}
