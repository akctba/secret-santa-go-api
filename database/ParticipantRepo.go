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
	sqlStmt := `SELECT p.group_id, p.user_id, u.user_name, u.user_email, u.gender, u.date_of_birth, p.joined_at
	FROM Participants p
	JOIN Users u ON u.user_id = p.user_id
	WHERE p.user_id = ?;`
	rows, err := db.Query(sqlStmt, userID)
	if err != nil {
		return participants, err
	}
	defer rows.Close()
	for rows.Next() {
		var participant models.UserParticipant
		var dateOfBirthValue any
		var joinedAtValue any

		err := rows.Scan(&participant.GroupID, &participant.UserID, &participant.UserName,
			&participant.UserEmail, &participant.Gender, &dateOfBirthValue, &joinedAtValue)
		if err != nil {
			return participants, err
		}

		participant.DateOfBirth, err = parseDBTime(dateOfBirthValue)
		if err != nil {
			return participants, err
		}
		participant.JoinedAt, err = parseDBTime(joinedAtValue)
		if err != nil {
			return participants, err
		}

		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return participants, err
	}
	return participants, nil
}

func GetParticipantsByGroupID(db *sql.DB, id string) ([]models.UserParticipant, error) {
	var participants []models.UserParticipant
	sqlStmt := `SELECT p.group_id, u.user_id, u.user_name, u.user_email, u.gender, u.date_of_birth, p.joined_at
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
		var dateOfBirthValue any
		var joinedAtValue any

		err := rows.Scan(&participant.GroupID, &participant.UserID, &participant.UserName,
			&participant.UserEmail, &participant.Gender, &dateOfBirthValue, &joinedAtValue)
		if err != nil {
			return participants, err
		}

		participant.DateOfBirth, err = parseDBTime(dateOfBirthValue)
		if err != nil {
			return participants, err
		}
		participant.JoinedAt, err = parseDBTime(joinedAtValue)
		if err != nil {
			return participants, err
		}

		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return participants, err
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
	defer rows.Close()
	for rows.Next() {
		var participant models.Participant
		var joinedAtValue any

		err := rows.Scan(&participant.GroupID, &participant.UserID, &joinedAtValue)
		if err != nil {
			return participants, err
		}

		participant.JoinedAt, err = parseDBTime(joinedAtValue)
		if err != nil {
			return participants, err
		}

		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return participants, err
	}
	return participants, nil
}
