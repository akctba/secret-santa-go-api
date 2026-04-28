package routes

import (
	"github.com/akctba/secret-santa-go-api/controllers"
	"github.com/gorilla/mux"
)

// Register attaches all application routes to the provided router.
func Register(r *mux.Router) {
	v1 := r.PathPrefix("/v1").Subrouter()

	// User endpoints
	v1.HandleFunc("/user", controllers.CreateUser).Methods("POST")
	v1.HandleFunc("/user/signin", controllers.Signin).Methods("POST")
	v1.HandleFunc("/user/refresh", controllers.RefreshToken).Methods("POST")
	v1.HandleFunc("/user/{id}", controllers.BearerAuth(controllers.GetUser)).Methods("GET")

	// Group endpoints
	v1.HandleFunc("/group", controllers.BearerAuth(controllers.CreateGroup)).Methods("POST")
	v1.HandleFunc("/group/{id}", controllers.BearerAuth(controllers.GetGroup)).Methods("GET")
	v1.HandleFunc("/group/{id}/participant", controllers.BearerAuth(controllers.AddParticipant)).Methods("POST")
	v1.HandleFunc("/group/{id}/draw", controllers.BearerAuth(controllers.RunDraw)).Methods("POST")
	v1.HandleFunc("/group/{id}/friend", controllers.BearerAuth(controllers.GetSecretFriend)).Methods("GET")
}
