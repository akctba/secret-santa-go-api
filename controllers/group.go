package controllers

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	randv2 "math/rand/v2"
	"net/http"
	"strings"

	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/models"
	"github.com/gorilla/mux"
)

// cryptoSource implements randv2.Source using crypto/rand for cryptographically secure randomness.
type cryptoSource struct{}

func (cryptoSource) Uint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("crypto/rand: failed to generate random number: %w", err))
	}
	return binary.LittleEndian.Uint64(b[:])
}

// CreateGroup handles POST /group. Persists a new group to the database.
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group models.Group
	if err := decodeRequestJSON(r, &group); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group.Name = strings.TrimSpace(group.Name)
	if group.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in CreateGroup: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
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

// GetGroup handles GET /group/{id}. Returns the group with the given ID.
func GetGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in GetGroup: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
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

// AddParticipant handles POST /group/{id}/participant. Adds a user to the group.
func AddParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	var request models.ParticipantRequest
	if err := decodeRequestJSON(r, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	request.GroupID = strings.TrimSpace(request.GroupID)
	if request.GroupID == "" || request.UserID <= 0 {
		http.Error(w, "group_id and user_id are required", http.StatusBadRequest)
		return
	}

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in AddParticipant: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
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

// RunDraw handles POST /group/{id}/draw. Shuffles participants and assigns secret friends.
func RunDraw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	db, err := getDB()
	if err != nil {
		log.Printf("failed to open db in RunDraw: %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.CloseDb(db)

	participants, err := database.GetParticipantsToDraw(db, groupID)
	if err != nil {
		http.Error(w, "Failed to get participants", http.StatusInternalServerError)
		return
	}

	if len(participants) == 0 {
		http.Error(w, "No participants to draw", http.StatusBadRequest)
		return
	}

	// A new Rand is created per request because math/rand/v2.Rand is not safe for concurrent use.
	// cryptoSource itself is stateless so construction overhead is negligible.
	randv2.New(cryptoSource{}).Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	// Assign secret friends in a circular manner so the last participant
	// receives the first as their secret friend.
	for i := range participants {
		if i+1 == len(participants) {
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

// GetSecretFriend handles GET /group/{id}/friend. Returns the authenticated user's assigned friend.
func GetSecretFriend(w http.ResponseWriter, r *http.Request) {
}
