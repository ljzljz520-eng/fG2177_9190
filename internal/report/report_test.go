package report

import (
	"testing"

	"courseworkledger/internal/domain"
)

func TestReportFormatting(t *testing.T) {
	student, _ := domain.NewStudent("s1", "1001", "Ada", "", "2026")
	assignment, _ := domain.NewAssignment("a1", "Pointers", "", 100, true, nil)
	submission, _ := domain.NewSubmission("sub", student.ID, assignment.ID, "answer", false)
	record, err := domain.NewSubmissionRecord(&student, &assignment, &submission)
	if err != nil {
		t.Fatal(err)
	}
	line, err := NewResultLine(record)
	if err != nil {
		t.Fatal(err)
	}
	document := BuildDocument("results", []ResultLine{line}, "1 record")
	if document.Render() == "" || StatusLabel(domain.StatusLate) != "late" {
		t.Fatal("report formatting failed")
	}
}
