package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/robinLudan/user-auth/internal/models"
	"github.com/robinLudan/user-auth/internal/storage"
	"github.com/robinLudan/user-auth/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type ApiServer struct {
	store storage.Storage
	http.Handler
}

var jsonContentType = "application/json"

func NewApiServer(store storage.Storage) *ApiServer {
	s := new(ApiServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle("POST /register", http.HandlerFunc(s.handleRegister))

	s.Handler = router
	return s
}

func (s *ApiServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	createUserReq := new(models.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		respondWithClientErr(w, http.StatusBadRequest, "Invalid payload")
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
