package controllers

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"server/models"
	u "server/utils"
	"strconv"
)

var UpdateClient = func(w http.ResponseWriter, r *http.Request) {

	client, successful := processPostClient(w, r)
	if !successful {
		return
	}
	log.Printf("client:%v", client)
	updatedClient, err := client.Update()
	if err != nil {
		log.Printf("Error updating client: %v", err)
	}
	u.RespondWithJSON(w, http.StatusOK, updatedClient)
}

func processPostClient(w http.ResponseWriter, r *http.Request) (*models.Client, bool) {
	client := &models.Client{}
	vars := mux.Vars(r)
	_, err := strconv.Atoi(vars["id"])
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return nil, false
	}
	err = json.NewDecoder(r.Body).Decode(client) //decode the request body into struct and failed if any error occur
	json.NewDecoder(r.Body).Decode(client)       //decode the request body into struct and failed if any error occur
	if err != nil {
		log.Print(err)
		u.Respond(w, u.Message(false, "Invalid request", http.StatusUnprocessableEntity))
		return nil, false
	}
	return client, true
}

var DeleteClient = func(w http.ResponseWriter, r *http.Request) {

	client := &models.Client{}
	err := json.NewDecoder(r.Body).Decode(client) //decode the request body into struct and failed if any error occur
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request", http.StatusUnprocessableEntity))
		return
	}

	resp := client.Delete()
	u.Respond(w, resp)
}

var GetClient = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	client, err := models.GetClient(uint(id))
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			u.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			u.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	u.RespondWithJSON(w, http.StatusOK, client)
}

var Logout = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	client, err := models.GetClient(uint(id))
	if err != nil {
		u.Message(false, "Stay with us !", http.StatusInternalServerError)
		return
	}
	client.Logout()
	u.Message(false, "Goodby", http.StatusOK)
}
