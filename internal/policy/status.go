package policy

import (
	"fmt"
	"sort"

	"courseworkledger/internal/domain"
)

type StatusAction string

const (
	ActionSubmit   StatusAction = "submit"
	ActionGrade    StatusAction = "grade"
	ActionReturn   StatusAction = "return"
	ActionRevise   StatusAction = "revise"
	ActionMarkLate StatusAction = "mark_late"
)

type StatusDecision struct {
	Allowed bool
	From    domain.SubmissionStatus
	To      domain.SubmissionStatus
	Action  StatusAction
	Reason  string
}

type StatusPolicy struct {
	allowLateSubmission bool
	requireFeedback     bool
}

func NewStatusPolicy(allowLateSubmission, requireFeedback bool) StatusPolicy {
	return StatusPolicy{
		allowLateSubmission: allowLateSubmission,
		requireFeedback:     requireFeedback,
	}
}

func (p StatusPolicy) Decide(current domain.SubmissionStatus, action StatusAction, hasFeedback bool) StatusDecision {
	decision := StatusDecision{From: current, Action: action, To: current}
	if !current.Valid() {
		decision.Reason = "current submission status is invalid"
		return decision
	}
	switch action {
	case ActionSubmit:
		if current == domain.StatusNotSubmitted || current == domain.StatusReturned {
			decision.Allowed = true
			decision.To = domain.StatusSubmitted
			decision.Reason = "submission accepted"
		} else {
			decision.Reason = "only missing or returned work can be submitted"
		}
	case ActionMarkLate:
		if !p.allowLateSubmission {
			decision.Reason = "late submissions are disabled"
		} else if current == domain.StatusNotSubmitted || current == domain.StatusSubmitted {
			decision.Allowed = true
			decision.To = domain.StatusLate
			decision.Reason = "submission marked late"
		} else {
			decision.Reason = "only new submissions can be marked late"
		}
	case ActionGrade:
		if current != domain.StatusSubmitted && current != domain.StatusLate {
			decision.Reason = "only submitted work can be graded"
		} else if p.requireFeedback && !hasFeedback {
			decision.Reason = "feedback is required before grading"
		} else {
			decision.Allowed = true
			decision.To = domain.StatusGraded
			decision.Reason = "grade accepted"
		}
	case ActionReturn:
		if current != domain.StatusGraded {
			decision.Reason = "only graded work can be returned"
		} else {
			decision.Allowed = true
			decision.To = domain.StatusReturned
			decision.Reason = "graded work returned"
		}
	case ActionRevise:
		if current != domain.StatusReturned && current != domain.StatusGraded {
			decision.Reason = "only graded or returned work can be revised"
		} else {
			decision.Allowed = true
			decision.To = domain.StatusSubmitted
			decision.Reason = "revision opened"
		}
	default:
		decision.Reason = fmt.Sprintf("unknown status action %q", action)
	}
	return decision
}

func (p StatusPolicy) NextActions(status domain.SubmissionStatus, hasFeedback bool) []StatusAction {
	candidates := []StatusAction{ActionSubmit, ActionMarkLate, ActionGrade, ActionReturn, ActionRevise}
	result := make([]StatusAction, 0, len(candidates))
	for _, action := range candidates {
		if p.Decide(status, action, hasFeedback).Allowed {
			result = append(result, action)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (p StatusPolicy) Apply(submission domain.Submission, action StatusAction, feedback string) (domain.Submission, StatusDecision, error) {
	decision := p.Decide(submission.Status, action, feedback != "")
	if !decision.Allowed {
		return domain.Submission{}, decision, fmt.Errorf("status action rejected: %s", decision.Reason)
	}
	status := decision.To
	patch := domain.SubmissionPatch{Status: &status}
	if feedback != "" {
		patch.Feedback = &feedback
	}
	updated, err := submission.Apply(patch)
	if err != nil {
		return domain.Submission{}, decision, err
	}
	return updated, decision, nil
}
