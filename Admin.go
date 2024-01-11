package main

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func to_bool(s string) bool {
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	}
	return false
}

type client_type struct {
	Id       string
	Username string
	Tg_id    string
	Rights   right_ad
	Token    string
}
type right_ad struct {
	Read  bool
	Write bool
	Admin bool
}

// Генерируем случайную строку - куки
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type sessionData map[string]string

func allUsers(str string) []client_type {
	// создаём дэфолтного клиента
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	// создаём соединение
	err = client.Connect(context.TODO())
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	// проверяем соединение
	err = client.Ping(context.TODO(), nil)
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	// обращаемся к коллекции clients из базы tg
	collection := client.Database("tg").Collection("clients")

	// Pass these options to the Find method
	options := options.Find()
	filter := bson.M{}

	// Here's an array in which you can store the decoded documents
	var results []client_type

	// Passing nil as the filter matches all documents in the collection
	cur, err := collection.Find(context.TODO(), filter, options)
	if err != nil {
		log.Fatal(err)
	}
	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var elem client_type
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		elem.Token = str
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	return results
}

var sessions = make(map[string]sessionData)

func main() {
	http.HandleFunc("/err", err)
	http.HandleFunc("/check", check)
	http.HandleFunc("/vibor", vibor)
	http.HandleFunc("/change", change)
	http.HandleFunc("/home", home_page)
	http.HandleFunc("/parsOk", successPars)
	http.HandleFunc("/change_rasp", change_rasp)
	http.ListenAndServe(":8080", nil)
}

func err(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	tmpl, _ := template.ParseFiles("templates/err.html")
	tmpl.Execute(w, token)
}

func successPars(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/successPars.html")
	tmpl.Execute(w, "")
}

func home_page(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if sessions[token]["username"] != "" && to_int(sessions[token]["expires_at"]) > int(time.Now().Unix()) {
		tmpl, _ := template.ParseFiles("templates/home.html")
		tmpl.Execute(w, token)
	} else {
		tmpl, _ := template.ParseFiles("templates/err.html")
		tmpl.Execute(w, token)
	}
}

func change_rasp(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if sessions[token]["username"] != "" && to_int(sessions[token]["expires_at"]) > int(time.Now().Unix()) {
		tmpl, _ := template.ParseFiles("templates/rasp.html")
		tmpl.Execute(w, "")
	} else {
		tmpl, _ := template.ParseFiles("templates/err.html")
		tmpl.Execute(w, token)
	}
}

func check(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("Token")
	if token != "" {
		info := decodeValid(token)
		var expiration float64 = (info["expires_at"]).(float64)
		if expiration >= float64(time.Now().Unix()) {
			rights := info["rights"].(map[string]interface{})
			admin := rights["Admin"].(bool)
			if admin == true {
				cookie := http.Cookie{
					Name:     "user_cookie",
					Value:    randString(16),
					Path:     "/",
					MaxAge:   3600,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				}
				sessions[cookie.Value] = make(sessionData)
				sessions[cookie.Value]["username"] = info["name"].(string)
				sessions[cookie.Value]["expires_at"] = strconv.FormatInt(time.Now().Add(time.Minute*time.Duration(15)).Unix(), 10)
				// Добавляем куки в ответ
				http.SetCookie(w, &cookie)
				http.Redirect(w, r, "/home?token="+cookie.Value, http.StatusSeeOther)
			} else {
				fmt.Fprint(w, "not admin")
			}
		} else {
			fmt.Fprint(w, "expiration is not ok")
		}
	} else {
		tmpl, _ := template.ParseFiles("templates/err.html")
		tmpl.Execute(w, token)
	}
}

func to_int(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func change(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if sessions[token]["username"] != "" && to_int(sessions[token]["expires_at"]) > int(time.Now().Unix()) {
		res := allUsers(token)
		tmpl, _ := template.ParseFiles("templates/page.html")
		tmpl.Execute(w, res)
	} else {
		tmpl, _ := template.ParseFiles("templates/err.html")
		tmpl.Execute(w, token)
	}
}

func vibor(w http.ResponseWriter, r *http.Request) {
	var rights right_ad
	var cli client_type
	rights.Read = to_bool(r.URL.Query().Get("Read"))
	rights.Write = to_bool(r.URL.Query().Get("Write"))
	rights.Admin = to_bool(r.URL.Query().Get("Admin"))
	cli.Id = r.URL.Query().Get("Id")
	cli.Token = r.URL.Query().Get("token")
	cli.Rights = rights
	tmpl, _ := template.ParseFiles("templates/vibor.html")
	tmpl.Execute(w, cli)
}

func decodeValid(tokenString string) jwt.MapClaims {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("7777777"), nil
	})
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	return claims
}

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpZXJzX2F0IjoxNzAxNzAxMzU4LCJpZCI6IjEyMzMyMSIsIm5hbWUiOiJQdXBhIiwicmlnaHRzIjp7IlJlYWQiOnRydWUsIldyaXRlIjpmYWxzZSwiQWRtaW4iOnRydWV9LCJ0Z19pZCI6IjExMjIzMyJ9._749wORYv1Bi5XzNZAofSQRN2BHsplmymM1oq0k66XM
