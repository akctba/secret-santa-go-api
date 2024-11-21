package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Group struct {
	GroupID       string    `json:"group_id"`
	Name          string    `json:"name"`
	DateCreated   time.Time `json:"date_created"`
	DateDraw      time.Time `json:"date_draw"`
	CreatorUserID int       `json:"creator_user_id"`
}

type User struct {
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	Password    string    `json:"password"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type Participant struct {
	GroupID      string    `json:"group_id"`
	UserID       int       `json:"user_id"`
	JoinedAt     time.Time `json:"joined_at"`
	FriendUserID int       `json:"friend_user_id"`
}

const (
	DbDriver = "sqlite3"
	DbName   = "secretsanta.db"
)

func main() {

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database tables
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS Users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_name TEXT,
		user_email TEXT,
		password TEXT,
		gender TEXT,
		date_of_birth TEXT
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS Groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_name TEXT,
		date_created TEXT,
		date_draw TEXT,
		creator_user_id INTEGER
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS Participants (
		group_id INTEGER,
		user_id INTEGER,
		joined_at TEXT,
		fried_user_id INTEGER,
		PRIMARY KEY (group_id, user_id)
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	r := mux.NewRouter()
	//User endpoints
	r.HandleFunc("/user", createUser).Methods("POST")
	r.HandleFunc("/user/signin", signin).Methods("POST")
	r.HandleFunc("/user/{id}", bearerAuth(getUser)).Methods("GET")
	//Group endpoints
	r.HandleFunc("/group", bearerAuth(createGroup)).Methods("POST")
	r.HandleFunc("/group/{id}", bearerAuth(getGroup)).Methods("GET")
	r.HandleFunc("/group/{id}/participant", bearerAuth(addParticipant)).Methods("POST")
	r.HandleFunc("/group/{id}/draw", bearerAuth(runDraw)).Methods("POST")
	r.HandleFunc("/group/{id}/friend", bearerAuth(getSecretFriend)).Methods("GET")

	http.ListenAndServe(":8080", r)
}

func signin(w http.ResponseWriter, r *http.Request) {
	//handle post request with user credentials to sign in, validate the user and return a token
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query the user from the database
	var dbUser User
	sqlStmt := `SELECT user_id, user_name, user_email, password
	FROM Users WHERE user_email = ?`
	row := db.QueryRow(sqlStmt, user.UserEmail)
	err = row.Scan(&dbUser.UserID, &dbUser.UserName, &dbUser.UserEmail, &dbUser.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("%q: %s\n", err, sqlStmt)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Compare the hashed password with the password from the request
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate a token
	token, err := generateToken(dbUser.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	//handle post request to create a user and save it on the database
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert the new user into the database
	sqlStmt := `
	INSERT INTO Users (user_name, user_email, password, gender, date_of_birth) VALUES (?, ?, ?, ?, ?);
	`

	_, err = db.Exec(sqlStmt, user.UserName, user.UserEmail, user.Password, user.Gender, user.DateOfBirth)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	//handle post request to create a group and save it on the database
	var group Group
	json.NewDecoder(r.Body).Decode(&group)

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert the new group into the database
	sqlStmt := `
	INSERT INTO Groups (group_name, date_created, creator_user_id) VALUES (?, ?, ?);
	`
	_, err = db.Exec(sqlStmt, group.GroupID, time.Now(), 1)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query the user from the database
	var user User
	sqlStmt := `SELECT user_id, user_name, user_email, password, gender, date_of_birth FROM Users WHERE user_id = ?`
	row := db.QueryRow(sqlStmt, userID)
	err = row.Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.Password, &user.Gender, &user.DateOfBirth)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("%q: %s\n", err, sqlStmt)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
func getGroup(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query the group from the database
	var group Group
	sqlStmt := `SELECT group_id, group_name, date_created, date_draw, creator_user_id FROM Groups WHERE group_id = ?`
	row := db.QueryRow(sqlStmt, groupID)
	err = row.Scan(&group.GroupID, &group.Name, &group.DateCreated, &group.DateDraw, &group.CreatorUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("%q: %s\n", err, sqlStmt)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func addParticipant(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	//handle post request to add a participant to a group and save it on the database
	var participant Participant
	json.NewDecoder(r.Body).Decode(&participant)

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert the new participant into the database
	sqlStmt := `
	INSERT INTO Participants (group_id, user_id, joined_at) VALUES (?, ?, ?);
	`
	_, err = db.Exec(sqlStmt, groupID, participant.UserID, time.Now())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(participant)
}

// runDraw will run the draw for the group and assign a secret friend to each participant
// get the list of participants for the group, shuffle it and assign the secret friend to each participant
func runDraw(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query the participants from the database
	var participants []Participant
	sqlStmt := `SELECT * FROM Participants WHERE group_id = ? AND friend_user_id IS NULL`
	rows, err := db.Query(sqlStmt, groupID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var participant Participant
		if err := rows.Scan(&participant); err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		participants = append(participants, participant)
	}

	if err := rows.Err(); err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Shuffle the participants
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	// Assign secret friends in a circular manner
	for i := range participants {
		if (i + 1) == len(participants) {
			participants[i].FriendUserID = participants[0].UserID
		} else {
			participants[i].FriendUserID = participants[i+1].UserID
		}

		// Update the participant in the database
		sqlStmt = `UPDATE Participants SET friend_user_id = ? WHERE group_id = ? AND user_id = ?`
		_, err = db.Exec(sqlStmt, participants[i].FriendUserID, groupID, participants[i].UserID)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getSecretFriend(w http.ResponseWriter, r *http.Request) {

}

// bearerAuth is a middleware function for Bearer Authentication
func bearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !validateToken(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// validateToken validates the Bearer token
func validateToken(token string) bool {
	// Replace with token validation logic
	return token == "your-secret-token"
}

// generateToken generates a token for the given user ID
func generateToken(userID int) (string, error) {
	// TODO: Replace with token generation logic
	return "your-secret-token", nil
}
