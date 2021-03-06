package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"server/models"
	"testing"
)

func TestLoginNonExistentUser(t *testing.T) {
	clearTable("clients")
	client := getSimpleClient()
	payload, err := json.Marshal(client)
	if err != nil {
		t.Errorf("error occurs when encoding client: %s", err.Error())
	}
	req, _ := http.NewRequest("POST", "/api/clients/login", bytes.NewBuffer(payload))
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, resp.Code)
}

func TestRegisterClients(t *testing.T) {
	clearTables()
	for i := 0; i < 3; i++ {
		client := addSimpleClient(t, "localhost")
		if client == nil {
			t.Error("Client is null")
			return
		}
		clientFromDb, err := models.GetClient(client.ID)
		if err != nil {
			t.Errorf("Error getting client; %s", err.Error())
			return
		}
		compareClient(client, clientFromDb, t)
		var bearer = "Bearer " + client.Token
		req, _ := http.NewRequest("PUT", "/api/clients/1/logout", nil)
		req.Header.Add("Authorization", bearer)
		resp := executeRequest(req)
		checkResponseCode(t, http.StatusOK, resp.Code)
	}
}


func TestLoginExistingClient(t *testing.T) {
	clearTables()
	//@TODO change test cause we send the adress of ip
	//clientRegisterd := addSimpleClient(t, "localhost")
	//client := useClient(t, clientRegisterd, "156.0.1.2")
	//clientFromDb, err := models.GetClient(client.ID)
	//if err != nil {
	//	t.Errorf("Error getting client; %s", err.Error())
	//	return
	//}
	//compareClient(client, clientFromDb, t)
}

func clientToBuffer(t *testing.T, client *models.Client) *bytes.Buffer {
	payload, err := json.Marshal(client)
	if err != nil {
		t.Errorf("error occurs when encoding client: %s", err.Error())
	}
	return bytes.NewBuffer(payload)
}

func TestLoginWrongCredential(t *testing.T) {
	clearTables()
	client := addSimpleClient(t, "localhost")
	c := &models.Client{
		Pseudo:   client.Pseudo,
		Password: "wrong_password",
	}
	req, _ := http.NewRequest("POST", "/api/clients/login", clientToBuffer(t, c))
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, resp.Code)
	var received string
	json.NewDecoder(resp.Body).Decode(received)
	if received != "" {
		t.Errorf("Client is not null:%v", received)
		return
	}
}

//func TestUpdateFriendApi(t *testing.T) {
//	clearTables()
//	updateClientSent := addSimpleClient(t, "localhost")
//	friend := addSimpleClient(t, "friend_ip")
//	var bearer = "Bearer " + updateClientSent.Token
//	updateClientSent.Friends = []*models.Client{
//		friend,
//	}
//	updateClientSent.Pseudo = "update_pseudo"
//	req, _ := http.NewRequest("PUT", "/api/clients/1/friends", clientToBuffer(t, friend))
//	req.Header.Add("Authorization", bearer)
//	resp := executeRequest(req)
//	if !checkResponseCode(t, http.StatusOK, resp.Code){
//		return
//	}
//	updateClientReceived := &models.Client{}
//	json.NewDecoder(resp.Body).Decode(updateClientReceived)
//	log.Printf("client received: %v", updateClientReceived)
//	compareClient(updateClientSent, updateClientReceived, t)
//	compareClientWithFriends(1, updateClientReceived, updateClientSent.Friends, t)
//}

func TestAddFriendApi(t *testing.T) {
	clearTables()
	updateClientSent := addSimpleClient(t, "localhost")
	friend := addSimpleClient(t, "friend_ip")
	friend = &models.Client{
		Pseudo: friend.Pseudo,
	}
	updateClientSent.Friends = []*models.Client{
		friend,
	}
	req, _ := http.NewRequest("POST", "/api/clients/1/friends", clientToBuffer(t, friend))
	var bearer = "Bearer " + updateClientSent.Token
	req.Header.Add("Authorization", bearer)
	resp := executeRequest(req)
	if !checkResponseCode(t, http.StatusOK, resp.Code) {
		return
	}
	updateClientReceived := &models.Client{}
	json.NewDecoder(resp.Body).Decode(updateClientReceived)
	test := &models.Client{}
	models.GetDB().Table("clients").Where("id = ?", updateClientReceived.ID).First(test)
	updateClientSent.Password = test.Password
	compareClient(updateClientSent, updateClientReceived, t)
	compareClientWithFriends(int(updateClientReceived.ID), updateClientReceived, updateClientReceived.Friends, t)
	client := &models.Client{}
	id := updateClientSent.ID
	models.GetDB().Table("clients").Where("id = ?", int(id)).First(client)
}

func TestLogoutApi(t *testing.T) {
	clearTables()
	newClient := addSimpleClient(t, "localhost")
	updateClientSent := useClient(t, newClient, "swaf")
	req, _ := http.NewRequest("PUT", "/api/clients/1/logout", nil)
	var bearer = "Bearer " + updateClientSent.Token
	req.Header.Add("Authorization", bearer)
	resp := executeRequest(req)
	if !checkResponseCode(t, http.StatusOK, resp.Code) {
		return
	}
	client := &models.Client{}
	models.GetDB().Table("clients").Where("id = ?", updateClientSent.ID).First(client)
}

func addSimpleClient(t *testing.T, ip string) *models.Client {
	req, _ := http.NewRequest("POST", "/api/clients/new", clientToBuffer(t, getSimpleClient()))
	req.RemoteAddr = ip
	resp := executeRequest(req)
	c := &models.Client{}
	json.NewDecoder(resp.Body).Decode(c)
	return c
}

func useClient(t *testing.T, client *models.Client, ip string) *models.Client {
	c := &models.Client{
		Pseudo:   client.Pseudo,
		Password: "test_password",
	}
	req, _ := http.NewRequest("POST", "/api/clients/login", clientToBuffer(t, c))
	req.RemoteAddr = ip
	resp := executeRequest(req)
	checkResponse(t, http.StatusOK, resp)
	json.NewDecoder(resp.Body).Decode(client)
	if client == nil {
		t.Error("Client is null")
		return nil
	}
	client.Ips = []*models.Ip{
		{
			Address: ip,
		},
	}
	return client
}
