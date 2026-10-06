package http

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/umam-src/catatan/internal/db"
)

const (
	sessionCookieName  = "catatan_session"
	sessionLifetime    = 24 * time.Hour
	passwordIterations = 600000
)

type authServer struct {
	db       *db.DB
	mu       sync.Mutex
	failures map[string]loginFailure
}

type loginFailure struct {
	count int
	until time.Time
}

type sessionUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func newAuthServer(d *db.DB) *authServer {
	return &authServer{db: d, failures: make(map[string]loginFailure)}
}

func (a *authServer) setup(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "asal permintaan tidak diizinkan", http.StatusForbidden)
		return
	}
	var count int
	if err := a.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM auth_credentials").Scan(&count); err != nil {
		serverError(w, err)
		return
	}
	if count != 0 {
		http.Error(w, "akun awal sudah dikonfigurasi", http.StatusConflict)
		return
	}
	var in struct {
		Username string `json:"username"`
		Email string `json:"email"`
		DisplayName string `json:"display_name"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) { return }
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.TrimSpace(in.Email)
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if err := validateLength(in.Username, 3, 80, "Nama pengguna"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := validateLength(in.Email, 0, 254, "Email"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := validateLength(in.DisplayName, 0, 120, "Nama tampilan"); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if len(in.Password) < 12 || len(in.Password) > 256 { http.Error(w, "Kata sandi harus 12 sampai 256 karakter", http.StatusBadRequest); return }

	hash, err := hashPassword(in.Password)
	if err != nil { serverError(w, err); return }
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = a.db.ExecContext(r.Context(), "UPDATE users SET username=?, email=?, display_name=?, updated_at=? WHERE id='local'", in.Username, in.Email, in.DisplayName, now); err != nil { serverError(w, err); return }
	if _, err = a.db.ExecContext(r.Context(), "INSERT INTO auth_credentials(user_id,password_hash,updated_at) VALUES('local',?,?)", hash, now); err != nil { serverError(w, err); return }
	user := sessionUser{ID: "local", Username: in.Username, DisplayName: in.DisplayName}
	if err := a.startSession(w, r, user.ID); err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (a *authServer) login(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "asal permintaan tidak diizinkan", http.StatusForbidden)
		return
	}
	key := clientKey(r)
	if !a.allowAttempt(key) { http.Error(w, "terlalu banyak percobaan login", http.StatusTooManyRequests); return }
	var in struct { Username string `json:"username"`; Password string `json:"password"` }
	if !decode(w, r, &in) { return }
	in.Username = strings.TrimSpace(in.Username)
	if len(in.Username) == 0 || len(in.Password) == 0 { http.Error(w, "nama pengguna dan kata sandi wajib diisi", http.StatusBadRequest); return }

	var user sessionUser
	var stored string
	err := a.db.QueryRowContext(r.Context(), "SELECT u.id,u.username,u.display_name,c.password_hash FROM users u JOIN auth_credentials c ON c.user_id=u.id WHERE u.username=? LIMIT 1", in.Username).Scan(&user.ID, &user.Username, &user.DisplayName, &stored)
	if err != nil || !verifyPassword(stored, in.Password) {
		a.recordFailure(key)
		http.Error(w, "nama pengguna atau kata sandi salah", http.StatusUnauthorized)
		return
	}
	a.clearFailures(key)
	if err := a.startSession(w, r, user.ID); err != nil { serverError(w, err); return }
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (a *authServer) logout(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "asal permintaan tidak diizinkan", http.StatusForbidden)
		return
	}
	if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", time.Now().UTC().Format(time.RFC3339Nano), tokenHash(c.Value))
	}
	http.SetCookie(w, clearSessionCookie(r))
	w.WriteHeader(http.StatusNoContent)
}

func (a *authServer) me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r.Context())
	if !ok { http.Error(w, "belum masuk", http.StatusUnauthorized); return }
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func withAuth(a *authServer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/auth/setup" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/logout" || r.URL.Path == "/api/auth/me" {
			next.ServeHTTP(w, r)
			return
		}
		user, ok := a.authenticate(r)
		if !ok { http.Error(w, "autentikasi diperlukan", http.StatusUnauthorized); return }
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authContextKey{}, user)))
	})
}

type authContextKey struct{}

func (a *authServer) authenticate(r *http.Request) (sessionUser, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" { return sessionUser{}, false }
	var u sessionUser
	var expires string
	err = a.db.QueryRowContext(r.Context(), "SELECT u.id,u.username,u.display_name,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.revoked_at IS NULL LIMIT 1", tokenHash(c.Value)).Scan(&u.ID, &u.Username, &u.DisplayName, &expires)
	if err != nil { return sessionUser{}, false }
	t, err := time.Parse(time.RFC3339Nano, expires)
	if err != nil || !t.After(time.Now().UTC()) {
		_, _ = a.db.ExecContext(r.Context(), "UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", time.Now().UTC().Format(time.RFC3339Nano), tokenHash(c.Value))
		return sessionUser{}, false
	}
	return u, true
}

func (a *authServer) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	token, err := randomToken(32)
	if err != nil { return err }
	now := time.Now().UTC()
	expires := now.Add(sessionLifetime)
	_, err = a.db.ExecContext(r.Context(), "INSERT INTO sessions(id,user_id,token_hash,expires_at,created_at) VALUES(?,?,?,?,?)", randomID(), userID, tokenHash(token), expires.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil { return err }
	http.SetCookie(w, sessionCookie(r, token, expires))
	return nil
}

func sessionCookie(r *http.Request, token string, expires time.Time) *http.Cookie {
	return &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", Expires: expires, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode}
}

func clearSessionCookie(r *http.Request) *http.Cookie {
	return &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode}
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil { return "", err }
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { panic("sumber acak sistem tidak tersedia") }
	return hex.EncodeToString(b)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil { return "", err }
	dk := pbkdf2SHA256([]byte(password), salt, passwordIterations, 32)
	return "pbkdf2-sha256$" + strconv.Itoa(passwordIterations) + "$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(dk), nil
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" { return false }
	iters, err := strconv.Atoi(parts[1])
	if err != nil || iters < 100000 || iters > 2000000 { return false }
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil { return false }
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil { return false }
	got := pbkdf2SHA256([]byte(password), salt, iters, len(want))
	return hmac.Equal(got, want)
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	const hashLen = 32
	out := make([]byte, 0, keyLen)
	for block := 1; len(out) < keyLen; block++ {
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(salt)
		_, _ = mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for j := 0; j < hashLen; j++ { t[j] ^= u[j] }
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func (a *authServer) allowAttempt(key string) bool {
	a.mu.Lock(); defer a.mu.Unlock()
	f := a.failures[key]
	if time.Now().Before(f.until) { return false }
	return true
}

func (a *authServer) recordFailure(key string) {
	a.mu.Lock(); defer a.mu.Unlock()
	f := a.failures[key]
	f.count++
	if f.count >= 10 { f.until = time.Now().Add(15 * time.Minute); f.count = 0 }
	a.failures[key] = f
}

func (a *authServer) clearFailures(key string) {
	a.mu.Lock(); delete(a.failures, key); a.mu.Unlock()
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil { return host }
	return r.RemoteAddr
}

func userFromRequest(ctx context.Context) (sessionUser, bool) {
	user, ok := ctx.Value(authContextKey{}).(sessionUser)
	return user, ok
}

func userID(r *http.Request) string {
	user, _ := userFromRequest(r.Context())
	return user.ID
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return origin == scheme+"://"+r.Host
}
