package domain

import (
	"sort"
	"strings"
)

type SubmissionQuery struct {
	StudentID       string
	AssignmentID    string
	Statuses        []SubmissionStatus
	MinimumScore    *int
	MaximumScore    *int
	GradedOnly      bool
	ContentContains string
}

type SubmissionSummary struct {
	Total      int
	Submitted  int
	Missing    int
	Graded     int
	Late       int
	ScoreTotal int
	ScoreCount int
}

func QuerySubmissions(list *SubmissionList, query SubmissionQuery) []*SubmissionRecord {
	if list == nil {
		return nil
	}
	statuses := make(map[SubmissionStatus]struct{}, len(query.Statuses))
	for _, status := range query.Statuses {
		statuses[status] = struct{}{}
	}
	needle := strings.ToLower(strings.TrimSpace(query.ContentContains))
	result := make([]*SubmissionRecord, 0, list.Len())
	for _, record := range list.Records() {
		if !recordMatches(record, query, statuses, needle) {
			continue
		}
		result = append(result, record)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Student.StudentNumber != result[j].Student.StudentNumber {
			return result[i].Student.StudentNumber < result[j].Student.StudentNumber
		}
		if result[i].Assignment.Name != result[j].Assignment.Name {
			return result[i].Assignment.Name < result[j].Assignment.Name
		}
		return result[i].Submission.ID < result[j].Submission.ID
	})
	return result
}

func SummarizeSubmissions(students []Student, assignments []Assignment, list *SubmissionList) SubmissionSummary {
	summary := SubmissionSummary{Total: len(students) * len(assignments)}
	for _, student := range students {
		if !student.Active {
			continue
		}
		for _, assignment := range assignments {
			record := list.FindByStudentAndAssignment(student.ID, assignment.ID)
			if record == nil || record.Submission == nil || !record.Submission.IsSubmitted() {
				summary.Missing++
				continue
			}
			summary.Submitted++
			if record.Submission.Status == StatusLate {
				summary.Late++
			}
			if record.Submission.Score != nil {
				summary.Graded++
				summary.ScoreTotal += *record.Submission.Score
				summary.ScoreCount++
			}
		}
	}
	return summary
}

func (s SubmissionSummary) AverageScore() (float64, bool) {
	if s.ScoreCount == 0 {
		return 0, false
	}
	return float64(s.ScoreTotal) / float64(s.ScoreCount), true
}

func recordMatches(record *SubmissionRecord, query SubmissionQuery, statuses map[SubmissionStatus]struct{}, needle string) bool {
	if record == nil || record.Submission == nil || record.Student == nil || record.Assignment == nil {
		return false
	}
	submission := record.Submission
	if query.StudentID != "" && submission.StudentID != query.StudentID {
		return false
	}
	if query.AssignmentID != "" && submission.AssignmentID != query.AssignmentID {
		return false
	}
	if len(statuses) > 0 {
		if _, exists := statuses[submission.Status]; !exists {
			return false
		}
	}
	if query.GradedOnly && !submission.IsGraded() {
		return false
	}
	if query.MinimumScore != nil && (submission.Score == nil || *submission.Score < *query.MinimumScore) {
		return false
	}
	if query.MaximumScore != nil && (submission.Score == nil || *submission.Score > *query.MaximumScore) {
		return false
	}
	if needle != "" && !strings.Contains(strings.ToLower(submission.Content), needle) {
		return false
	}
	return true
}
