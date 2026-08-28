package domain

import "testing"

func TestSubmissionListLifecycle(t *testing.T) {
	student, err := NewStudent("s1", "1001", "Ada", "", "2026")
	if err != nil {
		t.Fatal(err)
	}
	assignment, err := NewAssignment("a1", "Pointers", "", 100, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	submission, err := NewSubmission("sub", student.ID, assignment.ID, "answer", false)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewSubmissionRecord(&student, &assignment, &submission)
	if err != nil {
		t.Fatal(err)
	}
	list := NewSubmissionList()
	if err := list.Append(record); err != nil {
		t.Fatal(err)
	}
	if list.Len() != 1 || list.FindByStudentAndAssignment("s1", "a1") == nil {
		t.Fatal("list did not index record")
	}
	if !list.MoveToEnd("s1", "a1") {
		t.Fatal("list did not move record")
	}
	if err := list.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, ok := list.Delete("s1", "a1"); !ok || list.Len() != 0 {
		t.Fatal("list did not delete record")
	}
}
