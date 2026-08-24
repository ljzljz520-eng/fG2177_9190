package service

import (
	"fmt"
	"strings"

	"courseworkledger/internal/domain"
	"courseworkledger/internal/policy"
)

func (c *Catalog) Submit(studentID, assignmentID, content string, late bool) (domain.Submission, error) {
	student, found, err := c.repository.Student(studentID)
	if err != nil {
		return domain.Submission{}, err
	}
	if !found {
		return domain.Submission{}, fmt.Errorf("student %s was not found", studentID)
	}
	if !student.Active {
		return domain.Submission{}, fmt.Errorf("student %s is inactive", studentID)
	}
	_, found, err = c.repository.Assignment(assignmentID)
	if err != nil {
		return domain.Submission{}, err
	}
	if !found {
		return domain.Submission{}, fmt.Errorf("assignment %s was not found", assignmentID)
	}
	if existing, exists, findErr := c.repository.Submission(studentID, assignmentID); findErr != nil {
		return domain.Submission{}, findErr
	} else if exists {
		if existing.Status == domain.StatusGraded || existing.Status == domain.StatusReturned {
			return domain.Submission{}, fmt.Errorf("submission already graded; revise it before resubmitting")
		}
		return c.ReviseSubmission(studentID, assignmentID, content, late)
	}
	id := fmt.Sprintf("sub-%s-%s", studentID, assignmentID)
	submission, err := domain.NewSubmission(id, studentID, assignmentID, content, late)
	if err != nil {
		return domain.Submission{}, err
	}
	if err := c.repository.SaveSubmission(submission); err != nil {
		return domain.Submission{}, err
	}
	return submission, nil
}

func (c *Catalog) ReviseSubmission(studentID, assignmentID, content string, late bool) (domain.Submission, error) {
	submission, found, err := c.repository.Submission(studentID, assignmentID)
	if err != nil {
		return domain.Submission{}, err
	}
	if !found {
		return c.Submit(studentID, assignmentID, content, late)
	}
	status := domain.StatusSubmitted
	if late {
		status = domain.StatusLate
	}
	patch := domain.SubmissionPatch{Status: &status, Content: &content}
	updated, err := submission.Apply(patch)
	if err != nil {
		return domain.Submission{}, err
	}
	if err := c.repository.SaveSubmission(updated); err != nil {
		return domain.Submission{}, err
	}
	return updated, nil
}

func (c *Catalog) Grade(studentID, assignmentID string, score int, feedback string, scorer policy.ScoringPolicy) (domain.Submission, policy.ScoreDecision, error) {
	submission, found, err := c.repository.Submission(studentID, assignmentID)
	if err != nil {
		return domain.Submission{}, policy.ScoreDecision{}, err
	}
	if !found {
		return domain.Submission{}, policy.ScoreDecision{}, fmt.Errorf("submission for %s and %s was not found", studentID, assignmentID)
	}
	assignment, found, err := c.repository.Assignment(assignmentID)
	if err != nil {
		return domain.Submission{}, policy.ScoreDecision{}, err
	}
	if !found {
		return domain.Submission{}, policy.ScoreDecision{}, fmt.Errorf("assignment %s was not found", assignmentID)
	}
	updated, decision, err := scorer.Grade(submission, assignment, score, strings.TrimSpace(feedback))
	if err != nil {
		return domain.Submission{}, policy.ScoreDecision{}, err
	}
	if err := c.repository.SaveSubmission(updated); err != nil {
		return domain.Submission{}, policy.ScoreDecision{}, err
	}
	return updated, decision, nil
}

func (c *Catalog) ChangeStatus(studentID, assignmentID string, action policy.StatusAction, feedback string, statusPolicy policy.StatusPolicy) (domain.Submission, policy.StatusDecision, error) {
	submission, found, err := c.repository.Submission(studentID, assignmentID)
	if err != nil {
		return domain.Submission{}, policy.StatusDecision{}, err
	}
	if !found {
		return domain.Submission{}, policy.StatusDecision{}, fmt.Errorf("submission for %s and %s was not found", studentID, assignmentID)
	}
	updated, decision, err := statusPolicy.Apply(submission, action, strings.TrimSpace(feedback))
	if err != nil {
		return domain.Submission{}, decision, err
	}
	if err := c.repository.SaveSubmission(updated); err != nil {
		return domain.Submission{}, decision, err
	}
	return updated, decision, nil
}

func (c *Catalog) DeleteSubmission(studentID, assignmentID string) error {
	deleted, err := c.repository.DeleteSubmission(studentID, assignmentID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("submission for %s and %s was not found", studentID, assignmentID)
	}
	return nil
}

func (c *Catalog) SubmissionList() (*domain.SubmissionList, error) {
	return c.repository.BuildSubmissionList()
}

func (c *Catalog) Search(query domain.SubmissionQuery) ([]*domain.SubmissionRecord, error) {
	list, err := c.SubmissionList()
	if err != nil {
		return nil, err
	}
	return domain.QuerySubmissions(list, query), nil
}

func (c *Catalog) SubmissionSummary() (domain.SubmissionSummary, error) {
	students, err := c.repository.Students()
	if err != nil {
		return domain.SubmissionSummary{}, err
	}
	assignments, err := c.repository.Assignments()
	if err != nil {
		return domain.SubmissionSummary{}, err
	}
	list, err := c.SubmissionList()
	if err != nil {
		return domain.SubmissionSummary{}, err
	}
	return domain.SummarizeSubmissions(students, assignments, list), nil
}

func (c *Catalog) RequireRecord(studentID, assignmentID string) *domain.SubmissionRecord {
	list, err := c.SubmissionList()
	if err != nil {
		return nil
	}
	return list.FindByStudentAndAssignment(studentID, assignmentID)
}
