package helper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"api-students/app/model"
)

// Sentinel error untuk membedakan jenis kegagalan token.
// Middleware menggunakan ini untuk mengirim pesan yang berbeda:
// - ErrExpiredToken → "token sudah kedaluwarsa"
// - ErrInvalidToken → "token tidak valid"
var (
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
	ErrInvalidToken = errors.New("token tidak valid")
)

// JWTManager mengelola pembuatan dan parsing access token JWT.
type JWTManager struct {
	secret   []byte
	issuer   string
	ttl      time.Duration
}

// NewJWTManager membuat instance baru. secret harus minimal 32 karakter.
func NewJWTManager(secret, issuer string, ttlMinutes int) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

// GenerateAccess membuat access token JWT dengan algoritma HS256.
// Claims yang disertakan: sub (user id), iss, iat, exp, username, role.
func (m *JWTManager) GenerateAccess(userID int, username, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      strconv.Itoa(userID),
		"iss":      m.issuer,
		"iat":      now.Unix(),
		"exp":      now.Add(m.ttl).Unix(),
		"username": username,
		"role":     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("gagal menandatangani token: %w", err)
	}
	return signed, nil
}

// Parse memverifikasi dan mem-parsing access token.
// WAJIB memeriksa signing method secara eksplisit di keyfunc
// untuk mencegah algorithm confusion / alg:none attack.
func (m *JWTManager) Parse(tokenString string) (*model.AuthUser, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Cegah algorithm confusion: TOLAK jika bukan HMAC.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak diharapkan: %v", t.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		// Bedakan expired vs invalid.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Ambil data dari claims.
	sub, _ := claims.GetSubject()
	userID, err := strconv.Atoi(sub)
	if err != nil {
		return nil, ErrInvalidToken
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	return &model.AuthUser{
		ID:       userID,
		Username: username,
		Role:     role,
	}, nil
}
