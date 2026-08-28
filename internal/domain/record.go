package domain

import (
	"fmt"
	"strings"
)

type SubmissionRecord struct {
	Student    *Student
	Assignment *Assignment
	Submission *Submission
}

type GradeSnapshot struct {
	ID             string           `json:"id"`
	StudentID      string           `json:"student_id"`
	AssignmentID   string           `json:"assignment_id"`
	StudentNumber  string           `json:"student_number"`
	StudentName    string           `json:"student_name"`
	AssignmentName string           `json:"assignment_name"`
	Status         SubmissionStatus `json:"status"`
	Score          *int             `json:"score,omitempty"`
	MaximumScore   int              `json:"maximum_score"`
}

func NewSubmissionRecord(student *Student, assignment *Assignment, submission *Submission) (*SubmissionRecord, error) {
	if student == nil {
		return nil, fmt.Errorf("student is required")
	}
	if assignment == nil {
		return nil, fmt.Errorf("assignment is required")
	}
	if submission == nil {
		return nil, fmt.Errorf("submission is required")
	}
	if submission.StudentID != student.ID {
		return nil, fmt.Errorf("submission student does not match record student")
	}
	if submission.AssignmentID != assignment.ID {
		return nil, fmt.Errorf("submission assignment does not match record assignment")
	}
	return &SubmissionRecord{Student: student, Assignment: assignment, Submission: submission}, nil
}

func (r SubmissionRecord) Key() string {
	return SubmissionKey(r.Submission.StudentID, r.Submission.AssignmentID)
}

func (r SubmissionRecord) Snapshot(id string) (GradeSnapshot, error) {
	if r.Student == nil || r.Assignment == nil || r.Submission == nil {
		return GradeSnapshot{}, fmt.Errorf("complete submission record is required")
	}
	snapshot := GradeSnapshot{
		ID:             strings.TrimSpace(id),
		StudentID:      r.Student.ID,
		AssignmentID:   r.Assignment.ID,
		StudentNumber:  r.Student.StudentNumber,
		StudentName:    r.Student.Name,
		AssignmentName: r.Assignment.Name,
		Status:         r.Submission.Status,
		Score:          cloneScore(r.Submission.Score),
		MaximumScore:   r.Assignment.MaximumScore,
	}
	if err := snapshot.Validate(); err != nil {
		return GradeSnapshot{}, err
	}
	return snapshot, nil
}

func (s GradeSnapshot) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("snapshot id is required")
	}
	if s.StudentID == "" || s.AssignmentID == "" {
		return fmt.Errorf("snapshot references are required")
	}
	if s.StudentNumber == "" || s.StudentName == "" || s.AssignmentName == "" {
		return fmt.Errorf("snapshot display fields are required")
	}
	if !s.Status.Valid() {
		return fmt.Errorf("snapshot status is invalid")
	}
	if s.MaximumScore <= 0 {
		return fmt.Errorf("snapshot maximum score must be positive")
	}
	if s.Score != nil && (*s.Score < 0 || *s.Score > s.MaximumScore) {
		return fmt.Errorf("snapshot score is outside assignment range")
	}
	return nil
}

func (s GradeSnapshot) Percentage() (int, bool) {
	if s.Score == nil || s.MaximumScore <= 0 {
		return 0, false
	}
	return (*s.Score * 100) / s.MaximumScore, true
}

func SubmissionKey(studentID, assignmentID string) string {
	return strings.TrimSpace(studentID) + "\x00" + strings.TrimSpace(assignmentID)
}

func SplitSubmissionKey(key string) (string, string, bool) {
	parts := strings.SplitN(key, "\x00", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func cloneScore(score *int) *int {
	if score == nil {
		return nil
	}
	value := *score
	return &value
}
