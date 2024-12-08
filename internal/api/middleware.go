package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/robinLudan/user-auth/internal/models"
	"github.com/robinLudan/user-auth/internal/storage"
)

type authHandler func(http.ResponseWriter, *http.Request, *models.User)

func (s *ApiServer) auth(handler authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AuthHeader)
		if err != nil {
			respondWithClientErr(w, http.StatusUnauthorized, err.Error())
			return
		}

		userID, err := verifyToken(cookie.Value)
		if err != nil {
			respondWithClientErr(w, http.StatusUnauthorized, err.Error())
			return
		}

		user, err := s.store.GetUserByID(userID)
		if err != nil {
			if err == storage.ErrUserNotFound || user == nil {
				respondWithClientErr(w, http.StatusUnauthorized, err.Error())
				return
			}
			respondWithInternalErr(w, fmt.Sprintf("Error getting user with id: %v", err))
			return
		}

		handler(w, r, user)
	}
}

func logger(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("%s %s from %s, took %s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
		}()

		handler(w, r)
	}
}

type middleware func(http.HandlerFunc) http.HandlerFunc

func chain(handler http.HandlerFunc, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}
