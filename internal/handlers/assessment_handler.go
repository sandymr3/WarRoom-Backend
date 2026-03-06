package handlers

import (
	"encoding/json"
	"net/http"
	"war-room-backend/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AssessmentHandler struct {
	Service *services.AssessmentService
}

func NewAssessmentHandler(s *services.AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{Service: s}
}

// ============================================
// Request types
// ============================================

type SubmitResponseRequest struct {
	QuestionID   string          `json:"questionId"`
	ResponseData json.RawMessage `json:"responseData"`
}

type SubmitStageResponsesRequest struct {
	Responses map[string]json.RawMessage `json:"responses"`
}

// PhaseResponseItem is a single answer from the frontend array format.
type PhaseResponseItem struct {
	QuestionID       string         `json:"questionId"`
	Type             string         `json:"type"`
	Text             string         `json:"text,omitempty"`
	SelectedOptionID string         `json:"selectedOptionId,omitempty"`
	Allocations      map[string]int `json:"allocations,omitempty"`
}

// PhaseSubmitRequest collects all answers for a full phase.
type PhaseSubmitRequest struct {
	StageID   string              `json:"stageId"`
	Responses []PhaseResponseItem `json:"responses"` // array of response items from frontend
}

// CharactersRequest for setting chosen mentors/leaders/investors.
type CharactersRequest struct {
	SelectedMentors   []string `json:"selectedMentors"`   // 3 mentor IDs
	SelectedLeaders   []string `json:"selectedLeaders"`   // 3 leader IDs
	SelectedInvestors []string `json:"selectedInvestors"` // 3 investor IDs
}

// PhaseScenarioRequest for submitting the leader scenario answer.
type PhaseScenarioRequest struct {
	FromStage string `json:"fromStage"`
	ToStage   string `json:"toStage"`
	Response  string `json:"response"`
}

type MentorLifelineRequest struct {
	MentorID string `json:"mentorId"`
	Question string `json:"question"`
}

type SubmitPitchRequest struct {
	PitchText string `json:"pitchText"`
}

type InvestorResponseRequest struct {
	InvestorID string `json:"investorId"`
	Response   string `json:"response"`
}

// ============================================
// Helper: extract user ID from JWT
// ============================================

func getUserID(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["user_id"].(string)
}

// ============================================
// CRUD Endpoints
// ============================================

// POST /assessments - Create new assessment
func (h *AssessmentHandler) Create(c echo.Context) error {
	userID := getUserID(c)

	var req json.RawMessage
	if err := c.Bind(&req); err != nil {
		req = json.RawMessage(`{}`)
	}

	assessment, err := h.Service.CreateAssessment(userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create assessment"})
	}

	return c.JSON(http.StatusCreated, assessment)
}

// GET /assessments - List user's assessments
func (h *AssessmentHandler) List(c echo.Context) error {
	userID := getUserID(c)

	assessments, err := h.Service.ListAssessments(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list assessments"})
	}
	return c.JSON(http.StatusOK, assessments)
}

// GET /assessments/:id - Get assessment state
func (h *AssessmentHandler) Get(c echo.Context) error {
	assessmentID := c.Param("id")

	state, err := h.Service.GetAssessment(assessmentID)
	if err != nil {
		if err.Error() == "assessment not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Assessment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get assessment"})
	}

	return c.JSON(http.StatusOK, state)
}

// ============================================
// Response Endpoints
// ============================================

// POST /assessments/:id/responses - Submit response to current question
func (h *AssessmentHandler) SubmitResponse(c echo.Context) error {
	assessmentID := c.Param("id")

	req := new(SubmitResponseRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := h.Service.SubmitResponse(assessmentID, req.QuestionID, req.ResponseData)
	if err != nil {
		switch err.Error() {
		case "assessment not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case "invalid question ID":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit response"})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// POST /assessments/:id/stage-responses - Submit responses for the entire stage
func (h *AssessmentHandler) SubmitStageResponses(c echo.Context) error {
	assessmentID := c.Param("id")

	req := new(SubmitStageResponsesRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := h.Service.SubmitStageResponses(assessmentID, req.Responses)
	if err != nil {
		switch err.Error() {
		case "assessment not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case "invalid stage ID":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit stage responses"})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// ============================================
// Mentor Lifeline
// ============================================

// POST /assessments/:id/mentor - Use mentor lifeline
func (h *AssessmentHandler) UseMentorLifeline(c echo.Context) error {
	assessmentID := c.Param("id")

	req := new(MentorLifelineRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := h.Service.UseMentorLifeline(assessmentID, req.MentorID, req.Question)
	if err != nil {
		switch err.Error() {
		case "assessment not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case "no mentor lifelines remaining":
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		case "invalid mentor ID":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to use mentor lifeline"})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// ============================================
// War Room Endpoints
// ============================================

// POST /assessments/:id/warroom/pitch - Submit War Room pitch
func (h *AssessmentHandler) SubmitPitch(c echo.Context) error {
	assessmentID := c.Param("id")

	req := new(SubmitPitchRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := h.Service.SubmitPitch(assessmentID, req.PitchText)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit pitch"})
	}

	return c.JSON(http.StatusOK, result)
}

// POST /assessments/:id/warroom/respond - Respond to investor question
func (h *AssessmentHandler) RespondToInvestor(c echo.Context) error {
	assessmentID := c.Param("id")

	req := new(InvestorResponseRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	scorecard, err := h.Service.RespondToInvestor(assessmentID, req.InvestorID, req.Response)
	if err != nil {
		switch err.Error() {
		case "assessment not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case "invalid investor ID":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process investor response"})
		}
	}

	return c.JSON(http.StatusOK, scorecard)
}

// GET /assessments/:id/warroom/scorecard - Get investor scorecards
func (h *AssessmentHandler) GetScorecard(c echo.Context) error {
	assessmentID := c.Param("id")

	scorecards, err := h.Service.GetScorecards(assessmentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get scorecard"})
	}

	return c.JSON(http.StatusOK, scorecards)
}

// ============================================
// Report
// ============================================

// GET /assessments/:id/report - Generate or get evaluation report
func (h *AssessmentHandler) GetReport(c echo.Context) error {
	assessmentID := c.Param("id")

	report, err := h.Service.GenerateReport(assessmentID)
	if err != nil {
		if err.Error() == "assessment not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate report"})
	}

	return c.JSON(http.StatusOK, report)
}

// ============================================
// Phase Submit (new v2 flow)
// ============================================

// POST /assessments/:id/phase-submit
// Accepts all answers for a phase, auto-scores MCQ, queues AI for open text.
func (h *AssessmentHandler) SubmitPhase(c echo.Context) error {
	assessmentID := c.Param("id")
	req := new(PhaseSubmitRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request: " + err.Error()})
	}

	// Convert array of responses into map[questionId]json.RawMessage for the service
	responsesMap := make(map[string]json.RawMessage)
	for _, r := range req.Responses {
		raw, _ := json.Marshal(r)
		responsesMap[r.QuestionID] = raw
	}

	result, err := h.Service.SubmitPhase(assessmentID, req.StageID, responsesMap)
	if err != nil {
		switch err.Error() {
		case "assessment not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case "stage mismatch":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit phase: " + err.Error()})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// ============================================
// Character Selection
// ============================================

// GET /assessments/:id/characters
func (h *AssessmentHandler) GetCharacters(c echo.Context) error {
	assessmentID := c.Param("id")
	chars, err := h.Service.GetCharacters(assessmentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, chars)
}

// POST /assessments/:id/characters
func (h *AssessmentHandler) SetCharacters(c echo.Context) error {
	assessmentID := c.Param("id")
	req := new(CharactersRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if len(req.SelectedMentors) != 3 || len(req.SelectedLeaders) != 3 || len(req.SelectedInvestors) != 3 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Must select exactly 3 mentors, 3 leaders, and 3 investors"})
	}
	if err := h.Service.SetCharacters(assessmentID, req.SelectedMentors, req.SelectedLeaders, req.SelectedInvestors); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Characters saved"})
}

// ============================================
// Phase Scenario
// ============================================

// POST /assessments/:id/phase-scenario
func (h *AssessmentHandler) AnswerPhaseScenario(c echo.Context) error {
	assessmentID := c.Param("id")
	req := new(PhaseScenarioRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := h.Service.AnswerPhaseScenario(assessmentID, req.FromStage, req.ToStage, req.Response)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}
