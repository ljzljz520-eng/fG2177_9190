package importer

import (
	"strings"
	"testing"
)

func TestParseStudents(t *testing.T) {
	input := "id,student_number,name,email,cohort,active\n" +
		"s2,1002,Grace,grace@example.test,2026,true\n" +
		"s1,1001,Ada,ada@example.test,2026,true\n" +
		"s1,1001,Duplicate,dup@example.test,2026,true\n"
	result, err := ParseStudents(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Students) != 2 || len(result.Issues) != 1 {
		t.Fatalf("unexpected import result: %#v", result)
	}
	if result.Students[0].StudentNumber != "1001" {
		t.Fatal("students were not sorted")
	}
}

func TestParseAssignments(t *testing.T) {
	input := "id,name,description,maximum_score,required,tags\n" +
		"a1,Pointers,memory,100,true,go|memory\n"
	result, err := ParseAssignments(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Assignments) != 1 || !result.Assignments[0].HasTag("memory") {
		t.Fatalf("unexpected assignment result: %#v", result)
	}
}
