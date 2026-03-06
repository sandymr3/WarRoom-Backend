package models

import "time"

// PhaseScenario is the scenario question injected between phases.
// It is asked by one of the participant's chosen leaders.
type PhaseScenario struct {
	ID               string     `gorm:"primaryKey;type:varchar(191)" json:"id"`
	AssessmentID     string     `gorm:"column:assessment_id;index;not null;type:varchar(191)" json:"assessmentId"`
	FromStage        string     `gorm:"column:from_stage;not null" json:"fromStage"`
	ToStage          string     `gorm:"column:to_stage;not null" json:"toStage"`
	LeaderID         string     `gorm:"column:leader_id;not null" json:"leaderId"`
	LeaderName       string     `gorm:"column:leader_name;not null" json:"leaderName"`
	ScenarioTitle    string     `gorm:"column:scenario_title;not null" json:"scenarioTitle"`
	ScenarioSetup    string     `gorm:"column:scenario_setup;type:text;not null" json:"scenarioSetup"`
	LeaderPrompt     string     `gorm:"column:leader_prompt;type:text;not null" json:"leaderPrompt"` // personalized with leader name
	UserResponse     string     `gorm:"column:user_response;type:text" json:"userResponse"`
	ProficiencyScore *int       `gorm:"column:proficiency_score" json:"proficiencyScore"` // 1/2/3
	AIFeedback       string     `gorm:"column:ai_feedback;type:text" json:"aiFeedback"`
	AnsweredAt       *time.Time `gorm:"column:answered_at" json:"answeredAt"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
}
