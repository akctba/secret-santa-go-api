package controllers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/models"
	"github.com/gorilla/mux"
)

// CreateGroup handles POST /group. Persists a new group to the database.
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group models.Group
	json.NewDecoder(r.Body).Decode(&group)

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

// GetGroup handles GET /group/{id}. Returns the group with the given ID.
func GetGroup(w http.ResponseWriter, r *http.Request) {
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

// AddParticipant handles POST /group/{id}/participant. Adds a user to the group.
func AddParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

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

// RunDraw handles POST /group/{id}/draw. Shuffles participants and assigns secret friends.
func RunDraw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	db, err := database.GetDb()
	if err != nil {
		log.Fatal(err)
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

	rand.Shuffle(len(participants), func(i, j int) {
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
