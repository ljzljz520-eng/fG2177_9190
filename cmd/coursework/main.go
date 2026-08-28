package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/policy"
	"courseworkledger/internal/service"
	"courseworkledger/internal/store"
)

func main() {
	path := flag.String("db", filepath.Join(".", "coursework.db"), "database path")
	flag.Parse()
	database, err := store.Open(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer database.Close()
	repository, err := store.NewRepository(database)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	catalog, err := service.NewCatalog(repository)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if flag.NArg() == 0 {
		printHelp()
		return
	}
	if err := runCommand(catalog, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCommand(catalog *service.Catalog, args []string) error {
	switch args[0] {
	case "summary":
		summary, err := catalog.EnrollmentSummary()
		if err != nil {
			return err
		}
		fmt.Println(summary)
		return nil
	case "dashboard":
		dashboard, err := catalog.BuildDashboard()
		if err != nil {
			return err
		}
		fmt.Println(dashboard.RenderText())
		return nil
	case "add-student":
		if len(args) != 6 {
			return fmt.Errorf("usage: add-student id number name email cohort")
		}
		student, err := domain.NewStudent(args[1], args[2], args[3], args[4], args[5])
		if err != nil {
			return err
		}
		return catalog.RegisterStudent(student)
	case "add-assignment":
		if len(args) != 5 {
			return fmt.Errorf("usage: add-assignment id name maximum_score required")
		}
		var maximum int
		if _, err := fmt.Sscanf(args[3], "%d", &maximum); err != nil {
			return fmt.Errorf("maximum score must be an integer")
		}
		var required bool
		if _, err := fmt.Sscanf(args[4], "%t", &required); err != nil {
			return fmt.Errorf("required must be true or false")
		}
		assignment, err := domain.NewAssignment(args[1], args[2], "", maximum, required, nil)
		if err != nil {
			return err
		}
		return catalog.RegisterAssignment(assignment)
	case "submit":
		if len(args) < 4 || len(args) > 5 {
			return fmt.Errorf("usage: submit student_id assignment_id content [late]")
		}
		late := len(args) == 5 && args[4] == "late"
		submission, err := catalog.Submit(args[1], args[2], args[3], late)
		if err != nil {
			return err
		}
		fmt.Println(submission.Summary())
		return nil
	case "grade":
		if len(args) < 4 || len(args) > 5 {
			return fmt.Errorf("usage: grade student_id assignment_id score [feedback]")
		}
		var score int
		if _, err := fmt.Sscanf(args[3], "%d", &score); err != nil {
			return fmt.Errorf("score must be an integer")
		}
		feedback := ""
		if len(args) == 5 {
			feedback = args[4]
		}
		scorer, err := policy.NewScoringPolicy(60, nil)
		if err != nil {
			return err
		}
		submission, decision, err := catalog.Grade(args[1], args[2], score, feedback, scorer)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s %d%%\n", submission.ID, decision.Band.Name, decision.Percentage)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp() {
	fmt.Println("coursework commands: summary, dashboard, add-student, add-assignment, submit, grade")
}
