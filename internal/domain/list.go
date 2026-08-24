package domain

import "fmt"

type submissionNode struct {
	record *SubmissionRecord
	next   *submissionNode
	prev   *submissionNode
}

type SubmissionList struct {
	head  *submissionNode
	tail  *submissionNode
	index map[string]*submissionNode
	len   int
}

func NewSubmissionList() *SubmissionList {
	return &SubmissionList{index: make(map[string]*submissionNode)}
}

func (l *SubmissionList) Len() int {
	if l == nil {
		return 0
	}
	return l.len
}

func (l *SubmissionList) Append(record *SubmissionRecord) error {
	if l == nil {
		return fmt.Errorf("submission list is nil")
	}
	if record == nil || record.Student == nil || record.Assignment == nil || record.Submission == nil {
		return fmt.Errorf("complete submission record is required")
	}
	key := record.Key()
	if _, exists := l.index[key]; exists {
		return fmt.Errorf("submission record %q already exists", key)
	}
	node := &submissionNode{record: record, prev: l.tail}
	if l.tail == nil {
		l.head = node
		l.tail = node
	} else {
		l.tail.next = node
		l.tail = node
	}
	l.index[key] = node
	l.len++
	return nil
}

func (l *SubmissionList) Upsert(record *SubmissionRecord) error {
	if l == nil {
		return fmt.Errorf("submission list is nil")
	}
	if record == nil || record.Student == nil || record.Assignment == nil || record.Submission == nil {
		return fmt.Errorf("complete submission record is required")
	}
	key := record.Key()
	if node, exists := l.index[key]; exists {
		node.record = record
		return nil
	}
	return l.Append(record)
}

func (l *SubmissionList) FindByStudentAndAssignment(studentID, assignmentID string) *SubmissionRecord {
	if l == nil {
		return nil
	}
	node := l.index[SubmissionKey(studentID, assignmentID)]
	if node == nil {
		return nil
	}
	return node.record
}

func (l *SubmissionList) FindBySubmissionID(id string) (*SubmissionRecord, bool) {
	if l == nil {
		return nil, false
	}
	for node := l.head; node != nil; node = node.next {
		if node.record != nil && node.record.Submission != nil && node.record.Submission.ID == id {
			return node.record, true
		}
	}
	return nil, false
}

func (l *SubmissionList) Delete(studentID, assignmentID string) (*SubmissionRecord, bool) {
	if l == nil {
		return nil, false
	}
	key := SubmissionKey(studentID, assignmentID)
	node := l.index[key]
	if node == nil {
		return nil, false
	}
	if node.prev == nil {
		l.head = node.next
	} else {
		node.prev.next = node.next
	}
	if node.next == nil {
		l.tail = node.prev
	} else {
		node.next.prev = node.prev
	}
	delete(l.index, key)
	l.len--
	return node.record, true
}

func (l *SubmissionList) MoveToEnd(studentID, assignmentID string) bool {
	if l == nil || l.tail == nil {
		return false
	}
	node := l.index[SubmissionKey(studentID, assignmentID)]
	if node == nil {
		return false
	}
	if node == l.tail {
		return true
	}
	if node.prev == nil {
		l.head = node.next
	} else {
		node.prev.next = node.next
	}
	node.next.prev = node.prev
	node.prev = l.tail
	node.next = nil
	l.tail.next = node
	l.tail = node
	return true
}

func (l *SubmissionList) Records() []*SubmissionRecord {
	if l == nil {
		return nil
	}
	result := make([]*SubmissionRecord, 0, l.len)
	for node := l.head; node != nil; node = node.next {
		result = append(result, node.record)
	}
	return result
}

func (l *SubmissionList) ReverseRecords() []*SubmissionRecord {
	if l == nil {
		return nil
	}
	result := make([]*SubmissionRecord, 0, l.len)
	for node := l.tail; node != nil; node = node.prev {
		result = append(result, node.record)
	}
	return result
}

func (l *SubmissionList) ForEach(visit func(position int, record *SubmissionRecord) error) error {
	if l == nil {
		return fmt.Errorf("submission list is nil")
	}
	if visit == nil {
		return fmt.Errorf("submission visitor is nil")
	}
	position := 0
	for node := l.head; node != nil; node = node.next {
		if err := visit(position, node.record); err != nil {
			return err
		}
		position++
	}
	return nil
}

func (l *SubmissionList) Validate() error {
	if l == nil {
		return fmt.Errorf("submission list is nil")
	}
	count := 0
	var previous *submissionNode
	for node := l.head; node != nil; node = node.next {
		if node.prev != previous {
			return fmt.Errorf("submission list previous link is inconsistent at %d", count)
		}
		if node.record == nil || node.record.Submission == nil {
			return fmt.Errorf("submission list contains incomplete record at %d", count)
		}
		if indexed := l.index[node.record.Key()]; indexed != node {
			return fmt.Errorf("submission list index is inconsistent at %d", count)
		}
		previous = node
		count++
	}
	if previous != l.tail {
		return fmt.Errorf("submission list tail is inconsistent")
	}
	if count != l.len || count != len(l.index) {
		return fmt.Errorf("submission list length is inconsistent")
	}
	return nil
}
