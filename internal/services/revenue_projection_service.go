package services

import (
	"encoding/json"
	"math"
)

// RevenueProjectionService computes a simulated projected ARR for the leaderboard.
// Inputs are the assessment's simulation state + latest competency scores.
// This is NOT a real financial model — it's a simulation metric for engagement.
type RevenueProjectionService struct{}

func NewRevenueProjectionService() *RevenueProjectionService {
	return &RevenueProjectionService{}
}

type financialState struct {
	Capital  float64 `json:"capital"`
	Revenue  float64 `json:"revenue"`
	BurnRate float64 `json:"burnRate"`
	Runway   float64 `json:"runway"`
	Equity   float64 `json:"equity"`
}

type customerState struct {
	Count        int     `json:"count"`
	Retention    float64 `json:"retention"`
	Satisfaction float64 `json:"satisfaction"`
}

// ComputeRevenueProjection returns a projected ARR (in currency units as int64).
// stageIndex: 0–8 (0=ideation … 8=warroom)
// c4Score, c5Score: weighted average competency scores (1.0–3.0)
// financialStateJSON / customerStateJSON: raw JSON blobs from Assessment
func (s *RevenueProjectionService) ComputeRevenueProjection(
	stageIndex int,
	c4Score, c5Score float64,
	financialStateJSON, customerStateJSON json.RawMessage,
) int64 {
	var fin financialState
	var cust customerState

	if len(financialStateJSON) > 0 {
		json.Unmarshal(financialStateJSON, &fin)
	}
	if len(customerStateJSON) > 0 {
		json.Unmarshal(customerStateJSON, &cust)
	}

	// Base: simulate a small SaaS / service business
	// Price per customer assumed at ₹5,000/yr (adjustable)
	const pricePerCustomer = 5000.0

	baseRevenue := float64(cust.Count) * pricePerCustomer

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

	// Competency multiplier: C4 (financial discipline) and C5 (strategic thinking)
	// Each scored 1–3; combined normalized to 0.5–1.5 multiplier
	competencyMul := 0.5 + ((c4Score-1.0)/(3.0-1.0))*0.5 + ((c5Score-1.0)/(3.0-1.0))*0.5

	// Retention bonus: high retention = compounding revenue
	retentionBonus := 1.0 + (cust.Retention/100.0)*0.5

	projected := baseRevenue * stageMul * competencyMul * retentionBonus

	// Floor: even with 0 customers show some traction signal based on stage
	floor := stageMul * 100_000 // e.g. ₹1L at stage 1 baseline
	if projected < floor {
		projected = floor
	}

	// Cap at a reasonable simulation ceiling
	projected = math.Min(projected, 500_000_000) // ₹50Cr cap

	return int64(math.Round(projected))
}
