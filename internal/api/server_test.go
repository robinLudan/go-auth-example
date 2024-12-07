package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
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

func (s *StubStorage) GetUserByID(id uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (s *StubStorage) Register(user *models.User) error {
	return nil
}

func (s *StubStorage) GetUserByEmail(email string) (*models.User, error) {
	user, ok := s.userData[email]
	if !ok {
		return nil, errors.New("User not found")
	}
	return user, nil
}

func (s *StubStorage) Login(loginReq *models.LoginUserReq) error {
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
		assertJsonHeader(t, resp)
	})

	t.Run("returns error when payload has empty params", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "", // empty
			Password: "", // empty
		}
		resp := registerReq(createUser)
		assertEqual(t, resp.Code, http.StatusBadRequest)
		assertJsonHeader(t, resp)
	})

	t.Run("returns error when email is invalid", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "invalid", // doesn't include @
			Password: "secret",
		}
		resp := registerReq(createUser)
		assertEqual(t, resp.Code, http.StatusBadRequest)
		assertJsonHeader(t, resp)
	})

	t.Run("returns error when password is less than 8 chars", func(t *testing.T) {
		createUser := models.CreateUserRequest{
			Name:     "john doe",
			Email:    "johndoe@hotmail.com",
			Password: "secret", // less than 8 chars
		}
		resp := registerReq(createUser)
		assertStatus(t, resp.Code, http.StatusBadRequest)
		assertJsonHeader(t, resp)
	})
}

func TestHandleLogin(t *testing.T) {
	t.Run("validate email and password", func(t *testing.T) {
		server := NewApiServer(&StubStorage{}) // no user exits
		loginReq := models.LoginUserReq{
			Email:    "test@test.com",
			Password: "password",
		}
		payload, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)
		assertStatus(t, resp.Code, http.StatusUnauthorized)
	})

	t.Run("sets token on login", func(t *testing.T) {
		now := time.Now().UTC()
		password, _ := HashPassword("password")
		user := models.User{
			ID:        uuid.New(),
			Name:      "john",
			Email:     "john@test.com",
			Password:  password,
			CreatedAt: now,
			UpdatedAt: now,
		}

		store := &StubStorage{
			userData: make(map[string]*models.User),
		}
		store.userData[user.Email] = &user
		server := NewApiServer(store)

		reqPayload := models.LoginUserReq{
			Email:    user.Email,
			Password: "password",
		}
		payload, _ := json.Marshal(&reqPayload)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)

		var token string
		cookies := resp.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == AuthHeader {
				token = cookie.Value
				break
			}
		}
		if token == "" {
			t.Error("Cookie not set")
		}

		assertStatus(t, resp.Code, http.StatusOK)
	})
}

func assertJsonHeader(t testing.TB, resp *httptest.ResponseRecorder) {
	t.Helper()
	assertEqual(t, resp.Header().Get("Content-Type"), jsonContentType)
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
