package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"server/app"
	"server/models"
	"testing"
)

var a app.Application
var NbClient int

func TestMain(m *testing.M) {
	log.Print("Running test !")
	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	a = app.Application{}
	a.Initialize()

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("test_db_name")
	dbHost := os.Getenv("db_host")

	models.Open(username, password, dbName, dbHost)
	if !ensureTableExists() {
		fmt.Print("La db n'existe pas ")
		panic(m)
	}
	NbClient = 1
	code := m.Run()

	dropTables()
	os.Exit(code)
}

func ensureTableExists() bool {
	client := models.GetDB().HasTable(&models.Client{})
	address := models.GetDB().HasTable(&models.Ip{})
	clientClient := models.GetDB().HasTable(&models.Friendship{})
	clientAddress := models.GetDB().HasTable("client_address")
	return client && address && clientClient && clientAddress
}
func dropTables() {
	models.GetDB().DropTable(&models.Client{}, &models.Ip{}, &models.Friendship{}, "client_client", "client_address")
}

func checkResponseCode(t *testing.T, expected int, actual int) bool {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
		return false
	}
	return true
}

func checkResponse(t *testing.T, expected int, response *httptest.ResponseRecorder) bool {
	if !(checkResponseCode(t, expected, response.Code)) {
		t.Errorf("Error message %s", response.Header().Get("message"))
		t.Errorf("Error status %s", response.Header().Get("status"))
		return false
	}
	return true
}

func executeRequest(request *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, request)
	return rr
}

func clearTable(table string) {
	deletion := fmt.Sprintf("DELETE FROM %s", table)
	updateId := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table)
	models.GetDB().Exec(deletion)
	models.GetDB().Exec(updateId)
}

func clearTables() {
	clearTable("clients")
	clearTable("client_address")
	clearTable("friendships")
	clearTable("ips")
}

func getSimpleClient() *models.Client{
	pseudo := fmt.Sprintf("test_client_%d", NbClient)
	client := &models.Client{
		Pseudo: pseudo,
		Password: "test_password",
	}
	NbClient ++
	return client
}
func compareClientWithDb(t * testing.T, client *models.Client, password bool) {
	if client == nil {
		t.Error("The client is empty")
	}
	clientFromDB := &models.Client{}
	err := models.GetDB().Table("clients").Where("id = ?", client.ID).First(clientFromDB).Error
	if err != nil {
		t.Errorf("Error getting the client:%d", client.ID)
	}
	if clientFromDB == nil {
		t.Error("Client empty")
		return
	}
	if clientFromDB.Pseudo != client.Pseudo {
		t.Errorf("The pseudo expected was '%s', got '%s'", client.Pseudo, clientFromDB.Pseudo)
	}

	if lenCliDb:=len(clientFromDB.Ips); lenCliDb != len(client.Ips) {
		t.Errorf("Not the same amount of ips. Expected : '%d', got: '%d", len(client.Ips), lenCliDb)
		t.Errorf("for the client :%d", client.ID)
		return
	}
	for i:=0; i < len(clientFromDB.Ips); i++ {
		if tmp := clientFromDB.Ips[i].Address; tmp != client.Ips[i].Address {
			t.Errorf("Not the same address at index %d. Expected: '%s', got '%s'", i, client.Ips[i].Address, tmp)
		}
	}
	if !password {
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(clientFromDB.Password), []byte(client.Password))
	if err != nil { //Password does not match!
		t.Errorf("The password expected was '%s', got '%s'", clientFromDB.Password, client.Password)
		t.Errorf("Error: %v", err)
	}

}
func compareClient(client *models.Client, clientFromDB *models.Client, t *testing.T) {
	if clientFromDB == nil {
		t.Error("Client empty")
		return
	}
	if clientFromDB.Pseudo != client.Pseudo {
		t.Errorf("The pseudo expected was '%s', got '%s'", client.Pseudo, clientFromDB.Pseudo)
	}
	//err := bcrypt.CompareHashAndPassword([]byte(clientFromDB.Password), []byte(client.Password))
	//if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
	//	t.Errorf("The password expected was '%s', got '%s'", clientFromDB.Password, client.Password)
	//}
	if lenCliDb:=len(clientFromDB.Ips); lenCliDb != len(client.Ips) {
		t.Errorf("Not the same amount of ips. Expected : '%d', got: '%d", len(client.Ips), lenCliDb)
		t.Errorf("for the client :%d", client.ID)
		return
	}
	for i:=0; i < len(clientFromDB.Ips); i++ {
		if tmp := clientFromDB.Ips[i].Address; tmp != client.Ips[i].Address {
			t.Errorf("Not the same address at index %d. Expected: '%s', got '%s'", i, client.Ips[i].Address, tmp)
		}
	}
}

func compareClientWithFriends(idClient int, client *models.Client, friends []*models.Client, t *testing.T) {
	dbClient,_ := models.GetClient(uint(idClient))
	//var clients []models.Client
	//models.GetDB().Find(&clients)
	if dbClient == nil {
		t.Error("Client is empty")
		return
	}
	compareClient(client, dbClient, t)
	dbFriends := dbClient.Friends
	if l := len(dbFriends); l != len(friends){
		t.Errorf("Client is supposed to have '%v' friends, instead had '%v'", len(friends), l)
		return
	}
	for i := 0; i < len(friends); i++ {
		compareClient(friends[i], dbFriends[i], t)
	}
}