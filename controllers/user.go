package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akctba/secret-santa-go-api/auth"
	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type createUserRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type signinResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type userResponse struct {
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

func toUserResponse(user models.User) userResponse {
	resp := userResponse{
		UserID:      user.UserID,
		UserName:    user.UserName,
		UserEmail:   user.UserEmail,
		Gender:      user.Gender,
		DateOfBirth: user.DateOfBirth,
	}

	return resp
}

// Signin handles POST /user/signin. Validates credentials and returns access and refresh tokens.
func Signin(w http.ResponseWriter, r *http.Request) {
	var request models.UserSignin
	if err := decodeRequestJSON(r, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(request.UserEmail)
	password := strings.TrimSpace(request.Password)
	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	request.UserEmail = email
	request.Password = password

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in Signin: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.CloseDb(db)

	user, err := database.GetUserByEmail(db, request.UserEmail)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.CreateAccessToken(user.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.CreateRefreshToken(user.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(signinResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken handles POST /user/refresh. Validates a refresh token and returns a new access token.
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var request refreshTokenRequest
	if err := decodeRequestJSON(r, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	refreshToken := strings.TrimSpace(request.RefreshToken)
	if refreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	userID, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.CreateAccessToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(refreshTokenResponse{AccessToken: accessToken})
}

// CreateUser handles POST /user. Hashes the password and persists the new user.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeRequestJSON(r, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(request.UserName)
	email := strings.TrimSpace(request.Email)

	password := request.Password

	if name == "" || email == "" || password == "" {
		http.Error(w, "user_name, email and password are required", http.StatusBadRequest)
		return
	}

	user := models.User{
		UserName:  name,
		UserEmail: email,
		Password:  password,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in CreateUser: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.CloseDb(db)

	err = database.InsertUser(db, user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toUserResponse(user))
}

// GetUser handles GET /user/{id}. Returns the user with the given ID.
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in GetUser: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.CloseDb(db)

	user, err := database.GetUserByID(db, userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toUserResponse(user))
}
