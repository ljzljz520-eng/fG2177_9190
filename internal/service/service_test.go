package service

import (
	"testing"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/policy"
	"courseworkledger/internal/store"
)

func TestCatalogSearchAndDelete(t *testing.T) {
	database, err := store.Open(t.TempDir() + "/coursework.db")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	repository, err := store.NewRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog(repository)
	if err != nil {
		t.Fatal(err)
	}
	student, _ := domain.NewStudent("s1", "1001", "Ada", "", "2026")
	assignment, _ := domain.NewAssignment("a1", "Pointers", "", 100, true, nil)
	if err := catalog.RegisterStudent(student); err != nil {
		t.Fatal(err)
	}
	if err := catalog.RegisterAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Submit(student.ID, assignment.ID, "answer", false); err != nil {
		t.Fatal(err)
	}
	records, err := catalog.Search(domain.SubmissionQuery{GradedOnly: false})
	if err != nil || len(records) != 1 {
		t.Fatalf("search failed: %v", err)
	}
	statusPolicy := policy.NewStatusPolicy(true, false)
	if _, _, err := catalog.ChangeStatus(student.ID, assignment.ID, policy.ActionMarkLate, "", statusPolicy); err != nil {
		t.Fatal(err)
	}
	if err := catalog.DeleteSubmission(student.ID, assignment.ID); err != nil {
		t.Fatal(err)
	}
}
