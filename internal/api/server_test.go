package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robinLudan/user-auth/internal/models"
)

type StubStorage struct {
	userData map[string]*models.User
}

func NewStubStorage() *StubStorage {
	return &StubStorage{
		userData: make(map[string]*models.User),
	}
}

func (s *StubStorage) Register(user *models.User) error {
	return nil
}

func TestHandleSignUpUser(t *testing.T) {
	registerReq := func(userPayload models.CreateUserRequest) *httptest.ResponseRecorder {
		server := NewApiServer(&StubStorage{})
		payload, _ := json.Marshal(userPayload)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		return resp
	}

	t.Run("register user", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "johndoe@hotmail.com",
			Password: "mytopsecretlongpassword",
		}
		resp := registerReq(createUser)
		assertStatus(t, resp.Code, http.StatusCreated)
		assertEqual(t, resp.Result().Header.Get("Content-Type"), jsonContentType)
	})

	t.Run("returns error when payload has empty params", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "", // empty
			Password: "", // empty
		}
		resp := registerReq(createUser)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})

	t.Run("returns error when email is invalid", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "invalid", // doesn't include @
			Password: "secret",
		}
		resp := registerReq(createUser)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})

	t.Run("returns error when password is less than 8 chars", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "johndoe@hotmail.com",
			Password: "secret", // less than 8 chars
		}
		resp := registerReq(createUser)
		assertStatus(t, resp.Code, http.StatusBadRequest)
	})
}

func assertEqual(t testing.TB, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("Got %v, Want %v", got, want)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("Got status %d, Want %d", got, want)
	}
}
