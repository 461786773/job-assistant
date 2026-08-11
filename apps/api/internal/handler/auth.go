package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/auth"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{2,32}$`)

type enterBody struct {
	Username string `json:"username"`
}

// Login 初期版本：仅用户名，无密码。不存在则自动创建。
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	h.enterAs(w, r, http.StatusOK)
}

// Register 与 Login 相同（兼容旧前端）。
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	h.enterAs(w, r, http.StatusCreated)
}

func (h *Handler) enterAs(w http.ResponseWriter, r *http.Request, status int) {
	var body enterBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效的 JSON")
		return
	}
	username := strings.TrimSpace(body.Username)
	if username == "" {
		username = db.DefaultUsername
	}
	if !usernameRE.MatchString(username) {
		writeErr(w, http.StatusBadRequest, "用户名需 2–32 位字母、数字或下划线")
		return
	}
	user, err := h.Store.EnsureUser(username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := h.Tokens.Issue(user.ID, user.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	writeJSON(w, status, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeErr(w, http.StatusUnauthorized, "未登录")
		return
	}
	user, err := h.Store.GetUserByID(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "用户不存在")
		return
	}
	hasAssessment, err := h.Store.HasInitialAssessment(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	var latestID, latestAt, suggestedNeed, crisisLevel string
	if hasAssessment {
		if latest, err := h.Store.LatestInitialAssessment(claims.UserID); err == nil && latest != nil {
			latestID = latest.ID
			latestAt = latest.CompletedAt
			crisisLevel = latest.CrisisLevel
			var metrics struct {
				SuggestedNeed string `json:"suggestedNeed"`
			}
			_ = json.Unmarshal(latest.Metrics, &metrics)
			suggestedNeed = metrics.SuggestedNeed
			if suggestedNeed == "" {
				switch latest.PrimaryScene {
				case "job_search", "promotion", "communication":
					suggestedNeed = latest.PrimaryScene
				case "mixed":
					suggestedNeed = "unsure"
				default:
					suggestedNeed = "counsel_first"
				}
			}
		}
	}
	hasBigFive := false
	var bigFiveID, bigFiveTitle string
	if bf, err := h.Store.LatestBigFiveProfile(claims.UserID); err == nil && bf != nil {
		hasBigFive = true
		bigFiveID = bf.ID
		bigFiveTitle = bf.PersonaTitle
	}
	hasQuickSnapshot := false
	// §0.5-B / §11 #11：推荐合读画像 + 基线·最新 + 快照·最新
	if q, _ := h.Store.LatestQuickSelfCheck(claims.UserID); q != nil {
		hasQuickSnapshot = true
		// 无基线时，困扰偏高则优先建议先把心安下来
		if !hasAssessment && q.DistressScore >= 7 && suggestedNeed == "" {
			suggestedNeed = "counsel_first"
		}
		// 有基线但快照显示此刻很难受，且推荐偏「做事」时，轻轻改向疏导
		if hasAssessment && q.DistressScore >= 8 &&
			(suggestedNeed == "job_search" || suggestedNeed == "promotion") {
			suggestedNeed = "counsel_first"
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                   user.ID,
		"username":             user.Username,
		"primaryNeed":          user.PrimaryNeed,
		"hasInitialAssessment": hasAssessment,
		"latestAssessmentId":   latestID,
		"latestAssessmentAt":   latestAt,
		"suggestedNeed":        suggestedNeed,
		"crisisLevel":          crisisLevel,
		"hasBigFiveProfile":    hasBigFive,
		"latestBigFiveId":      bigFiveID,
		"bigFivePersonaTitle":  bigFiveTitle,
		"hasQuickSnapshot":     hasQuickSnapshot,
	})
}

func requireUser(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := auth.UserFromContext(r.Context())
	if claims == nil || claims.UserID == "" {
		writeErr(w, http.StatusUnauthorized, "请先登录")
		return nil
	}
	return claims
}
