package service

import (
	"fmt"
	"sort"
	"strings"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/report"
	"courseworkledger/internal/store"
)

type DashboardRow struct {
	StudentNumber string
	StudentName   string
	Assignment    string
	Status        domain.SubmissionStatus
	Score         string
	Percent       string
	Message       string
}

type Dashboard struct {
	Rows        []DashboardRow
	Total       int
	Submitted   int
	Missing     int
	Graded      int
	Average     string
	Warnings    []string
	GeneratedBy string
}

func (c *Catalog) BuildDashboard() (Dashboard, error) {
	students, err := c.repository.Students()
	if err != nil {
		return Dashboard{}, err
	}
	assignments, err := c.repository.Assignments()
	if err != nil {
		return Dashboard{}, err
	}
	list, err := c.SubmissionList()
	if err != nil {
		return Dashboard{}, err
	}
	dashboard := Dashboard{
		Rows:        make([]DashboardRow, 0, len(students)*len(assignments)),
		Total:       len(students) * len(assignments),
		Warnings:    make([]string, 0),
		GeneratedBy: "coursework-ledger",
	}
	for _, student := range students {
		if !student.Active {
			continue
		}
		for _, assignment := range assignments {
			record := list.FindByStudentAndAssignment(student.ID, assignment.ID)
			var row DashboardRow
			if record == nil {
				// A student with no submission for this assignment is rendered as a
				// not-submitted row instead of aborting the dashboard, so the
				// remaining grades and the class average still display.
				row = dashboardRowForMissing(student, assignment)
			} else {
				row = dashboardRowFromRecord(record)
			}
			dashboard.Rows = append(dashboard.Rows, row)
			dashboard.countRow(row)
		}
	}
	dashboard.finishAverage()
	sort.SliceStable(dashboard.Rows, func(i, j int) bool {
		if dashboard.Rows[i].StudentNumber != dashboard.Rows[j].StudentNumber {
			return dashboard.Rows[i].StudentNumber < dashboard.Rows[j].StudentNumber
		}
		return dashboard.Rows[i].Assignment < dashboard.Rows[j].Assignment
	})
	return dashboard, nil
}

func dashboardRowForMissing(student domain.Student, assignment domain.Assignment) DashboardRow {
	return DashboardRow{
		StudentNumber: student.StudentNumber,
		StudentName:   student.Name,
		Assignment:    assignment.Name,
		Status:        domain.StatusNotSubmitted,
		Score:         "-",
		Percent:       "-",
		Message:       "not submitted",
	}
}

func dashboardRowFromRecord(record *domain.SubmissionRecord) DashboardRow {
	row := DashboardRow{
		StudentNumber: record.Student.StudentNumber,
		StudentName:   record.Student.Name,
		Assignment:    record.Assignment.Name,
		Status:        record.Submission.Status,
		Score:         "-",
		Percent:       "-",
		Message:       "not submitted",
	}
	if record.Submission.Score != nil {
		row.Score = fmt.Sprintf("%d/%d", *record.Submission.Score, record.Assignment.MaximumScore)
		if percent, ok := submissionPercent(record); ok {
			row.Percent = fmt.Sprintf("%d%%", percent)
		}
		row.Message = "graded"
	} else if record.Submission.IsSubmitted() {
		row.Message = "submitted, awaiting grade"
	}
	return row
}

func submissionPercent(r *domain.SubmissionRecord) (int, bool) {
	if r == nil || r.Submission == nil || r.Assignment == nil || r.Submission.Score == nil {
		return 0, false
	}
	return (*r.Submission.Score * 100) / r.Assignment.MaximumScore, true
}

func (d *Dashboard) countRow(row DashboardRow) {
	if d == nil {
		return
	}
	if row.Status == domain.StatusNotSubmitted {
		d.Missing++
	} else {
		d.Submitted++
	}
	if row.Status == domain.StatusGraded || row.Status == domain.StatusReturned {
		d.Graded++
	}
	if row.Message == "not submitted" {
		d.Warnings = append(d.Warnings, row.StudentNumber+" has not submitted "+row.Assignment)
	}
}

func (d *Dashboard) finishAverage() {
	if d == nil {
		return
	}
	total := 0
	count := 0
	for _, row := range d.Rows {
		if row.Score == "-" {
			continue
		}
		parts := strings.SplitN(row.Score, "/", 2)
		var score, maximum int
		if _, err := fmt.Sscanf(parts[0], "%d", &score); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &maximum); err != nil || maximum == 0 {
			continue
		}
		total += score * 100 / maximum
		count++
	}
	if count > 0 {
		d.Average = fmt.Sprintf("%d%%", total/count)
	} else {
		d.Average = "-"
	}
}

func (d Dashboard) RenderText() string {
	lines := []string{fmt.Sprintf("Coursework dashboard: %d total, %d submitted, %d graded, average %s", d.Total, d.Submitted, d.Graded, d.Average)}
	for _, row := range d.Rows {
		lines = append(lines, fmt.Sprintf("%s %s | %s | %s | %s", row.StudentNumber, row.StudentName, row.Assignment, row.Status, row.Score))
	}
	for _, warning := range d.Warnings {
		lines = append(lines, "warning: "+warning)
	}
	return strings.Join(lines, "\n")
}

func (c *Catalog) BuildReportDocument() (report.SummaryDocument, error) {
	list, err := c.SubmissionList()
	if err != nil {
		return report.SummaryDocument{}, err
	}
	lines := make([]report.ResultLine, 0, list.Len())
	for _, record := range list.Records() {
		line, lineErr := report.NewResultLine(record)
		if lineErr != nil {
			return report.SummaryDocument{}, lineErr
		}
		lines = append(lines, line)
	}
	document := report.BuildDocument("Coursework submissions", lines, fmt.Sprintf("%d records", len(lines)))
	return document, nil
}

func (c *Catalog) SaveDashboardSnapshots(dashboard Dashboard) ([]domain.GradeSnapshot, error) {
	snapshots := make([]domain.GradeSnapshot, 0, len(dashboard.Rows))
	for index, row := range dashboard.Rows {
		student, found, err := c.repository.StudentByNumber(row.StudentNumber)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("student %s was not found", row.StudentNumber)
		}
		assignments, err := c.repository.Assignments()
		if err != nil {
			return nil, err
		}
		var assignment domain.Assignment
		for _, candidate := range assignments {
			if candidate.Name == row.Assignment {
				assignment = candidate
				break
			}
		}
		if assignment.ID == "" {
			return nil, fmt.Errorf("assignment %s was not found", row.Assignment)
		}
		record := c.RequireRecord(student.ID, assignment.ID)
		if record == nil {
			continue
		}
		snapshot, err := record.Snapshot(fmt.Sprintf("snapshot-%03d", index+1))
		if err != nil {
			return nil, err
		}
		if err := c.repository.SaveSnapshot(snapshot); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func (c *Catalog) RestoreState(state store.State) error {
	return c.repository.ReplaceState(state)
}
