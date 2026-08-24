package main

import (
	"strings"
	"testing"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/policy"
	"courseworkledger/internal/report"
	"courseworkledger/internal/service"
	"courseworkledger/internal/store"
)

func openTestCatalog(t *testing.T) (*service.Catalog, *store.Database) {
	t.Helper()
	database, err := store.Open(t.TempDir() + "/coursework.db")
	if err != nil {
		t.Fatal(err)
	}
	repository, err := store.NewRepository(database)
	if err != nil {
		database.Close()
		t.Fatal(err)
	}
	catalog, err := service.NewCatalog(repository)
	if err != nil {
		database.Close()
		t.Fatal(err)
	}
	return catalog, database
}

func sampleStudent(t *testing.T, id, number, name string) domain.Student {
	t.Helper()
	student, err := domain.NewStudent(id, number, name, number+"@school.test", "2026")
	if err != nil {
		t.Fatal(err)
	}
	return student
}

func sampleAssignment(t *testing.T) domain.Assignment {
	t.Helper()
	assignment, err := domain.NewAssignment("a1", "Pointers", "", 100, true, []string{"go"})
	if err != nil {
		t.Fatal(err)
	}
	return assignment
}

func TestWorkflowStudentIntakeAndRestore(t *testing.T) {
	path := t.TempDir() + "/coursework.db"
	database, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := store.NewRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := service.NewCatalog(repository)
	if err != nil {
		t.Fatal(err)
	}
	student := sampleStudent(t, "s1", "1001", "Ada")
	assignment := sampleAssignment(t)
	if err := catalog.RegisterStudent(student); err != nil {
		t.Fatal(err)
	}
	if err := catalog.RegisterAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Submit(student.ID, assignment.ID, "linked list", false); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	reopenedRepository, err := store.NewRepository(reopened)
	if err != nil {
		t.Fatal(err)
	}
	restoredCatalog, err := service.NewCatalog(reopenedRepository)
	if err != nil {
		t.Fatal(err)
	}
	restoredList, err := restoredCatalog.SubmissionList()
	if err != nil {
		t.Fatal(err)
	}
	if restoredList.Len() != 1 {
		t.Fatalf("expected one restored submission, got %d", restoredList.Len())
	}
}

func TestWorkflowSubmissionGradingAndDashboard(t *testing.T) {
	catalog, database := openTestCatalog(t)
	defer database.Close()
	student := sampleStudent(t, "s1", "1001", "Ada")
	if err := catalog.RegisterStudent(student); err != nil {
		t.Fatal(err)
	}
	assignment := sampleAssignment(t)
	if err := catalog.RegisterAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Submit(student.ID, assignment.ID, "solution", false); err != nil {
		t.Fatal(err)
	}
	scorer, err := policy.NewScoringPolicy(60, nil)
	if err != nil {
		t.Fatal(err)
	}
	graded, decision, err := catalog.Grade(student.ID, assignment.ID, 92, "clear", scorer)
	if err != nil {
		t.Fatal(err)
	}
	if graded.Status != domain.StatusGraded || decision.Band.Name != "excellent" {
		t.Fatalf("unexpected grade: %#v %#v", graded, decision)
	}
	dashboard, err := catalog.BuildDashboard()
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Average != "92%" || dashboard.Graded != 1 {
		t.Fatalf("unexpected dashboard: %#v", dashboard)
	}
	document, err := catalog.BuildReportDocument()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(document.Render(), "Pointers") {
		t.Fatal("report did not include assignment")
	}
}

func TestWorkflowStatusCorrectionAndExport(t *testing.T) {
	catalog, database := openTestCatalog(t)
	defer database.Close()
	student := sampleStudent(t, "s1", "1001", "Ada")
	assignment := sampleAssignment(t)
	if err := catalog.RegisterStudent(student); err != nil {
		t.Fatal(err)
	}
	if err := catalog.RegisterAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Submit(student.ID, assignment.ID, "first", false); err != nil {
		t.Fatal(err)
	}
	scorer, err := policy.NewScoringPolicy(60, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := catalog.Grade(student.ID, assignment.ID, 80, "revise", scorer); err != nil {
		t.Fatal(err)
	}
	statusPolicy := policy.NewStatusPolicy(true, false)
	if _, _, err := catalog.ChangeStatus(student.ID, assignment.ID, policy.ActionReturn, "revise section two", statusPolicy); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ReviseSubmission(student.ID, assignment.ID, "second", false); err != nil {
		t.Fatal(err)
	}
	dashboard, err := catalog.BuildDashboard()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.SaveDashboardSnapshots(dashboard); err != nil {
		t.Fatal(err)
	}
	data, err := catalog.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := service.ParseExport(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := parsed["snapshots"]; !ok {
		t.Fatal("export did not include snapshots")
	}
}

func TestMissingSubmissionReturnsFriendlyResult(t *testing.T) {
	catalog, database := openTestCatalog(t)
	defer database.Close()
	if err := catalog.RegisterStudent(sampleStudent(t, "s1", "1001", "Ada")); err != nil {
		t.Fatal(err)
	}
	if err := catalog.RegisterStudent(sampleStudent(t, "s2", "1002", "Grace")); err != nil {
		t.Fatal(err)
	}
	if err := catalog.RegisterAssignment(sampleAssignment(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Submit("s1", "a1", "solution", false); err != nil {
		t.Fatal(err)
	}
	dashboard, err := catalog.BuildDashboard()
	if err != nil {
		t.Fatalf("missing submissions should render a friendly warning: %v", err)
	}
	if len(dashboard.Warnings) != 1 || !strings.Contains(dashboard.Warnings[0], "1002") {
		t.Fatalf("unexpected warnings: %#v", dashboard.Warnings)
	}
}

func TestReportFormatting(t *testing.T) {
	student := sampleStudent(t, "s1", "1001", "Ada")
	assignment := sampleAssignment(t)
	submission, err := domain.NewSubmission("sub", student.ID, assignment.ID, "answer", false)
	if err != nil {
		t.Fatal(err)
	}
	record, err := domain.NewSubmissionRecord(&student, &assignment, &submission)
	if err != nil {
		t.Fatal(err)
	}
	line, err := report.NewResultLine(record)
	if err != nil {
		t.Fatal(err)
	}
	document := report.BuildDocument("results", []report.ResultLine{line}, "1 record")
	if !strings.Contains(document.Render(), "results") {
		t.Fatal("report title missing")
	}
}
