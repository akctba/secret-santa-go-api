package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestHandlersReturnInternalServerErrorWhenDBUnavailable(t *testing.T) {
	originalGetDB := getDB
	getDB = func() (*sql.DB, error) {
		return nil, errors.New("db unavailable")
	}
	t.Cleanup(func() {
		getDB = originalGetDB
	})

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		vars    map[string]string
	}{
		{
			name:    "Signin returns 500 on DB failure",
			handler: Signin,
			method:  http.MethodPost,
			path:    "/user/signin",
			body:    `{"email":"a@example.com","password":"secret"}`,
		},
		{
			name:    "CreateUser returns 500 on DB failure",
			handler: CreateUser,
			method:  http.MethodPost,
			path:    "/user",
			body:    `{"user_name":"alice","email":"alice@example.com","password":"secret"}`,
		},
		{
			name:    "GetUser returns 500 on DB failure",
			handler: GetUser,
			method:  http.MethodGet,
			path:    "/user/1",
			vars:    map[string]string{"id": "1"},
		},
		{
			name:    "CreateGroup returns 500 on DB failure",
			handler: CreateGroup,
			method:  http.MethodPost,
			path:    "/group",
			body:    `{"group_name":"xmas"}`,
		},
		{
			name:    "GetGroup returns 500 on DB failure",
			handler: GetGroup,
			method:  http.MethodGet,
			path:    "/group/1",
			vars:    map[string]string{"id": "1"},
		},
		{
			name:    "AddParticipant returns 500 on DB failure",
			handler: AddParticipant,
			method:  http.MethodPost,
			path:    "/group/1/participant",
			body:    `{"group_id":"1","user_id":1}`,
			vars:    map[string]string{"id": "1"},
		},
		{
			name:    "RunDraw returns 500 on DB failure",
			handler: RunDraw,
			method:  http.MethodPost,
			path:    "/group/1/draw",
			vars:    map[string]string{"id": "1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}

			rr := httptest.NewRecorder()
			tc.handler(rr, req)

			if rr.Code != http.StatusInternalServerError {
				t.Fatalf("expected status %d, got %d, body: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
			}
		})
	}
}
