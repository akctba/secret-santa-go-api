package routes

import (
	"github.com/akctba/secret-santa-go-api/controllers"
	"github.com/gorilla/mux"
)

// Register attaches all application routes to the provided router.
func Register(r *mux.Router) {
	// User endpoints
	r.HandleFunc("/user", controllers.CreateUser).Methods("POST")
	r.HandleFunc("/user/signin", controllers.Signin).Methods("POST")
	r.HandleFunc("/user/{id}", controllers.BearerAuth(controllers.GetUser)).Methods("GET")

	// Group endpoints
	r.HandleFunc("/group", controllers.BearerAuth(controllers.CreateGroup)).Methods("POST")
	r.HandleFunc("/group/{id}", controllers.BearerAuth(controllers.GetGroup)).Methods("GET")
	r.HandleFunc("/group/{id}/participant", controllers.BearerAuth(controllers.AddParticipant)).Methods("POST")
	r.HandleFunc("/group/{id}/draw", controllers.BearerAuth(controllers.RunDraw)).Methods("POST")
	r.HandleFunc("/group/{id}/friend", controllers.BearerAuth(controllers.GetSecretFriend)).Methods("GET")
}
