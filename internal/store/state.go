package store

import (
	"fmt"
	"sort"

	"courseworkledger/internal/domain"
	bolt "go.etcd.io/bbolt"
)

type State struct {
	Students    []domain.Student       `json:"students"`
	Assignments []domain.Assignment    `json:"assignments"`
	Submissions []domain.Submission    `json:"submissions"`
	Snapshots   []domain.GradeSnapshot `json:"snapshots"`
}

type Counts struct {
	Students    int
	Assignments int
	Submissions int
	Snapshots   int
}

func (s State) Validate() error {
	studentIDs := make(map[string]struct{}, len(s.Students))
	studentNumbers := make(map[string]struct{}, len(s.Students))
	for _, student := range s.Students {
		if err := student.Validate(); err != nil {
			return err
		}
		if _, exists := studentIDs[student.ID]; exists {
			return fmt.Errorf("duplicate student id %s", student.ID)
		}
		if _, exists := studentNumbers[student.StudentNumber]; exists {
			return fmt.Errorf("duplicate student number %s", student.StudentNumber)
		}
		studentIDs[student.ID] = struct{}{}
		studentNumbers[student.StudentNumber] = struct{}{}
	}
	assignmentIDs := make(map[string]struct{}, len(s.Assignments))
	for _, assignment := range s.Assignments {
		if err := assignment.Validate(); err != nil {
			return err
		}
		if _, exists := assignmentIDs[assignment.ID]; exists {
			return fmt.Errorf("duplicate assignment id %s", assignment.ID)
		}
		assignmentIDs[assignment.ID] = struct{}{}
	}
	submissionKeys := make(map[string]struct{}, len(s.Submissions))
	for _, submission := range s.Submissions {
		if err := submission.Validate(); err != nil {
			return err
		}
		if _, exists := studentIDs[submission.StudentID]; !exists {
			return fmt.Errorf("submission references unknown student %s", submission.StudentID)
		}
		if _, exists := assignmentIDs[submission.AssignmentID]; !exists {
			return fmt.Errorf("submission references unknown assignment %s", submission.AssignmentID)
		}
		key := domain.SubmissionKey(submission.StudentID, submission.AssignmentID)
		if _, exists := submissionKeys[key]; exists {
			return fmt.Errorf("duplicate submission key %s", key)
		}
		submissionKeys[key] = struct{}{}
	}
	snapshotIDs := make(map[string]struct{}, len(s.Snapshots))
	for _, snapshot := range s.Snapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
		if _, exists := snapshotIDs[snapshot.ID]; exists {
			return fmt.Errorf("duplicate snapshot id %s", snapshot.ID)
		}
		snapshotIDs[snapshot.ID] = struct{}{}
	}
	return nil
}

func (r *Repository) ExportState() (State, error) {
	students, err := r.Students()
	if err != nil {
		return State{}, err
	}
	assignments, err := r.Assignments()
	if err != nil {
		return State{}, err
	}
	submissions, err := r.Submissions()
	if err != nil {
		return State{}, err
	}
	snapshots, err := r.Snapshots()
	if err != nil {
		return State{}, err
	}
	return State{
		Students:    students,
		Assignments: assignments,
		Submissions: submissions,
		Snapshots:   snapshots,
	}, nil
}

func (r *Repository) ReplaceState(state State) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("validate replacement state: %w", err)
	}
	return r.database.update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketStudents, bucketAssignments, bucketSubmissions, bucketSnapshots} {
			bucket := tx.Bucket(name)
			keys := make([][]byte, 0)
			if err := bucket.ForEach(func(key, _ []byte) error {
				keys = append(keys, append([]byte(nil), key...))
				return nil
			}); err != nil {
				return err
			}
			for _, key := range keys {
				if err := bucket.Delete(key); err != nil {
					return err
				}
			}
		}
		for _, student := range state.Students {
			data, err := encodeStudent(student)
			if err != nil {
				return err
			}
			if err := tx.Bucket(bucketStudents).Put([]byte(student.ID), data); err != nil {
				return err
			}
		}
		for _, assignment := range state.Assignments {
			data, err := encodeAssignment(assignment)
			if err != nil {
				return err
			}
			if err := tx.Bucket(bucketAssignments).Put([]byte(assignment.ID), data); err != nil {
				return err
			}
		}
		for _, submission := range state.Submissions {
			data, err := encodeSubmission(submission)
			if err != nil {
				return err
			}
			key := domain.SubmissionKey(submission.StudentID, submission.AssignmentID)
			if err := tx.Bucket(bucketSubmissions).Put([]byte(key), data); err != nil {
				return err
			}
		}
		for _, snapshot := range state.Snapshots {
			data, err := encodeSnapshot(snapshot)
			if err != nil {
				return err
			}
			if err := tx.Bucket(bucketSnapshots).Put([]byte(snapshot.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) Counts() (Counts, error) {
	result := Counts{}
	err := r.database.view(func(tx *bolt.Tx) error {
		result.Students = tx.Bucket(bucketStudents).Stats().KeyN
		result.Assignments = tx.Bucket(bucketAssignments).Stats().KeyN
		result.Submissions = tx.Bucket(bucketSubmissions).Stats().KeyN
		result.Snapshots = tx.Bucket(bucketSnapshots).Stats().KeyN
		return nil
	})
	return result, err
}

func (r *Repository) BuildSubmissionList() (*domain.SubmissionList, error) {
	state, err := r.ExportState()
	if err != nil {
		return nil, err
	}
	studentByID := make(map[string]*domain.Student, len(state.Students))
	for index := range state.Students {
		student := state.Students[index]
		studentByID[student.ID] = &student
	}
	assignmentByID := make(map[string]*domain.Assignment, len(state.Assignments))
	for index := range state.Assignments {
		assignment := state.Assignments[index]
		assignmentByID[assignment.ID] = &assignment
	}
	sort.SliceStable(state.Submissions, func(i, j int) bool {
		if state.Submissions[i].StudentID != state.Submissions[j].StudentID {
			return state.Submissions[i].StudentID < state.Submissions[j].StudentID
		}
		return state.Submissions[i].AssignmentID < state.Submissions[j].AssignmentID
	})
	list := domain.NewSubmissionList()
	for index := range state.Submissions {
		submission := state.Submissions[index]
		student := studentByID[submission.StudentID]
		assignment := assignmentByID[submission.AssignmentID]
		record, err := domain.NewSubmissionRecord(student, assignment, &submission)
		if err != nil {
			return nil, fmt.Errorf("restore submission %s: %w", submission.ID, err)
		}
		if err := list.Append(record); err != nil {
			return nil, err
		}
	}
	return list, nil
}
