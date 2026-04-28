package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

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

// Signin handles POST /user/signin. Validates credentials and returns a bearer token.
func Signin(w http.ResponseWriter, r *http.Request) {
	var request models.UserSignin
	json.NewDecoder(r.Body).Decode(&request)

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
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

	token, err := auth.CreateToken(user.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// CreateUser handles POST /user. Hashes the password and persists the new user.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
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

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	err = database.InsertUser(db, user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser handles GET /user/{id}. Returns the user with the given ID.
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	user, err := database.GetUserByID(db, userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
