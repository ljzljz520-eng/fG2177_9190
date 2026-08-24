package policy

import (
	"fmt"
	"sort"

	"courseworkledger/internal/domain"
)

type GradeBand struct {
	Name       string
	MinimumPct int
	Message    string
}

type ScoreDecision struct {
	Score      int
	Maximum    int
	Percentage int
	Band       GradeBand
	Passing    bool
}

type ScoringPolicy struct {
	passingPercentage int
	bands             []GradeBand
}

func NewScoringPolicy(passingPercentage int, bands []GradeBand) (ScoringPolicy, error) {
	if passingPercentage < 0 || passingPercentage > 100 {
		return ScoringPolicy{}, fmt.Errorf("passing percentage must be between 0 and 100")
	}
	if len(bands) == 0 {
		bands = DefaultGradeBands()
	}
	normalized := append([]GradeBand(nil), bands...)
	for _, band := range normalized {
		if band.Name == "" {
			return ScoringPolicy{}, fmt.Errorf("grade band name is required")
		}
		if band.MinimumPct < 0 || band.MinimumPct > 100 {
			return ScoringPolicy{}, fmt.Errorf("grade band minimum must be between 0 and 100")
		}
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].MinimumPct > normalized[j].MinimumPct
	})
	for index := 1; index < len(normalized); index++ {
		if normalized[index-1].MinimumPct == normalized[index].MinimumPct {
			return ScoringPolicy{}, fmt.Errorf("grade bands must have unique minimums")
		}
	}
	return ScoringPolicy{passingPercentage: passingPercentage, bands: normalized}, nil
}

func DefaultGradeBands() []GradeBand {
	return []GradeBand{
		{Name: "excellent", MinimumPct: 90, Message: "exceeds expectations"},
		{Name: "strong", MinimumPct: 80, Message: "meets expectations well"},
		{Name: "satisfactory", MinimumPct: 70, Message: "meets expectations"},
		{Name: "developing", MinimumPct: 60, Message: "requires targeted revision"},
		{Name: "incomplete", MinimumPct: 0, Message: "requires substantial revision"},
	}
}

func (p ScoringPolicy) Evaluate(score, maximum int) (ScoreDecision, error) {
	if maximum <= 0 {
		return ScoreDecision{}, fmt.Errorf("maximum score must be positive")
	}
	if score < 0 || score > maximum {
		return ScoreDecision{}, fmt.Errorf("score %d is outside range 0..%d", score, maximum)
	}
	percentage := (score * 100) / maximum
	band, ok := p.BandForPercentage(percentage)
	if !ok {
		return ScoreDecision{}, fmt.Errorf("no grade band covers %d percent", percentage)
	}
	return ScoreDecision{
		Score:      score,
		Maximum:    maximum,
		Percentage: percentage,
		Band:       band,
		Passing:    percentage >= p.passingPercentage,
	}, nil
}

func (p ScoringPolicy) BandForPercentage(percentage int) (GradeBand, bool) {
	if percentage < 0 || percentage > 100 {
		return GradeBand{}, false
	}
	for _, band := range p.bands {
		if percentage >= band.MinimumPct {
			return band, true
		}
	}
	return GradeBand{}, false
}

func (p ScoringPolicy) Grade(submission domain.Submission, assignment domain.Assignment, score int, feedback string) (domain.Submission, ScoreDecision, error) {
	if !submission.IsSubmitted() {
		return domain.Submission{}, ScoreDecision{}, fmt.Errorf("submission is not ready for grading")
	}
	decision, err := p.Evaluate(score, assignment.MaximumScore)
	if err != nil {
		return domain.Submission{}, ScoreDecision{}, err
	}
	status := domain.StatusGraded
	patch := domain.SubmissionPatch{Status: &status, Score: &score, Feedback: &feedback}
	updated, err := submission.Apply(patch)
	if err != nil {
		return domain.Submission{}, ScoreDecision{}, err
	}
	return updated, decision, nil
}

func (p ScoringPolicy) PassingPercentage() int {
	return p.passingPercentage
}

func (p ScoringPolicy) Bands() []GradeBand {
	return append([]GradeBand(nil), p.bands...)
}
