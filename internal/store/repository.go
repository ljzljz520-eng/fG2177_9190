package store

import (
	"fmt"
	"sort"

	"courseworkledger/internal/domain"
	bolt "go.etcd.io/bbolt"
)

type Repository struct {
	database *Database
}

func NewRepository(database *Database) (*Repository, error) {
	if database == nil {
		return nil, fmt.Errorf("database is required")
	}
	if database.IsClosed() {
		return nil, fmt.Errorf("database is closed")
	}
	return &Repository{database: database}, nil
}

func (r *Repository) Database() *Database {
	if r == nil {
		return nil
	}
	return r.database
}

func (r *Repository) SaveStudent(student domain.Student) error {
	data, err := encodeStudent(student)
	if err != nil {
		return err
	}
	return r.database.update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketStudents)
		if existing := bucket.Get([]byte(student.ID)); existing != nil {
			stored, decodeErr := decodeStudent(existing)
			if decodeErr != nil {
				return decodeErr
			}
			if stored.StudentNumber != student.StudentNumber {
				return fmt.Errorf("student number cannot be changed")
			}
		}
		if conflict, conflictID := findStudentNumber(bucket, student.StudentNumber, student.ID); conflict {
			return fmt.Errorf("student number %q belongs to %s", student.StudentNumber, conflictID)
		}
		return bucket.Put([]byte(student.ID), data)
	})
}

func (r *Repository) Student(id string) (domain.Student, bool, error) {
	var result domain.Student
	var found bool
	err := r.database.view(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketStudents).Get([]byte(id))
		if data == nil {
			return nil
		}
		value, err := decodeStudent(data)
		if err != nil {
			return err
		}
		result = value
		found = true
		return nil
	})
	return result, found, err
}

func (r *Repository) StudentByNumber(number string) (domain.Student, bool, error) {
	var result domain.Student
	var found bool
	err := r.database.view(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketStudents).ForEach(func(_, data []byte) error {
			student, err := decodeStudent(data)
			if err != nil {
				return err
			}
			if student.StudentNumber == number {
				result = student
				found = true
			}
			return nil
		})
	})
	return result, found, err
}

func (r *Repository) Students() ([]domain.Student, error) {
	result := make([]domain.Student, 0)
	err := r.database.view(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketStudents).ForEach(func(_, data []byte) error {
			student, err := decodeStudent(data)
			if err != nil {
				return err
			}
			result = append(result, student)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return domain.SortStudents(result), nil
}

func (r *Repository) DeleteStudent(id string) (bool, error) {
	deleted := false
	err := r.database.update(func(tx *bolt.Tx) error {
		students := tx.Bucket(bucketStudents)
		if students.Get([]byte(id)) == nil {
			return nil
		}
		if hasSubmissionForStudent(tx.Bucket(bucketSubmissions), id) {
			return fmt.Errorf("student %s has submissions", id)
		}
		if err := students.Delete([]byte(id)); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *Repository) SaveAssignment(assignment domain.Assignment) error {
	data, err := encodeAssignment(assignment)
	if err != nil {
		return err
	}
	return r.database.update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketAssignments).Put([]byte(assignment.ID), data)
	})
}

func (r *Repository) Assignment(id string) (domain.Assignment, bool, error) {
	var result domain.Assignment
	var found bool
	err := r.database.view(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketAssignments).Get([]byte(id))
		if data == nil {
			return nil
		}
		value, err := decodeAssignment(data)
		if err != nil {
			return err
		}
		result = value
		found = true
		return nil
	})
	return result, found, err
}

func (r *Repository) Assignments() ([]domain.Assignment, error) {
	result := make([]domain.Assignment, 0)
	err := r.database.view(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketAssignments).ForEach(func(_, data []byte) error {
			assignment, err := decodeAssignment(data)
			if err != nil {
				return err
			}
			result = append(result, assignment)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return domain.SortAssignments(result), nil
}

func (r *Repository) DeleteAssignment(id string) (bool, error) {
	deleted := false
	err := r.database.update(func(tx *bolt.Tx) error {
		assignments := tx.Bucket(bucketAssignments)
		if assignments.Get([]byte(id)) == nil {
			return nil
		}
		if hasSubmissionForAssignment(tx.Bucket(bucketSubmissions), id) {
			return fmt.Errorf("assignment %s has submissions", id)
		}
		if err := assignments.Delete([]byte(id)); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *Repository) SaveSubmission(submission domain.Submission) error {
	data, err := encodeSubmission(submission)
	if err != nil {
		return err
	}
	return r.database.update(func(tx *bolt.Tx) error {
		if tx.Bucket(bucketStudents).Get([]byte(submission.StudentID)) == nil {
			return fmt.Errorf("student %s does not exist", submission.StudentID)
		}
		if tx.Bucket(bucketAssignments).Get([]byte(submission.AssignmentID)) == nil {
			return fmt.Errorf("assignment %s does not exist", submission.AssignmentID)
		}
		key := domain.SubmissionKey(submission.StudentID, submission.AssignmentID)
		return tx.Bucket(bucketSubmissions).Put([]byte(key), data)
	})
}

func (r *Repository) Submission(studentID, assignmentID string) (domain.Submission, bool, error) {
	var result domain.Submission
	var found bool
	err := r.database.view(func(tx *bolt.Tx) error {
		key := domain.SubmissionKey(studentID, assignmentID)
		data := tx.Bucket(bucketSubmissions).Get([]byte(key))
		if data == nil {
			return nil
		}
		value, err := decodeSubmission(data)
		if err != nil {
			return err
		}
		result = value
		found = true
		return nil
	})
	return result, found, err
}

func (r *Repository) Submissions() ([]domain.Submission, error) {
	result := make([]domain.Submission, 0)
	err := r.database.view(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSubmissions).ForEach(func(_, data []byte) error {
			submission, err := decodeSubmission(data)
			if err != nil {
				return err
			}
			result = append(result, submission)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].StudentID != result[j].StudentID {
			return result[i].StudentID < result[j].StudentID
		}
		return result[i].AssignmentID < result[j].AssignmentID
	})
	return result, nil
}

func (r *Repository) DeleteSubmission(studentID, assignmentID string) (bool, error) {
	deleted := false
	err := r.database.update(func(tx *bolt.Tx) error {
		key := []byte(domain.SubmissionKey(studentID, assignmentID))
		bucket := tx.Bucket(bucketSubmissions)
		if bucket.Get(key) == nil {
			return nil
		}
		if err := bucket.Delete(key); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *Repository) SaveSnapshot(snapshot domain.GradeSnapshot) error {
	data, err := encodeSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.database.update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSnapshots).Put([]byte(snapshot.ID), data)
	})
}

func (r *Repository) Snapshot(id string) (domain.GradeSnapshot, bool, error) {
	var result domain.GradeSnapshot
	var found bool
	err := r.database.view(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketSnapshots).Get([]byte(id))
		if data == nil {
			return nil
		}
		value, err := decodeSnapshot(data)
		if err != nil {
			return err
		}
		result = value
		found = true
		return nil
	})
	return result, found, err
}

func (r *Repository) Snapshots() ([]domain.GradeSnapshot, error) {
	result := make([]domain.GradeSnapshot, 0)
	err := r.database.view(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSnapshots).ForEach(func(_, data []byte) error {
			snapshot, err := decodeSnapshot(data)
			if err != nil {
				return err
			}
			result = append(result, snapshot)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].StudentNumber != result[j].StudentNumber {
			return result[i].StudentNumber < result[j].StudentNumber
		}
		if result[i].AssignmentName != result[j].AssignmentName {
			return result[i].AssignmentName < result[j].AssignmentName
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (r *Repository) DeleteSnapshot(id string) (bool, error) {
	deleted := false
	err := r.database.update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketSnapshots)
		if bucket.Get([]byte(id)) == nil {
			return nil
		}
		if err := bucket.Delete([]byte(id)); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func findStudentNumber(bucket *bolt.Bucket, number, exceptID string) (bool, string) {
	var conflict bool
	var conflictID string
	_ = bucket.ForEach(func(key, data []byte) error {
		student, err := decodeStudent(data)
		if err != nil {
			return nil
		}
		if student.StudentNumber == number && student.ID != exceptID {
			conflict = true
			conflictID = string(key)
		}
		return nil
	})
	return conflict, conflictID
}

func hasSubmissionForStudent(bucket *bolt.Bucket, studentID string) bool {
	result := false
	_ = bucket.ForEach(func(key, _ []byte) error {
		student, _, ok := domain.SplitSubmissionKey(string(key))
		if ok && student == studentID {
			result = true
		}
		return nil
	})
	return result
}

func hasSubmissionForAssignment(bucket *bolt.Bucket, assignmentID string) bool {
	result := false
	_ = bucket.ForEach(func(key, _ []byte) error {
		_, assignment, ok := domain.SplitSubmissionKey(string(key))
		if ok && assignment == assignmentID {
			result = true
		}
		return nil
	})
	return result
}
