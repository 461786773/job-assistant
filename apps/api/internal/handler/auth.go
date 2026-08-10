package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zhangyongjie/job-assistant/internal/auth"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{2,32}$`)

type registerBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效的 JSON")
		return
	}
	username := strings.TrimSpace(body.Username)
	password := body.Password
	if !usernameRE.MatchString(username) {
		writeErr(w, http.StatusBadRequest, "用户名需 2–32 位字母、数字或下划线")
		return
	}
	if utf8.RuneCountInString(password) < 6 {
		writeErr(w, http.StatusBadRequest, "密码至少 6 位")
		return
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "无法创建账号")
		return
	}
	user := &db.User{
		ID:           newID(),
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    db.Now(),
	}
	if err := h.Store.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "用户名已存在") {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := h.Tokens.Issue(user.ID, user.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效的 JSON")
		return
	}
	username := strings.TrimSpace(body.Username)
	user, err := h.Store.GetUserByUsername(username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil || !auth.CheckPassword(user.PasswordHash, body.Password) {
		writeErr(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	token, err := h.Tokens.Issue(user.ID, user.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
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
