package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ewallet-backend/db"
	"ewallet-backend/utils"
)

type contextkey string

const (
	UserIDContextKey contextkey = "userID"
)

func SignSafeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		userID := r.Header.Get("X-UserID")
		nonce := r.Header.Get("X-Nonce")
		signature := r.Header.Get("X-Signature")
		timestampStr := r.Header.Get("X-Timestamp")

		if nonce == "" || signature == "" || timestampStr == "" {
			http.Error(w, "Header tidak lengkap", http.StatusBadRequest)
			return
		}

		pubKey, err := db.GetUserPublicKey(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		var canonicalBody []byte
		if len(bodyBytes) == 0 {
			canonicalBody = []byte{}
		} else {
			canonicalBody, err = utils.CanonicalizeBody(bodyBytes)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to canonicalize body: %v", err), http.StatusBadRequest)
				return
			}
		}

		message := fmt.Sprintf("%s|%s|%s|%s", userID, timestampStr, nonce, string(canonicalBody))
		if err := VerifySignature(message, signature, pubKey); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		used, _ := db.RedisClient.Get(ctx, nonce).Result()
		if used != "" {
			http.Error(w, "replay attack detected (nonce used)", http.StatusUnauthorized)
			return
		}

		db.RedisClient.SetNX(ctx, nonce, "1", 5*time.Minute)

		if err := ValidateTimestamp(timestampStr); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctxWithUser := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := r.Header.Get("Origin")
		if allowedOrigin == "" {
			if r.Header.Get("Origin") != "" {
				w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "https://sveltekit-signsafe.vercel.app") // Fallback jika tidak ada header Origin
			}
		} else {
			// Jika ada header Origin, gunakan itu (atau validasi terhadap whitelist Anda)
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		// Log untuk debugging (opsional, bisa dihapus setelah masalah teratasi)
		// log.Printf("CorsMiddleware: Path: %s, Method: %s", r.URL.Path, r.Method)

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-UserID, X-Nonce, X-Timestamp, X-Signature, X-Requested-With, Origin, Accept") // Tambahkan Origin, Accept

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK) // Menggunakan StatusOK, bukan StatusNoContent, untuk potensi kompatibilitas yang lebih baik
			return                       // Penting untuk mengakhiri respons di sini untuk preflight
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserID(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDContextKey).(string)
	if !ok {
		return ""
	}
	return userID
}
