package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Assignment struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	MaximumScore int      `json:"maximum_score"`
	Required     bool     `json:"required"`
	Tags         []string `json:"tags"`
}

type AssignmentPatch struct {
	Name         *string
	Description  *string
	MaximumScore *int
	Required     *bool
	Tags         []string
}

func NewAssignment(id, name, description string, maximumScore int, required bool, tags []string) (Assignment, error) {
	assignment := Assignment{
		ID:           strings.TrimSpace(id),
		Name:         strings.TrimSpace(name),
		Description:  strings.TrimSpace(description),
		MaximumScore: maximumScore,
		Required:     required,
		Tags:         normalizeTags(tags),
	}
	if err := assignment.Validate(); err != nil {
		return Assignment{}, err
	}
	return assignment, nil
}

func (a Assignment) Validate() error {
	var problems []string
	if a.ID == "" {
		problems = append(problems, "assignment id is required")
	}
	if a.Name == "" {
		problems = append(problems, "assignment name is required")
	}
	if a.MaximumScore <= 0 {
		problems = append(problems, "maximum score must be positive")
	}
	if a.MaximumScore > 1000 {
		problems = append(problems, "maximum score must not exceed 1000")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (a Assignment) Label() string {
	requirement := "optional"
	if a.Required {
		requirement = "required"
	}
	return fmt.Sprintf("%s [%d points, %s]", a.Name, a.MaximumScore, requirement)
}

func (a Assignment) HasTag(tag string) bool {
	target := strings.ToLower(strings.TrimSpace(tag))
	for _, candidate := range a.Tags {
		if strings.ToLower(candidate) == target {
			return true
		}
	}
	return false
}

func (a Assignment) Apply(patch AssignmentPatch) (Assignment, error) {
	updated := a
	if patch.Name != nil {
		updated.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Description != nil {
		updated.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.MaximumScore != nil {
		updated.MaximumScore = *patch.MaximumScore
	}
	if patch.Required != nil {
		updated.Required = *patch.Required
	}
	if patch.Tags != nil {
		updated.Tags = normalizeTags(patch.Tags)
	}
	if err := updated.Validate(); err != nil {
		return Assignment{}, err
	}
	return updated, nil
}

func SortAssignments(assignments []Assignment) []Assignment {
	result := append([]Assignment(nil), assignments...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Required != result[j].Required {
			return result[i].Required
		}
		if result[i].Name == result[j].Name {
			return result[i].ID < result[j].ID
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func FilterAssignmentsByTag(assignments []Assignment, tag string) []Assignment {
	if strings.TrimSpace(tag) == "" {
		return SortAssignments(assignments)
	}
	result := make([]Assignment, 0, len(assignments))
	for _, assignment := range assignments {
		if assignment.HasTag(tag) {
			result = append(result, assignment)
		}
	}
	return SortAssignments(result)
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		cleaned := strings.TrimSpace(tag)
		key := strings.ToLower(cleaned)
		if cleaned == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, cleaned)
	}
	sort.Strings(result)
	return result
}
