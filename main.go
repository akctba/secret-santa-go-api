package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/models"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database.CreateTables()

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

	// Compare the hashed password with the password from the request
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate a token
	token, err := CreateToken(user.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	//handle post request to create a user and save it on the database
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Open a connection to the SQLite database
	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	err = database.InsertUser(db, user)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	//handle post request to create a group and save it on the database
	var group models.Group
	json.NewDecoder(r.Body).Decode(&group)

	// Open a connection to the SQLite database
	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	err = database.InsertGroup(db, group)
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL
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

func getGroup(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	group, err := database.GetGroupByID(db, groupID)
	if err != nil {
		http.Error(w, "Failed to get group", http.StatusInternalServerError)
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
	var request models.ParticipantRequest
	json.NewDecoder(r.Body).Decode(&request)

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	group, err := database.GetGroupByID(db, groupID)
	if err != nil {
		http.Error(w, "Failed to get group", http.StatusInternalServerError)
		return
	}
	if request.GroupID != group.GroupID {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	err = database.InsertParticipant(db, request)
	if err != nil {
		http.Error(w, "Failed to add participant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(request)
}

// runDraw will run the draw for the group and assign a secret friend to each participant
// get the list of participants for the group, shuffle it and assign the secret friend to each participant
func runDraw(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDb(db)

	// Query the participants from the database
	participants, err := database.GetParticipantsToDraw(db, groupID)
	if err != nil {
		http.Error(w, "Failed to get participants", http.StatusInternalServerError)
		return
	}

	// If there are no participants to draw, return an error
	if len(participants) == 0 {
		http.Error(w, "No participants to draw", http.StatusBadRequest)
		return
	}

	// Shuffle the participants
	rand.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	// Assign secret friends in a circular manner
	// so the last participant will have the first participant as a secret friend
	// and that is what makes the event more fun!
	for i := range participants {
		if (i + 1) == len(participants) {
			participants[i].FriendUserID = participants[0].UserID
		} else {
			participants[i].FriendUserID = participants[i+1].UserID
		}

		err = database.UpdateParticipant(db, participants[i])
		if err != nil {
			http.Error(w, "Failed to update participant", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
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
		_, err := ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
