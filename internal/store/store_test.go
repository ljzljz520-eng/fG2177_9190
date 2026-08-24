package store

import (
	"testing"

	"courseworkledger/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := t.TempDir() + "/coursework.db"
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	student, _ := domain.NewStudent("s1", "1001", "Ada", "", "2026")
	assignment, _ := domain.NewAssignment("a1", "Pointers", "", 100, true, nil)
	submission, _ := domain.NewSubmission("sub", student.ID, assignment.ID, "answer", false)
	score := 91
	submission.Score = &score
	submission.Status = domain.StatusGraded
	snapshot := domain.GradeSnapshot{ID: "snap", StudentID: student.ID, AssignmentID: assignment.ID, StudentNumber: student.StudentNumber, StudentName: student.Name, AssignmentName: assignment.Name, Status: submission.Status, Score: &score, MaximumScore: assignment.MaximumScore}
	for _, save := range []func() error{func() error { return repository.SaveStudent(student) }, func() error { return repository.SaveAssignment(assignment) }, func() error { return repository.SaveSubmission(submission) }, func() error { return repository.SaveSnapshot(snapshot) }} {
		if err := save(); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	reopenedRepository, err := NewRepository(reopened)
	if err != nil {
		t.Fatal(err)
	}
	state, err := reopenedRepository.ExportState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Students) != 1 || len(state.Assignments) != 1 || len(state.Submissions) != 1 || len(state.Snapshots) != 1 {
		t.Fatalf("state did not survive reopen: %#v", state)
	}
}
