package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/robinLudan/go-auth-example/internal/models"
	"github.com/robinLudan/go-auth-example/internal/storage"
	"github.com/robinLudan/go-auth-example/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type ApiServer struct {
	store storage.Storage
	http.Handler
}

var (
	jsonContentType = "application/json"
	invalidPayload  = "invalid payload"
)

func NewApiServer(store storage.Storage) *ApiServer {
	s := new(ApiServer)
	s.store = store

	router := http.NewServeMux()

	router.Handle("POST /register", chain(s.handleRegister, logger))
	router.Handle("POST /login", chain(s.handleLogin, logger))
	router.Handle("GET /me", chain(s.auth(s.handleGetUser), logger))

	s.Handler = router
	return s
}

func (s *ApiServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	createUserReq := new(models.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		respondWithClientErr(w, http.StatusBadRequest, invalidPayload)
		return
	}

	if err := utils.Validate.Struct(createUserReq); err != nil {
		respondWithClientErr(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := HashPassword(createUserReq.Password)
	if err != nil {
		respondWithInternalErr(w, fmt.Sprintf("Failed to hash password: %v", err))
		return
	}

	now := time.Now()
	newUser := &models.User{
		ID:        uuid.New(),
		Name:      createUserReq.Name,
		Email:     createUserReq.Email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Register(newUser); err != nil {
		if err == storage.ErrEmailExists {
			respondWithClientErr(w, http.StatusConflict, err.Error())
			return
		}
		respondWithInternalErr(w, fmt.Sprintf("Failed to register user: %v", err))
		return
	}

	respondWithJson(w, http.StatusCreated, newUser, "user")
}

func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	reqPayload := new(models.LoginUserReq)
	if err := json.NewDecoder(r.Body).Decode(reqPayload); err != nil {
		respondWithClientErr(w, http.StatusBadRequest, invalidPayload)
		return
	}

	user, err := s.store.GetUserByEmail(reqPayload.Email)
	if err != nil {
		if err == storage.ErrUserNotFound || user == nil {
			respondWithClientErr(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondWithInternalErr(w, fmt.Sprintf("Error getting user with email: %v", err))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqPayload.Password)); err != nil {
		respondWithClientErr(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := createToken(user.ID)
	if err != nil {
		respondWithInternalErr(w, fmt.Sprintf("Error creating JWT key: %v", err))
		return
	}

	// set the token as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthHeader,
		Value:    token,
		Expires:  time.Now().Add(expiryDuration),
		HttpOnly: true,
		Secure:   false, // TODO: set to true in production
	})
}

func (s *ApiServer) handleGetUser(w http.ResponseWriter, r *http.Request, user *models.User) {
	respondWithJson(w, http.StatusOK, user, "user")
}

func respondWithJson(w http.ResponseWriter, code int, model interface{}, key string) {
	writeHeader(w, code)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			key: model,
		},
	})
	if err != nil {
		respondWithInternalErr(w, fmt.Sprintf("Failed to parse response: %v", err))
	}
}

func respondWithClientErr(w http.ResponseWriter, code int, message string) {
	writeHeader(w, code)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
		},
	})
	if err != nil {
		respondWithInternalErr(w, fmt.Sprintf("Failed to parse response: %v", err))
	}
}

func respondWithInternalErr(w http.ResponseWriter, message string) {
	writeHeader(w, http.StatusInternalServerError)
	log.Println(message)
}

func writeHeader(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(code)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
