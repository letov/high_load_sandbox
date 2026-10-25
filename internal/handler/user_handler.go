package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/letov/high_load_sandbox/internal/auth"
	"github.com/letov/high_load_sandbox/internal/model"
	"github.com/letov/high_load_sandbox/internal/repository"
)

type UserHandler struct {
	users     *repository.UserRepository
	jwtSecret []byte
}

func NewUserHandler(users *repository.UserRepository, jwtSecret []byte) *UserHandler {
	return &UserHandler{users: users, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
	City      string `json:"city"`
}

type registerResponse struct {
	ID int64 `json:"id"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Login) == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "login and password are required")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("hash password: %v", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	user := model.User{
		Login:     req.Login,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		BirthDate: req.BirthDate,
		Gender:    req.Gender,
		Interests: req.Interests,
		City:      req.City,
	}

	id, err := h.users.Create(r.Context(), user, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrLoginTaken) {
			writeError(w, http.StatusConflict, "login already taken")
			return
		}
		log.Printf("create user: %v", err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{ID: id})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		writeError(w, http.StatusUnauthorized, "missing or invalid bearer token")
		return
	}

	idParam := r.PathValue("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		log.Printf("get user: %v", err)
		writeError(w, http.StatusInternalServerError, "could not fetch user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) authenticate(r *http.Request) (int64, error) {
	const prefix = "Bearer "

	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return 0, errors.New("missing bearer token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return 0, errors.New("empty bearer token")
	}

	return auth.ParseJWT(h.jwtSecret, token)
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, hash, err := h.users.GetPasswordHashByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid login or password")
			return
		}
		log.Printf("get password hash: %v", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	if !auth.ComparePassword(hash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid login or password")
		return
	}

	token, err := auth.GenerateJWT(h.jwtSecret, userID)
	if err != nil {
		log.Printf("generate jwt: %v", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json response: %v", err)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
