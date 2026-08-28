package domain

import (
	"errors"
	"fmt"
	"strings"
)

type SubmissionStatus string

const (
	StatusNotSubmitted SubmissionStatus = "not_submitted"
	StatusSubmitted    SubmissionStatus = "submitted"
	StatusLate         SubmissionStatus = "late"
	StatusGraded       SubmissionStatus = "graded"
	StatusReturned     SubmissionStatus = "returned"
)

type Submission struct {
	ID           string           `json:"id"`
	StudentID    string           `json:"student_id"`
	AssignmentID string           `json:"assignment_id"`
	Status       SubmissionStatus `json:"status"`
	Content      string           `json:"content"`
	Score        *int             `json:"score,omitempty"`
	Feedback     string           `json:"feedback"`
	Revision     int              `json:"revision"`
}

type SubmissionPatch struct {
	Status   *SubmissionStatus
	Content  *string
	Score    *int
	Feedback *string
}

func NewSubmission(id, studentID, assignmentID, content string, late bool) (Submission, error) {
	status := StatusSubmitted
	if late {
		status = StatusLate
	}
	submission := Submission{
		ID:           strings.TrimSpace(id),
		StudentID:    strings.TrimSpace(studentID),
		AssignmentID: strings.TrimSpace(assignmentID),
		Status:       status,
		Content:      strings.TrimSpace(content),
		Revision:     1,
	}
	if err := submission.Validate(); err != nil {
		return Submission{}, err
	}
	return submission, nil
}

func (s Submission) Validate() error {
	var problems []string
	if s.ID == "" {
		problems = append(problems, "submission id is required")
	}
	if s.StudentID == "" {
		problems = append(problems, "submission student is required")
	}
	if s.AssignmentID == "" {
		problems = append(problems, "submission assignment is required")
	}
	if !s.Status.Valid() {
		problems = append(problems, "submission status is invalid")
	}
	if s.Status != StatusNotSubmitted && strings.TrimSpace(s.Content) == "" {
		problems = append(problems, "submitted work must include content")
	}
	if s.Score != nil && *s.Score < 0 {
		problems = append(problems, "score must not be negative")
	}
	if s.Revision < 1 {
		problems = append(problems, "revision must be positive")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (s SubmissionStatus) Valid() bool {
	switch s {
	case StatusNotSubmitted, StatusSubmitted, StatusLate, StatusGraded, StatusReturned:
		return true
	default:
		return false
	}
}

func (s Submission) IsSubmitted() bool {
	switch s.Status {
	case StatusSubmitted, StatusLate, StatusGraded, StatusReturned:
		return true
	default:
		return false
	}
}

func (s Submission) IsGraded() bool {
	return s.Score != nil && (s.Status == StatusGraded || s.Status == StatusReturned)
}

func (s Submission) Summary() string {
	score := "ungraded"
	if s.Score != nil {
		score = fmt.Sprintf("%d", *s.Score)
	}
	return fmt.Sprintf("%s: %s, score %s, revision %d", s.ID, s.Status, score, s.Revision)
}

func (s Submission) Apply(patch SubmissionPatch) (Submission, error) {
	updated := s
	if patch.Status != nil {
		if !canTransition(s.Status, *patch.Status) {
			return Submission{}, fmt.Errorf("cannot transition submission from %s to %s", s.Status, *patch.Status)
		}
		updated.Status = *patch.Status
	}
	if patch.Content != nil {
		updated.Content = strings.TrimSpace(*patch.Content)
		if updated.Content != s.Content {
			updated.Revision++
			updated.Score = nil
			updated.Feedback = ""
			if updated.Status == StatusGraded || updated.Status == StatusReturned {
				updated.Status = StatusSubmitted
			}
		}
	}
	if patch.Score != nil {
		score := *patch.Score
		updated.Score = &score
	}
	if patch.Feedback != nil {
		updated.Feedback = strings.TrimSpace(*patch.Feedback)
	}
	if err := updated.Validate(); err != nil {
		return Submission{}, err
	}
	return updated, nil
}

func canTransition(from, to SubmissionStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusNotSubmitted:
		return to == StatusSubmitted || to == StatusLate
	case StatusSubmitted, StatusLate:
		return to == StatusGraded || to == StatusReturned || (from == StatusSubmitted && to == StatusLate)
	case StatusGraded:
		return to == StatusReturned || to == StatusSubmitted
	case StatusReturned:
		return to == StatusSubmitted || to == StatusGraded
	default:
		return false
	}
}
