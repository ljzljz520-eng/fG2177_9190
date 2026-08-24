package domain

import "testing"

func TestStudentValidationAndFilter(t *testing.T) {
	student, err := NewStudent("s1", "1001", "Ada", "ada@example.test", "2026")
	if err != nil {
		t.Fatal(err)
	}
	if student.DisplayName() != "Ada (1001)" {
		t.Fatalf("unexpected display name %q", student.DisplayName())
	}
	updated, err := student.Apply(StudentPatch{Name: stringPointer("Ada Lovelace")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Ada Lovelace" {
		t.Fatal("student patch was not applied")
	}
	if !student.Matches(StudentFilter{NameContains: "ad"}) {
		t.Fatal("student filter did not match")
	}
	if _, err := NewStudent("", "", "", "bad", ""); err == nil {
		t.Fatal("invalid student was accepted")
	}
}

func stringPointer(value string) *string {
	return &value
}
