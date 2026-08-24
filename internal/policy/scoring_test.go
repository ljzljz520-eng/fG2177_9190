package policy

import (
	"testing"

	"courseworkledger/internal/domain"
)

func TestScoringPolicy(t *testing.T) {
	policy, err := NewScoringPolicy(60, nil)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := policy.Evaluate(85, 100)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Band.Name != "strong" || !decision.Passing {
		t.Fatalf("unexpected score decision: %#v", decision)
	}
	student, _ := domain.NewStudent("s1", "1001", "Ada", "", "2026")
	assignment, _ := domain.NewAssignment("a1", "Pointers", "", 100, true, nil)
	submission, _ := domain.NewSubmission("sub", student.ID, assignment.ID, "answer", false)
	graded, _, err := policy.Grade(submission, assignment, 90, "good")
	if err != nil || graded.Status != domain.StatusGraded {
		t.Fatalf("grade application failed: %v", err)
	}
}
