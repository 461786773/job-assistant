package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"uname"`
	Exp      int64  `json:"exp"`
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

// LoadOrCreateSecret reads JA_JWT_SECRET, or a persisted file, or creates one.
func LoadOrCreateSecret(envSecret, persistPath string) (string, error) {
	if s := strings.TrimSpace(envSecret); s != "" {
		return s, nil
	}
	if data, err := os.ReadFile(persistPath); err == nil {
		s := strings.TrimSpace(string(data))
		if s != "" {
			return s, nil
		}
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := base64.RawURLEncoding.EncodeToString(b)
	if err := os.WriteFile(persistPath, []byte(s+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("persist jwt secret: %w", err)
	}
	return s, nil
}

func (m *TokenManager) Issue(userID, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Exp:      time.Now().Add(m.ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(enc))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return enc + "." + sig, nil
}

func (m *TokenManager) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效令牌")
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, fmt.Errorf("无效令牌")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("无效令牌")
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, fmt.Errorf("无效令牌")
	}
	if claims.UserID == "" || claims.Exp < time.Now().Unix() {
		return nil, fmt.Errorf("登录已过期，请重新登录")
	}
	return &claims, nil
}
