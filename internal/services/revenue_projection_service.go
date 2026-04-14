package services

import (
	"math"
)

// RevenueProjectionService computes a simulated projected ARR for the leaderboard.
// The projection is personalized based on each user's competency scores and
// average proficiency across all their responses.
type RevenueProjectionService struct{}

func NewRevenueProjectionService() *RevenueProjectionService {
	return &RevenueProjectionService{}
}

// ComputeRevenueProjection returns a projected ARR (in currency units as int64).
//
// stageIndex: 0–8 (0=ideation … 8=warroom)
// allCompScores: map of competency code → weighted average score (1.0–3.0) across ALL completed stages
// avgProficiency: average proficiency score across all responses (1.0–3.0)
// totalResponses: how many questions the user has answered so far
func (s *RevenueProjectionService) ComputeRevenueProjection(
	stageIndex int,
	allCompScores map[string]float64,
	avgProficiency float64,
	totalResponses int,
) int64 {
	// Revenue should only start after Phase 0 (Commitment) completes.
	// When Phase 0 completes, stageIndex is 0, so it calculates the initial revenue for Phase 1.
	if stageIndex < 0 {
		return 0
	}

	// Stage growth multiplier — revenue potential increases as you progress
	stageMultipliers := []float64{
		1.0,  // Ideation
		1.2,  // Vision
		1.5,  // Commitment
		2.0,  // Validation
		3.5,  // Growth A
		5.0,  // Growth B / Expansion
		7.0,  // Scale
		9.0,  // War Room Prep
		12.0, // War Room
	}
	idx := stageIndex
	if idx < 0 {
		idx = 0
	}
	if idx >= len(stageMultipliers) {
		idx = len(stageMultipliers) - 1
	}
	stageMul := stageMultipliers[idx]

	// Overall competency multiplier from ALL 8 competencies (not just C4/C5).
	// Each competency is scored 1–3.  Average across those that exist;
	// missing competencies are treated as the neutral baseline (2.0).
	compSum := 0.0
	compCount := 0
	for _, code := range []string{"C1", "C2", "C3", "C4", "C5", "C6", "C7", "C8"} {
		if score, ok := allCompScores[code]; ok && score > 0 {
			compSum += score
			compCount++
		}
	}
	avgComp := 2.0 // neutral baseline
	if compCount > 0 {
		avgComp = compSum / float64(compCount)
	}

	// Map avgComp (1.0–3.0) to a multiplier range (0.4–1.6).
	// Score 1.0 → 0.4, Score 2.0 → 1.0, Score 3.0 → 1.6
	competencyMul := 0.4 + (avgComp-1.0)*0.6

	// Proficiency multiplier from the user's average answer quality.
	// avgProficiency (1–3) mapped to 0.5–1.5
	if avgProficiency < 1.0 {
		avgProficiency = 2.0 // default when no responses yet
	}
	proficiencyMul := 0.5 + (avgProficiency-1.0)*0.5

	// Engagement multiplier: more questions answered = better traction signal.
	// Diminishing returns via log; 10 answers ≈ 1.0x, 30 ≈ 1.15x, 60 ≈ 1.25x
	engagementMul := 1.0
	if totalResponses > 0 {
		engagementMul = 1.0 + 0.15*math.Log10(float64(totalResponses))
	}

	// Base revenue tied to stage (replaces the old customer-count approach
	// which was always 0). This is the "potential market" unlocked at each stage.
	baseRevenue := stageMul * 100_000

	projected := baseRevenue * competencyMul * proficiencyMul * engagementMul

	// Cap at a reasonable simulation ceiling
	projected = math.Min(projected, 500_000_000) // ₹50Cr cap

	return int64(math.Round(projected))
}
