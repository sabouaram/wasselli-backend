package middlewares

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"wasselli-backend/resources"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func GenerateJWT(userID string, role string, duration time.Duration) (string, error) {
	var (
		tokenString string
		rsaPRIVATE  *rsa.PrivateKey
	)

	_, currentFile, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))))

	privateKeyPath := filepath.Join(dir, "runtime", "private.key")

	privateKeyData, err := os.ReadFile(privateKeyPath)

	if err != nil {
		return "", err
	}

	rsaPRIVATE, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)

	if err != nil {
		return "", err
	}

	claims := resources.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Wasselli App",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err = token.SignedString(rsaPRIVATE)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWT(token string) (claims *resources.Claims, ok bool) {

	var (
		err           error
		jwtToken      *jwt.Token
		publicKeyData []byte
		rsaPub        *rsa.PublicKey
	)

	_, currentFile, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))))

	publicKeyPath := filepath.Join(dir, "runtime", "public.pem")

	if publicKeyData, err = os.ReadFile(publicKeyPath); err != nil {
		return nil, false
	}

	rsaPub, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyData)

	if err != nil {
		return nil, false
	}

	jwtToken, err = jwt.ParseWithClaims(
		token,
		&resources.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok = token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return rsaPub, nil
		},
	)

	if err != nil || !jwtToken.Valid {
		return nil, false
	}

	claims, ok = jwtToken.Claims.(*resources.Claims)

	if !ok {
		return nil, false
	}

	return claims, true
}

func JwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	var (
		claims *resources.Claims
		ok     bool
	)

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, ok = validateJWT(tokenString)

		if !ok {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

func GetClaimsFromContext(r *http.Request) *resources.Claims {

	claims, ok := r.Context().Value(ClaimsKey).(*resources.Claims)

	if !ok || claims == nil {
		return nil
	}

	return claims
}
