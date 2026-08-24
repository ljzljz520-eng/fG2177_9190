package domain

import "testing"

func TestAssignmentRulesAndSorting(t *testing.T) {
	assignment, err := NewAssignment("a1", "Pointers", "memory", 100, true, []string{"go", "Go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(assignment.Tags) != 1 || !assignment.HasTag("GO") {
		t.Fatalf("tags were not normalized: %#v", assignment.Tags)
	}
	if assignment.Label() != "Pointers [100 points, required]" {
		t.Fatalf("unexpected label %q", assignment.Label())
	}
	optional, err := NewAssignment("a2", "Reading", "", 10, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	sorted := SortAssignments([]Assignment{optional, assignment})
	if sorted[0].ID != "a1" {
		t.Fatal("required assignment was not first")
	}
	if _, err := NewAssignment("bad", "", "", 0, false, nil); err == nil {
		t.Fatal("invalid assignment was accepted")
	}
}
