package database

import (
	"database/sql"
	"log"

	"github.com/akctba/secret-santa-go-api/models"
)

//this file will contain all the database operations for the User model

func InsertUser(db *sql.DB, user models.User) error {
	sqlStmt := `INSERT INTO Users(user_name, user_email, password

	) VALUES (?, ?, ?);`
	_, err := db.Exec(sqlStmt, user.UserName, user.UserEmail, user.Password)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User
	sqlStmt := `SELECT user_id, user_name, user_email, password
	FROM Users WHERE user_email = ?;`
	row := db.QueryRow(sqlStmt, email)
	err := row.Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.Password)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByID(db *sql.DB, id int) (models.User, error) {
	var user models.User
	sqlStmt := `SELECT user_id, user_name, user_email, password
	FROM Users WHERE user_id = ?;`
	row := db.QueryRow(sqlStmt, id)
	err := row.Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.Password)
	if err != nil {
		return user, err
	}
	return user, nil
}

func UpdateUser(db *sql.DB, user models.User) error {
	sqlStmt := `UPDATE Users SET user_name = ?, user_email = ?, password = ?
	WHERE user_id = ?;`
	_, err := db.Exec(sqlStmt, user.UserName, user.UserEmail, user.Password, user.UserID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func DeleteUser(db *sql.DB, id int) error {
	sqlStmt := `DELETE FROM Users WHERE user_id = ?;`
	_, err := db.Exec(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func GetAllUsers(db *sql.DB) ([]models.User, error) {
	var users []models.User
	sqlStmt := `SELECT user_id, user_name, user_email, password FROM Users;`
	rows, err := db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.Password)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserGroups(db *sql.DB, id int) ([]models.Group, error) {
	var groups []models.Group
	sqlStmt := `SELECT g.group_id, g.group_name, g.date_created, g.date_draw, g.creator_user_id
	FROM Groups g
	JOIN Participants p ON g.group_id = p.group_id
	WHERE p.user_id = ?;`
	rows, err := db.Query(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return groups, err
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		err = rows.Scan(&group.GroupID, &group.Name, &group.DateCreated, &group.DateDraw, &group.CreatorUserID)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return groups, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func GetUserFriend(db *sql.DB, userId int, groupId int) ([]models.User, error) {
	var users []models.User
	sqlStmt := `SELECT u.user_id, u.user_name, u.user_email, u.password
	FROM Users u
	JOIN Participants p ON u.user_id = p.friend_user_id
	WHERE p.user_id = ? AND p.group_id = ?;`
	rows, err := db.Query(sqlStmt, userId, groupId)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.Password)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserParticipant(db *sql.DB, userId int, groupId int) (models.Participant, error) {
	var participant models.Participant
	sqlStmt := `SELECT group_id, user_id, joined_at, friend_user_id
	FROM Participants WHERE user_id = ? AND group_id = ?;`
	row := db.QueryRow(sqlStmt, userId, groupId)
	err := row.Scan(&participant.GroupID, &participant.UserID, &participant.JoinedAt, &participant.FriendUserID)
	if err != nil {
		return participant, err
	}
	return participant, nil
}

func GetGroupParticipants(db *sql.DB, groupId int) ([]models.UserParticipant, error) {
	var participants []models.UserParticipant
	sqlStmt := `SELECT p.group_id, p.user_id, u.user_name, u.user_email, u.gender, u.date_of_birth, p.joined_at
	FROM Participants p
	JOIN Users u ON p.user_id = u.user_id
	WHERE p.group_id = ?;`
	rows, err := db.Query(sqlStmt, groupId)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return participants, err
	}
	defer rows.Close()

	for rows.Next() {
		var participant models.UserParticipant
		err = rows.Scan(&participant.GroupID, &participant.UserID, &participant.UserName,
			&participant.UserEmail, &participant.Gender, &participant.DateOfBirth,
			&participant.JoinedAt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return participants, err
		}
		participants = append(participants, participant)
	}
	return participants, nil
}
