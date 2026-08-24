package service

import (
	"fmt"
	"sort"
	"strings"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/importer"
	"courseworkledger/internal/store"
)

type Catalog struct {
	repository *store.Repository
}

func NewCatalog(repository *store.Repository) (*Catalog, error) {
	if repository == nil {
		return nil, fmt.Errorf("repository is required")
	}
	return &Catalog{repository: repository}, nil
}

func (c *Catalog) Repository() *store.Repository {
	if c == nil {
		return nil
	}
	return c.repository
}

func (c *Catalog) RegisterStudent(student domain.Student) error {
	if err := student.Validate(); err != nil {
		return err
	}
	return c.repository.SaveStudent(student)
}

func (c *Catalog) RegisterStudents(students []domain.Student) (int, error) {
	count := 0
	for _, student := range domain.SortStudents(students) {
		if err := c.RegisterStudent(student); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (c *Catalog) ImportStudents(result importer.StudentImportResult) (int, error) {
	if len(result.Issues) > 0 {
		return 0, fmt.Errorf("student import contains issues: %s", importer.FormatStudentIssues(result.Issues))
	}
	return c.RegisterStudents(result.Students)
}

func (c *Catalog) UpdateStudent(id string, patch domain.StudentPatch) (domain.Student, error) {
	student, found, err := c.repository.Student(id)
	if err != nil {
		return domain.Student{}, err
	}
	if !found {
		return domain.Student{}, fmt.Errorf("student %s was not found", id)
	}
	updated, err := student.Apply(patch)
	if err != nil {
		return domain.Student{}, err
	}
	if err := c.repository.SaveStudent(updated); err != nil {
		return domain.Student{}, err
	}
	return updated, nil
}

func (c *Catalog) FindStudentByNumber(number string) (domain.Student, error) {
	student, found, err := c.repository.StudentByNumber(strings.TrimSpace(number))
	if err != nil {
		return domain.Student{}, err
	}
	if !found {
		return domain.Student{}, fmt.Errorf("student number %s was not found", number)
	}
	return student, nil
}

func (c *Catalog) Students(filter domain.StudentFilter) ([]domain.Student, error) {
	students, err := c.repository.Students()
	if err != nil {
		return nil, err
	}
	return domain.FilterStudents(students, filter), nil
}

func (c *Catalog) RemoveStudent(id string) error {
	deleted, err := c.repository.DeleteStudent(id)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("student %s was not found", id)
	}
	return nil
}

func (c *Catalog) RegisterAssignment(assignment domain.Assignment) error {
	if err := assignment.Validate(); err != nil {
		return err
	}
	return c.repository.SaveAssignment(assignment)
}

func (c *Catalog) RegisterAssignments(assignments []domain.Assignment) (int, error) {
	count := 0
	for _, assignment := range domain.SortAssignments(assignments) {
		if err := c.RegisterAssignment(assignment); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (c *Catalog) ImportAssignments(result importer.AssignmentImportResult) (int, error) {
	if len(result.Issues) > 0 {
		return 0, fmt.Errorf("assignment import contains issues: %s", importer.FormatAssignmentIssues(result.Issues))
	}
	return c.RegisterAssignments(result.Assignments)
}

func (c *Catalog) UpdateAssignment(id string, patch domain.AssignmentPatch) (domain.Assignment, error) {
	assignment, found, err := c.repository.Assignment(id)
	if err != nil {
		return domain.Assignment{}, err
	}
	if !found {
		return domain.Assignment{}, fmt.Errorf("assignment %s was not found", id)
	}
	updated, err := assignment.Apply(patch)
	if err != nil {
		return domain.Assignment{}, err
	}
	if err := c.repository.SaveAssignment(updated); err != nil {
		return domain.Assignment{}, err
	}
	return updated, nil
}

func (c *Catalog) Assignments(tag string) ([]domain.Assignment, error) {
	assignments, err := c.repository.Assignments()
	if err != nil {
		return nil, err
	}
	return domain.FilterAssignmentsByTag(assignments, tag), nil
}

func (c *Catalog) RemoveAssignment(id string) error {
	deleted, err := c.repository.DeleteAssignment(id)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("assignment %s was not found", id)
	}
	return nil
}

func (c *Catalog) EnrollmentSummary() (string, error) {
	students, err := c.Students(domain.StudentFilter{ActiveOnly: true})
	if err != nil {
		return "", err
	}
	assignments, err := c.Assignments("")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d active students, %d assignments", len(students), len(assignments)), nil
}

func sortStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
