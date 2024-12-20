package models

import "time"

type Group struct {
	GroupID       string    `json:"group_id"`
	Name          string    `json:"name"`
	DateCreated   time.Time `json:"date_created"`
	DateDraw      time.Time `json:"date_draw"`
	CreatorUserID int       `json:"creator_user_id"`
}

type UserSignin struct {
	UserEmail string `json:"email"`
	Password  string `json:"password"`
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

type ParticipantRequest struct {
	GroupID string `json:"group_id"`
	UserID  int    `json:"user_id"`
}

type UserParticipant struct {
	UserID      int       `json:"user_id"`
	GroupID     string    `json:"group_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
	JoinedAt    time.Time `json:"joined_at"`
}
