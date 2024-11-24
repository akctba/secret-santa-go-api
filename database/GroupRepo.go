package database

//this file will contain all the database operations for the Group model

import (
	"database/sql"
	"log"

	"github.com/akctba/secret-santa-go-api/models"
)

func InsertGroup(db *sql.DB, group models.Group) error {
	sqlStmt := `INSERT INTO Groups(group_id, name, date_created, date_draw, creator_user_id
	) VALUES (?, ?, ?, ?, ?);`
	_, err := db.Exec(sqlStmt, group.GroupID, group.Name, group.DateCreated, group.DateDraw, group.CreatorUserID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func GetGroupByID(db *sql.DB, id string) (models.Group, error) {
	var group models.Group
	sqlStmt := `SELECT group_id, name, date_created, date_draw, creator_user_id
	FROM Groups WHERE group_id = ?;`
	row := db.QueryRow(sqlStmt, id)
	err := row.Scan(&group.GroupID, &group.Name, &group.DateCreated, &group.DateDraw, &group.CreatorUserID)
	if err != nil {
		return group, err
	}
	return group, nil
}

func UpdateGroup(db *sql.DB, group models.Group) error {
	sqlStmt := `UPDATE Groups SET name = ?, date_created = ?, date_draw = ?, creator_user_id = ?
	WHERE group_id = ?;`
	_, err := db.Exec(sqlStmt, group.Name, group.DateCreated, group.DateDraw, group.CreatorUserID, group.GroupID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func DeleteGroup(db *sql.DB, id string) error {
	sqlStmt := `DELETE FROM Groups WHERE group_id = ?;`
	_, err := db.Exec(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func GetGroupsByUserID(db *sql.DB, id int) ([]models.Group, error) {
	var groups []models.Group
	sqlStmt := `SELECT group_id, name, date_created, date_draw, creator_user_id
	FROM Groups WHERE creator_user_id = ?;`
	rows, err := db.Query(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return groups, err
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.GroupID, &group.Name, &group.DateCreated, &group.DateDraw, &group.CreatorUserID)
		if err != nil {
			log.Fatal(err)
		}
		groups = append(groups, group)
	}
	return groups, nil
}
