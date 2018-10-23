package models

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"os"
	u "server/utils"
)

/*
JWT claims struct
*/
type Token struct {
	UserId uint
	jwt.StandardClaims
}
//  Accounts []CustomizeAccount `gorm:"many2many:PersonAccount;foreignkey:idPerson;association_foreignkey:idAccount;association_jointable_foreignkey:account_id;jointable_foreignkey:person_id;"`
//a struct to rep user client
type Client struct {
	ID 			uint	 `json:"id"`
	Pseudo   	string   `json:"pseudo"`
	Password 	string   `json:"password"`
	Ips      	[]Ip     `json:"ips";gorm:"many2many:client_address"`
	Friends  	[]Client `json:"friends";gorm:"many2many:client_client;association_jointable_foreignkey:friend_id"`
	Token    	string   `json:"token";sql:"-"`
}

//Validate incoming user details...
func (client *Client) Validate() (map[string] interface{}, bool) {

	if len(client.Pseudo) < 6 {
		return u.Message(false, "Pseudo address is required"), false
	}

	if len(client.Password) < 6 {
		return u.Message(false, "Password is required and need at least 6 characters"), false
	}

	//Pseudo must be unique
	temp := &Client{}

	//check for errors and duplicate pseudos
	err := GetDB().Table("clients").Where("pseudo = ?", client.Pseudo).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Pseudo != "" {
		return u.Message(false, "Pseudo address already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true
}

func (client *Client) Create() (*Client, error) {

	if _, ok := client.Validate(); !ok {
		return nil, errors.New("The client is not valid")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(client.Password), bcrypt.DefaultCost)
	client.Password = string(hashedPassword)

	GetDB().Create(client)

	if client.ID <= 0 {

		return nil, errors.New("Failed to create client, connection error.")
	}

	//Create new JWT token for the newly registered client
	tk := &Token{UserId: client.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(hashedPassword))
	client.Token = tokenString
	//for test it sucked
	client.Password = "" //delete password

	return client, nil
}

func Login(pseudo string, password string) (map[string]interface{}) {

	client := &Client{}
	err := GetDB().Table("clients").Where("pseudo = ?", pseudo).First(client).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Pseudo address not found")
		}
		return u.Message(false, "Connection error. Please retry")
	}

	err = bcrypt.CompareHashAndPassword([]byte(client.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Worked! Logged In
	client.Password = ""

	//Create JWT token
	tk := &Token{UserId: client.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	client.Token = tokenString //Store the token in the response

	resp := u.Message(true, "Logged In")
	resp["client"] = client
	return resp
}

func GetClient(u uint) (*Client, error) {

	client := &Client{}
	err := GetDB().Table("clients").Where("id = ?", u).First(client).Error
	GetDB().Preload("Ips").First(&client)
	GetDB().Preload("Friends").First(&client)
	if err != nil {
		return nil, err
	}
	if client.Pseudo == "" { //User not found!
		return nil, errors.New("Pseudo is empty")
	}

	client.Password = ""
	return client, nil
}

func GetClientFromPseudo(friend *Client) (*Client, error) {
	pseudo := friend.Pseudo
	client := &Client{}
	err := GetDB().Table("clients").Where("pseudo = ?", pseudo).First(client).Error
	GetDB().Preload("Ips").First(&client)
	GetDB().Preload("Friends").First(&client)
	if err != nil {
		return nil, err
	}
	if client.Pseudo == "" { //User not found!
		return nil, errors.New("Pseudo is empty")
	}

	client.Password = ""
	return client, nil
}

func (client *Client) Update() (map[string] interface{})  {
	//check if ip has changed
	if len(client.Ips) > 0 {
		GetDB().Model(&client).Association("Ips").Replace(client.Ips)
	}

	GetDB().Save(&client)

	resp := u.Message(true, "Client updated")
	resp["client"] = client
	return resp
}

func (client *Client) Delete() (map[string] interface{}) {
	response := u.Message(true, "Client has been created")
	response["client"] = client
	return response
}

func (client *Client) AddFriend(friend Client) (map[string] interface{}){
	response := u.Message(true, "Client has been created")
	client.Friends = append(client.Friends, friend)
	client.Update()
	response["client"] = client
	return response
}

func (client *Client) RemoveFriend(friend *Client) {
	friends := client.Friends
	indice := -1
	for i:=0; i < len(friends); i ++ {
		if friend.ID == friends[i].ID {
			indice = i
		}
	}
	if indice == -1 {
		return
	}
	friends[indice] = friends[len(friends) - 1]
	friends = friends[:len(friends) - 1]
	client.Friends = friends
	client.Update()
}

