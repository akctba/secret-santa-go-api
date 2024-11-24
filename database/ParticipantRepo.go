package database

//this file will contain all the database operations for the Participant model

import (
	"database/sql"
	"log"
	"time"

	"github.com/akctba/secret-santa-go-api/models"
)

func InsertParticipant(db *sql.DB, participant models.ParticipantRequest) error {
	sqlStmt := `INSERT INTO Participants(group_id, user_id, joined_at
	) VALUES (?, ?, ?);`
	_, err := db.Exec(sqlStmt, participant.GroupID, participant.UserID, time.Now())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func UpdateParticipant(db *sql.DB, participant models.Participant) error {
	sqlStmt := `UPDATE Participants SET joined_at = ?, friend_user_id = ?
	WHERE user_id = ? AND group_id = ?;`
	_, err := db.Exec(sqlStmt, participant.JoinedAt, participant.FriendUserID, participant.UserID, participant.GroupID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func DeleteParticipant(db *sql.DB, userId int, groupId int) error {
	sqlStmt := `DELETE FROM Participants WHERE user_id = ? AND group_id = ?;`
	_, err := db.Exec(sqlStmt, userId, groupId)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func GetParticipantByUserID(db *sql.DB, userID string) ([]models.UserParticipant, error) {
	var participants []models.UserParticipant
	sqlStmt := `SELECT group_id, user_id, joined_at, user_name, user_email, gender, date_of_birth
	FROM Participants WHERE user_id = ?;`
	rows, err := db.Query(sqlStmt, userID)
	if err != nil {
		return participants, err
	}
	for rows.Next() {
		var participant models.UserParticipant
		err := rows.Scan(&participant.GroupID, &participant.UserID, &participant.JoinedAt, &participant.UserName,
			&participant.UserEmail, &participant.Gender, &participant.DateOfBirth)
		if err != nil {
			return participants, err
		}
		participants = append(participants, participant)
	}
	return participants, nil
}

func GetParticipantsByGroupID(db *sql.DB, id string) ([]models.UserParticipant, error) {
	var participants []models.UserParticipant
	sqlStmt := `SELECT u.user_id, u.user_name, p.joined_at, u.user_email, 
	FROM Users u
	JOIN Participants p ON u.user_id = p.user_id
	WHERE p.group_id = ?;`
	rows, err := db.Query(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return participants, err
	}
	defer rows.Close()
	for rows.Next() {
		var participant models.UserParticipant
		err := rows.Scan(&participant.UserID, &participant.UserName)
		if err != nil {
			log.Fatal(err)
		}
		participants = append(participants, participant)
	}
	return participants, nil
}

func GetParticipantsToDraw(db *sql.DB, groupID string) ([]models.Participant, error) {
	var participants []models.Participant
	sqlStmt := `SELECT p.group_id, p.user_id, p.joined_at
	FROM Participants p
	WHERE p.group_id = ? AND p.friend_user_id IS NULL;`
	rows, err := db.Query(sqlStmt, groupID)
	if err != nil {
		return participants, err
	}
	for rows.Next() {
		var participant models.Participant
		err := rows.Scan(&participant.GroupID, &participant.UserID, &participant.JoinedAt)
		if err != nil {
			return participants, err
		}
		participants = append(participants, participant)
	}
	return participants, nil
}
