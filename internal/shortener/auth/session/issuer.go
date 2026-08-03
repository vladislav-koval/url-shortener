package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Issuer подписывает и проверяет сессионный токен: собственная минимальная
// реализация JWT (HS256) через stdlib crypto/hmac, без внешней библиотеки —
// в этой среде разработки нет доступа к proxy.golang.org, чтобы подтянуть
// стандартный golang-jwt/jwt. Публичный контракт (Issue/Verify) сделан таким,
// чтобы позже можно было безболезненно заменить на `go get
// github.com/golang-jwt/jwt/v5` и переписать только этот файл — остальной код
// (service, транспорт) от конкретной реализации не зависит.
//
// Алгоритм всегда HS256 и не читается из токена — так называемая
// "alg confusion"-атака (токен с alg=none или другим алгоритмом) здесь
// структурно невозможна, а не просто отклоняется валидацией.
//
// Verify пока не используется нигде в приложении — авторизации остальных
// запросов ещё нет, это следующий шаг (мидлварь поверх этого же Issuer).
// Issue без Verify был бы бесполезен по построению — это не задел на
// будущее, а обе стороны одного примитива.
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

var ErrInvalidToken = errors.New("invalid session token")

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func NewIssuer(cfg Config) *Issuer {
	return &Issuer{
		secret: []byte(cfg.Secret),
		ttl:    cfg.TTL,
	}
}

func (i *Issuer) Issue(userID uuid.UUID) (string, error) {
	now := time.Now()

	headerJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(jwtClaims{
		Sub: userID.String(),
		Iat: now.Unix(),
		Exp: now.Add(i.ttl).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	signingInput := encodeSegment(headerJSON) + "." + encodeSegment(claimsJSON)

	return signingInput + "." + encodeSegment(i.sign(signingInput)), nil
}

func (i *Issuer) Verify(token string) (uuid.UUID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("malformed token: %w", ErrInvalidToken)
	}

	headerPart, claimsPart, signaturePart := parts[0], parts[1], parts[2]

	var h jwtHeader
	if err := decodeSegment(headerPart, &h); err != nil || h.Alg != "HS256" {
		return uuid.Nil, fmt.Errorf("unsupported jwt header: %w", ErrInvalidToken)
	}

	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return uuid.Nil, fmt.Errorf("decode signature: %w", ErrInvalidToken)
	}

	expected := i.sign(headerPart + "." + claimsPart)
	if subtle.ConstantTimeCompare(signature, expected) != 1 {
		return uuid.Nil, fmt.Errorf("signature mismatch: %w", ErrInvalidToken)
	}

	var c jwtClaims
	if err := decodeSegment(claimsPart, &c); err != nil {
		return uuid.Nil, fmt.Errorf("decode claims: %w", ErrInvalidToken)
	}

	if time.Now().Unix() >= c.Exp {
		return uuid.Nil, fmt.Errorf("token expired: %w", ErrInvalidToken)
	}

	userID, err := uuid.Parse(c.Sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse user id from token subject: %w", ErrInvalidToken)
	}

	return userID, nil
}

func (i *Issuer) sign(signingInput string) []byte {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(signingInput))

	return mac.Sum(nil)
}

func encodeSegment(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeSegment(segment string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, v)
}
