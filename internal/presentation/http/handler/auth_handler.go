package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"authmicro/internal/domain/service"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegistrationRequest represents the request for starting registration
type RegistrationRequest struct {
	FirstName             string `json:"firstName"`
	LastName              string `json:"lastName"`
	Nickname              string `json:"nickname"`
	Email                 string `json:"email"`
	AcceptedPrivacyPolicy bool   `json:"acceptedPrivacyPolicy"`
}

// RegistrationResponse represents the response for starting registration
type RegistrationResponse struct {
	SessionID string `json:"sessionId"`
	Email     string `json:"email"`
}

// StartRegistration handles the registration start request
func (h *AuthHandler) StartRegistration(w http.ResponseWriter, r *http.Request) {
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Start registration process
	session, err := h.authService.CreateRegistrationSession(
		r.Context(),
		req.FirstName,
		req.LastName,
		req.Nickname,
		req.Email,
		req.AcceptedPrivacyPolicy,
	)

	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			respondWithError(w, http.StatusBadRequest, "Invalid input")
		case errors.Is(err, service.ErrInvalidEmail):
			respondWithError(w, http.StatusBadRequest, "Invalid email format")
		case errors.Is(err, service.ErrInvalidNickname):
			respondWithError(w, http.StatusBadRequest, "Invalid nickname format")
		case errors.Is(err, service.ErrNicknameExists):
			respondWithError(w, http.StatusConflict, "Nickname already exists")
		case errors.Is(err, service.ErrPrivacyPolicyRequired):
			respondWithError(w, http.StatusBadRequest, "Privacy policy acceptance required")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to start registration")
		}
		return
	}

	// Respond with session ID
	resp := RegistrationResponse{
		SessionID: session.ID,
		Email:     session.Email,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// VerificationRequest represents the request for email verification
type VerificationRequest struct {
	SessionID string `json:"sessionId"`
	Code      string `json:"code"`
}

// VerifyEmail handles the email verification request
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify email
	err := h.authService.ConfirmEmail(r.Context(), req.SessionID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			respondWithError(w, http.StatusNotFound, "Session not found")
		case errors.Is(err, service.ErrInvalidCode):
			respondWithError(w, http.StatusBadRequest, "Invalid or expired verification code")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// ResendCodeRequest represents the request for resending verification code
type ResendCodeRequest struct {
	SessionID string `json:"sessionId"`
}

// ResendVerificationCode handles the request to resend verification code
func (h *AuthHandler) ResendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req ResendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Resend verification code
	session, err := h.authService.ResendCodeEmail(r.Context(), req.SessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			respondWithError(w, http.StatusNotFound, "Session not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to resend verification code")
		}
		return
	}

	resp := RegistrationResponse{
		SessionID: session.ID,
		Email:     session.Email,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// LoginRequest represents the request for login
type LoginRequest struct {
	Email string `json:"email"`
}

// LoginResponse represents the response for login
type LoginResponse struct {
	Email string `json:"email"`
}

// Login handles the login request
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Start login process
	session, err := h.authService.SendLoginCodeEmail(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			respondWithError(w, http.StatusBadRequest, "Invalid email format")
		default:
			// Don't reveal if the email exists or not for security
			respondWithJSON(w, http.StatusOK, LoginResponse{Email: req.Email})
			return
		}
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{Email: session.Email})
}

// VerifyLoginRequest represents the request for verifying login
type VerifyLoginRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// VerifyLoginResponse represents the response for verifying login
type VerifyLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// VerifyLogin handles the login verification request
func (h *AuthHandler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	var req VerifyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify login
	tokens, err := h.authService.ConfirmLogin(r.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCode):
			respondWithError(w, http.StatusBadRequest, "Invalid or expired verification code")
		case errors.Is(err, service.ErrUserNotFound):
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to verify login")
		}
		return
	}

	resp := VerifyLoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// RefreshTokenRequest represents the request for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshToken handles the token refresh request
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest

	// Try to get token from request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body parsing fails, try to get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			req.RefreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid request format")
			return
		}
	}

	// Refresh token
	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenExpired):
			respondWithError(w, http.StatusUnauthorized, "Token has expired")
		case errors.Is(err, service.ErrTokenInvalid):
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
		case errors.Is(err, service.ErrUserNotFound):
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to refresh token")
		}
		return
	}

	resp := VerifyLoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Error marshaling JSON response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
