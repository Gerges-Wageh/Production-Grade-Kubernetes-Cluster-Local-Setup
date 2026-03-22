package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"posts/internal/db"

	"github.com/gorilla/mux"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post db.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := db.InsertPost(&post); err != nil {
		http.Error(w, "failed to insert post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	post, err := db.GetPostByID(id)
	if err != nil {
		http.Error(w, "failed to get post", http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.NotFound(w, r)
		return
	}

	json.NewEncoder(w).Encode(post)
}
