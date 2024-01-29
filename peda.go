package peda

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aiteung/atapi"
	"github.com/aiteung/atmessage"
	"github.com/whatsauth/wa"
	"github.com/whatsauth/watoken"
)

func GCFHandler(MONGOCONNSTRINGENV, dbname, collectionname string) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	datagedung := GetAllBangunanLineString(mconn, collectionname)
	return GCFReturnStruct(datagedung)
}

func GCFPostHandler(PASETOPRIVATEKEYENV, MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	var Response Credential
	Response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var datauser User
	err := json.NewDecoder(r.Body).Decode(&datauser)
	if err != nil {
		Response.Message = "error parsing application/json: " + err.Error()
	} else {
		if IsPasswordValid(mconn, collectionname, datauser) {
			Response.Status = true
			tokenstring, err := watoken.Encode(datauser.Username, os.Getenv(PASETOPRIVATEKEYENV))
			if err != nil {
				Response.Message = "Gagal Encode Token : " + err.Error()
			} else {
				Response.Message = "Selamat Datang"
				Response.Token = tokenstring
			}
		} else {
			Response.Message = "Password Salah"
		}
	}

	return GCFReturnStruct(Response)
}

func GCFReturnStruct(DataStuct any) string {
	jsondata, _ := json.Marshal(DataStuct)
	return string(jsondata)
}

func CreateUser(token, mongoenv, dbname, collname string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(mongoenv, dbname)
	var datauser User
	err := json.NewDecoder(r.Body).Decode(&datauser)
	if err != nil {
		response.Message = "error parsing application/json: " + err.Error()
	} else {
		response.Status = true
		hash, hashErr := HashPassword(datauser.Password)
		if hashErr != nil {
			response.Message = "Gagal Hash Password" + err.Error()
		}
		InsertUserdata(mconn, collname, datauser.Username, datauser.Role, hash)
		// Prepare and send a WhatsApp message with registration details
		var username = datauser.Username
		var password = datauser.Password
		var nohp = datauser.No_whatsapp
		dt := &wa.TextMessage{
			To:       nohp,
			IsGroup:  false,
			Messages: "Selamat anda berhasil registrasi, berikut adalah username anda: " + username + "\nDan ini adalah password anda: " + password + "\nDisimpan baik baik ya",
		}
	
		// Make an API call to send WhatsApp message
		atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv(token), dt, "https://api.wa.my.id/api/send/message/text")
	
		response.Message = "Berhasil Input data"
	}
	return GCFReturnStruct(response)
}
