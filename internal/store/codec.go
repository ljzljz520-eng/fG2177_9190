package store

import (
	"encoding/json"
	"fmt"

	"courseworkledger/internal/domain"
)

func encodeStudent(value domain.Student) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, fmt.Errorf("validate student: %w", err)
	}
	return encodeJSON("student", value)
}

func decodeStudent(data []byte) (domain.Student, error) {
	var value domain.Student
	if err := decodeJSON("student", data, &value); err != nil {
		return domain.Student{}, err
	}
	if err := value.Validate(); err != nil {
		return domain.Student{}, fmt.Errorf("validate stored student: %w", err)
	}
	return value, nil
}

func encodeAssignment(value domain.Assignment) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, fmt.Errorf("validate assignment: %w", err)
	}
	return encodeJSON("assignment", value)
}

func decodeAssignment(data []byte) (domain.Assignment, error) {
	var value domain.Assignment
	if err := decodeJSON("assignment", data, &value); err != nil {
		return domain.Assignment{}, err
	}
	if err := value.Validate(); err != nil {
		return domain.Assignment{}, fmt.Errorf("validate stored assignment: %w", err)
	}
	return value, nil
}

func encodeSubmission(value domain.Submission) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, fmt.Errorf("validate submission: %w", err)
	}
	return encodeJSON("submission", value)
}

func decodeSubmission(data []byte) (domain.Submission, error) {
	var value domain.Submission
	if err := decodeJSON("submission", data, &value); err != nil {
		return domain.Submission{}, err
	}
	if err := value.Validate(); err != nil {
		return domain.Submission{}, fmt.Errorf("validate stored submission: %w", err)
	}
	return value, nil
}

func encodeSnapshot(value domain.GradeSnapshot) ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, fmt.Errorf("validate grade snapshot: %w", err)
	}
	return encodeJSON("grade snapshot", value)
}

func decodeSnapshot(data []byte) (domain.GradeSnapshot, error) {
	var value domain.GradeSnapshot
	if err := decodeJSON("grade snapshot", data, &value); err != nil {
		return domain.GradeSnapshot{}, err
	}
	if err := value.Validate(); err != nil {
		return domain.GradeSnapshot{}, fmt.Errorf("validate stored grade snapshot: %w", err)
	}
	return value, nil
}

func encodeJSON(kind string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", kind, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("encode %s: empty result", kind)
	}
	return data, nil
}

func decodeJSON(kind string, data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("decode %s: empty data", kind)
	}
	if target == nil {
		return fmt.Errorf("decode %s: nil target", kind)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", kind, err)
	}
	return nil
}
