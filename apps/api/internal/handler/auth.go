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
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"username": user.Username,
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
