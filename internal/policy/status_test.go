package policy

import (
	"testing"

	"courseworkledger/internal/domain"
)

func TestStatusPolicy(t *testing.T) {
	policy := NewStatusPolicy(true, true)
	decision := policy.Decide(domain.StatusSubmitted, ActionGrade, false)
	if decision.Allowed || decision.To != domain.StatusSubmitted {
		t.Fatal("grading without required feedback was allowed")
	}
	decision = policy.Decide(domain.StatusSubmitted, ActionGrade, true)
	if !decision.Allowed || decision.To != domain.StatusGraded {
		t.Fatal("grading with feedback was rejected")
	}
	if len(policy.NextActions(domain.StatusReturned, false)) != 2 {
		t.Fatal("unexpected actions for returned work")
	}
}
