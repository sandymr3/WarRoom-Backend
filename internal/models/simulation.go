package models

import "encoding/json"

// ============================================
// SIMULATION CONFIG (loaded from JSON)
// ============================================

type SimulationConfig struct {
	Name      string     `json:"name"`
	Version   string     `json:"version"`
	Levels    []int      `json:"levels"` // [1, 2]
	Mentors   []Mentor   `json:"mentors"`
	Investors []Investor `json:"investors"`
	Leaders   []Leader   `json:"leaders"`

	Competencies []Competency              `json:"competencies"`
	StageWeights map[string]map[string]int `json:"stage_weights"` // stage -> competency -> weight

	Stages           []SimStage        `json:"stages"`
	PhaseTransitions []PhaseTransition `json:"phase_transitions"` // scenario questions between stages
}

// ============================================
// PERSONAS
// ============================================

type Mentor struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Gender         string `json:"gender"`
	Specialization string `json:"specialization"`
	Avatar         string `json:"avatar"`
	Bio            string `json:"bio"`
	GuidanceStyle  string `json:"guidance_style"`
	Tone           string `json:"tone"`
}

type Investor struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Gender            string   `json:"gender"`
	PrimaryLens       string   `json:"primary_lens"`
	BiasTraitName     string   `json:"bias_trait_name"`
	Avatar            string   `json:"avatar"`
	Bio               string   `json:"bio"`
	SignatureQuestion string   `json:"signature_question"`
	WalkOutTrigger    string   `json:"walk_out_trigger"`
	Tone              string   `json:"tone"`
	Characteristics   []string `json:"characteristics"`
}

type Leader struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Gender         string `json:"gender"`
	Specialization string `json:"specialization"`
	Avatar         string `json:"avatar"`
	Bio            string `json:"bio"`
	Tone           string `json:"tone"`
}

// ============================================
// COMPETENCY DEFINITION
// ============================================

type Competency struct {
	Code           string   `json:"code"` // C1-C8
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	WhatItMeasures []string `json:"what_it_measures"`
	Developing     []string `json:"developing"` // P1 indicators
	Strong         []string `json:"strong"`     // P2 indicators
	Advanced       []string `json:"advanced"`   // P3 indicators
}

// ============================================
// SIMULATION STAGE DEFINITION
// ============================================

type SimStage struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Title              string            `json:"title"`
	StageNumber        int               `json:"stage_number"` // -2, -1, 0, 1, 2, 3, 4
	Goal               string            `json:"goal"`
	ResearchBackground string            `json:"research_background,omitempty"`
	WhatHappens        []string          `json:"what_happens"`
	PressureInjected   []string          `json:"pressure_injected"`
	DurationMinutes    int               `json:"duration_minutes"`
	SimulatedMonths    []int             `json:"simulated_months"`
	Competencies       []string          `json:"competencies"`                 // ["C1", "C2"]
	ProficiencyRubric  map[string]string `json:"proficiency_rubric,omitempty"` // P1, P2, P3 → description
	Questions          []SimQuestion     `json:"questions"`
}

type SimQuestion struct {
	QID           string                 `json:"q_id"`
	Type          string                 `json:"type"` // open_text, multiple_choice, scenario, budget_allocation, etc.
	Text          string                 `json:"text"`
	ContextText   string                 `json:"context_text,omitempty"`
	PressureText  string                 `json:"pressure_text,omitempty"`
	Assess        []string               `json:"assess"` // ["C1", "C2"]
	Section       string                 `json:"section,omitempty"`
	Tag           string                 `json:"tag,omitempty"`
	ScenarioStep  string                 `json:"scenario_step,omitempty"` // environment, problem, decision, consequence
	ScenarioGroup string                 `json:"scenario_group,omitempty"`
	Options       []SimOption            `json:"options,omitempty"`
	AIEvalPrompt  string                 `json:"ai_eval_prompt,omitempty"`
	ScoringGuide  map[string]interface{} `json:"scoring_guide,omitempty"`
	Next          string                 `json:"next,omitempty"`
	FollowUp      *SimQuestion           `json:"follow_up,omitempty"`
}

type SimOption struct {
	ID          string                 `json:"id"`
	Text        string                 `json:"text"`
	Proficiency int                    `json:"proficiency,omitempty"` // 1=P1, 2=P2, 3=P3
	Feedback    string                 `json:"feedback,omitempty"`    // shown immediately for MCQ
	Signal      string                 `json:"signal,omitempty"`
	Next        string                 `json:"next,omitempty"`
	Warning     string                 `json:"warning,omitempty"`
	Impact      map[string]interface{} `json:"impact,omitempty"`
}

// ============================================
// PHASE TRANSITION SCENARIO
// ============================================

// PhaseTransition defines a scenario question injected between two stages.
// It is asked by one of the participant's chosen leaders.
type PhaseTransition struct {
	FromStage    string                  `json:"from_stage"`
	ToStage      string                  `json:"to_stage"`
	IsCheckpoint bool                    `json:"is_checkpoint,omitempty"`
	Scenario     PhaseTransitionScenario `json:"scenario"`
}

type PhaseTransitionScenario struct {
	CaseTitle            string            `json:"case_title"`
	Setup                string            `json:"setup"`
	LeaderPromptTemplate string            `json:"leader_prompt_template"`
	CompetenciesAssessed []string          `json:"competencies_assessed"`
	IsCheckpoint         bool              `json:"is_checkpoint,omitempty"`
	ProficiencyRubric    map[string]string `json:"proficiency_rubric"` // "P1", "P2", "P3" → description
}

// JSON handling helper for SimQuestion
func (q *SimQuestion) UnmarshalJSON(data []byte) error {
	type Alias SimQuestion
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(q),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
