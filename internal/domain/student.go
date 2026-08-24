package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Student struct {
	ID            string `json:"id"`
	StudentNumber string `json:"student_number"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Cohort        string `json:"cohort"`
	Active        bool   `json:"active"`
}

type StudentPatch struct {
	Name   *string
	Email  *string
	Cohort *string
	Active *bool
}

type StudentFilter struct {
	Cohort       string
	NameContains string
	ActiveOnly   bool
}

func NewStudent(id, number, name, email, cohort string) (Student, error) {
	student := Student{
		ID:            strings.TrimSpace(id),
		StudentNumber: strings.TrimSpace(number),
		Name:          strings.TrimSpace(name),
		Email:         strings.TrimSpace(email),
		Cohort:        strings.TrimSpace(cohort),
		Active:        true,
	}
	if err := student.Validate(); err != nil {
		return Student{}, err
	}
	return student, nil
}

func (s Student) Validate() error {
	var problems []string
	if s.ID == "" {
		problems = append(problems, "student id is required")
	}
	if s.StudentNumber == "" {
		problems = append(problems, "student number is required")
	}
	if s.Name == "" {
		problems = append(problems, "student name is required")
	}
	if s.Email != "" && (!strings.Contains(s.Email, "@") || strings.HasPrefix(s.Email, "@") || strings.HasSuffix(s.Email, "@")) {
		problems = append(problems, "student email is invalid")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (s Student) DisplayName() string {
	if s.StudentNumber == "" {
		return s.Name
	}
	return fmt.Sprintf("%s (%s)", s.Name, s.StudentNumber)
}

func (s Student) Matches(filter StudentFilter) bool {
	if filter.ActiveOnly && !s.Active {
		return false
	}
	if filter.Cohort != "" && !strings.EqualFold(s.Cohort, filter.Cohort) {
		return false
	}
	needle := strings.ToLower(strings.TrimSpace(filter.NameContains))
	if needle != "" && !strings.Contains(strings.ToLower(s.Name), needle) {
		return false
	}
	return true
}

func (s Student) Apply(patch StudentPatch) (Student, error) {
	updated := s
	if patch.Name != nil {
		updated.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Email != nil {
		updated.Email = strings.TrimSpace(*patch.Email)
	}
	if patch.Cohort != nil {
		updated.Cohort = strings.TrimSpace(*patch.Cohort)
	}
	if patch.Active != nil {
		updated.Active = *patch.Active
	}
	if err := updated.Validate(); err != nil {
		return Student{}, err
	}
	return updated, nil
}

func SortStudents(students []Student) []Student {
	result := append([]Student(nil), students...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].StudentNumber == result[j].StudentNumber {
			return result[i].Name < result[j].Name
		}
		return result[i].StudentNumber < result[j].StudentNumber
	})
	return result
}

func FindStudentByNumber(students []Student, number string) (Student, bool) {
	target := strings.TrimSpace(number)
	for _, student := range students {
		if student.StudentNumber == target {
			return student, true
		}
	}
	return Student{}, false
}

func FilterStudents(students []Student, filter StudentFilter) []Student {
	result := make([]Student, 0, len(students))
	for _, student := range students {
		if student.Matches(filter) {
			result = append(result, student)
		}
	}
	return SortStudents(result)
}
