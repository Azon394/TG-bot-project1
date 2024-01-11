package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type clients struct { // Структура данных пользователя
	Id       string `json:"Id"`
	Username string `json:"Username"`
	Tg_id    string `json:"Tg_id"`
	Rights   right  `json:"Rights"`
	Group    string `json:"Group"`
	Subgroup string `json:"Subgroup"`
}

type right struct { // Структура данных прав пользователя
	Read  bool
	Write bool
	Admin bool
}

type userData struct { // Данные полученнные после авторизации через гитхаб
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Глобальная переменная для проверки, что пользователь дал доступ
var authenticating struct {
	is_done bool
	code    string
}

const (
	CLIENT_IDS     = "b5c2620b90b77b0de4f9"
	CLIENT_SECRETS = "f5a755268f590eeef28eeff9c70dd45e9acff0c2"
	JWTCODE        = "7777777"
	MYIP           = "26.251.53.172"
	KIRILLIP       = "26.233.179.99"
)

func addClient(id, username, tg_id, group, subgroup string, rights right) {
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
	// создаём переменную в виде структуры clients
	current_client := clients{id, username, tg_id, rights, group, subgroup}
	// добавляем одиночный документ в коллекцию
	insertResult, err := collection.InsertOne(context.TODO(), current_client)
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	// выводим внутренний ID добавленного документа
	log.Println("Inserted a single document: ", insertResult.InsertedID)
} // Функция добавления данных нового клиента по умолчанию в бд

func findClient(f, id string) clients { // возвращает данные о пользоваете в виде структуры
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017")) // создаём дэфолтного клиента
	if err != nil {                                                                        // проверяем ошибку если она есть
		var norights right
		log.Println(err)
		return clients{"", "", "", norights, "", ""}
	}
	// создаём соединение
	err = client.Connect(context.TODO())
	if err != nil { // проверяем ошибку если она есть
		var norights right
		log.Println(err)
		return clients{"", "", "", norights, "", ""}
	}
	// проверяем соединение
	err = client.Ping(context.TODO(), nil)
	if err != nil { // проверяем ошибку если она есть
		var norights right
		log.Println(err)
		return clients{"", "", "", norights, "", ""}
	}
	// обращаемся к коллекции clients из базы tg
	collection := client.Database("tg").Collection("clients")
	// создаём фильтр по которму мы будем искать клиента. был взят именно ID потому что они не повторяются
	filter := bson.D{{f, id}}
	// создаём переменную в которую будем записывать полученного клиента в результате поиска
	var result clients
	// собственно ищем
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil { // проверяем ошибку если она есть то возвращаем пустую структуру вида clients
		var norights right
		log.Println(err)
		return clients{"", "", "", norights, "", ""}
	}
	log.Println("Client was found")
	return result // возвращаем в виде структуры clients
} // Функция для нахождения клиента в бд по его GIT_ID или CHAT_ID

func updateData(Tg_id, key, value string) (string, error) {
	// создаём дэфолтного клиента
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil { // проверяем ошибку если она есть
		return "", err
	}
	// создаём соединение
	err = client.Connect(context.TODO())
	if err != nil { // проверяем ошибку если она есть
		return "", err
	}
	// проверяем соединение
	err = client.Ping(context.TODO(), nil)
	if err != nil { // проверяем ошибку если она есть
		return "", err
	}
	// обращаемся к коллекции clients из базы tg
	collection := client.Database("tg").Collection("clients")
	filter := bson.D{{"tg_id", Tg_id}}
	update := bson.D{
		{"$set", bson.D{
			{key, value},
		}},
	}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return "", err
	}
	log.Println("Client data was updated")
	return "true", nil
} // Функция для заполнения недостающих данных о пользователе

func updateClient(gitid, key, value, Other1, Other2 string) {
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
	filter := bson.D{{"id", gitid}}
	var update bson.D
	if key == "read" || key == "write" || key == "admin" {
		prava := []string{"read", "write", "admin"}
		for index, val := range prava {
			if val == key {
				prava = append(prava[:index], prava[index+1:]...)
				break
			}
		}
		if value == "false" {
			upda := bson.D{
				{"$set", bson.D{
					{"rights", bson.D{
						{key, false}, {prava[0], to_boolean(Other1)}, {prava[1], to_boolean(Other2)},
					}},
				},
				},
			}
			update = upda
		} else if value == "true" {
			upda := bson.D{
				{"$set", bson.D{
					{"rights", bson.D{
						{key, true}, {prava[0], to_boolean(Other1)}, {prava[1], to_boolean(Other2)},
					},
					}},
				},
			}
			update = upda
		}
	} else {
		upda := bson.D{
			{"$set", bson.D{
				{key, value},
			},
			},
		}
		update = upda
	}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Client was updated")
} // Функция для изменения данных о ползователе через модуль администрирования

func is_in_data(code string) bool { // Проверка существует пользоваетль с данным ID в бд
	curUser := findClient("id", code)
	if curUser.Id == "" {
		return false
	}
	return true
}

func to_boolean(s string) bool {
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	}
	return false
}

func main() {
	http.HandleFunc("/reg", reghandler)
	http.HandleFunc("/add", addhandler)
	http.HandleFunc("/find", findhandler)
	http.HandleFunc("/oauth", handleoauth)
	http.HandleFunc("/update", updatehandler)
	http.HandleFunc("/to_admin", adminhandler)
	http.HandleFunc("/getJWT/admin", giveJWThandler)
	http.HandleFunc("/checkAbout", checkAbouthandler)
	http.HandleFunc("/updateData", updateDatahandler)
	http.HandleFunc("/getJWT/schedule", giveJWTShedulehandler)
	http.HandleFunc("/getJWT/comment", giveJWTSheduleCommenthandler)
	http.HandleFunc("/getJWT/studloc", giveJWTSheduleStudlochandler)
	http.HandleFunc("/getJWT/prepodloc", giveJWTShedulePrepodlochandler)
	http.ListenAndServe(":8081", nil)
}

func giveJWTSheduleCommenthandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	lesson_number := r.FormValue("lesson_number")
	main_group := r.FormValue("main_group")
	sub_group := r.FormValue("sub_group")
	oddevenweek := r.FormValue("oddevenweek")
	comment := r.FormValue("comment")
	weekday := r.FormValue("weekday")
	token := zacodeSheduleComment(action, lesson_number, main_group, sub_group, oddevenweek, comment, weekday)
	fmt.Fprint(w, token)
}

func giveJWTSheduleStudlochandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	main_group := r.FormValue("main_group")
	sub_group := r.FormValue("sub_group")
	token := zacodeSheduleStudloc(action, main_group, sub_group)
	fmt.Fprint(w, token)
}

func giveJWTShedulePrepodlochandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	prepod := r.FormValue("prepod")
	token := zacodeShedulePrepodloc(action, prepod)
	fmt.Fprint(w, token)
}

func updateDatahandler(w http.ResponseWriter, r *http.Request) {
	Tg_id := r.FormValue("chatid")
	data := r.FormValue("data")
	datatype := r.FormValue("datatype")
	chk, err := updateData(Tg_id, datatype, data)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(w, chk)
	}
}

func giveJWTShedulehandler(w http.ResponseWriter, r *http.Request) {
	GIT_ID := r.FormValue("gitid")
	action := r.FormValue("action")
	person := findClient("id", GIT_ID)
	token := zacodeShedule(person, action)
	fmt.Fprint(w, token)
}

func giveJWThandler(w http.ResponseWriter, r *http.Request) {
	GIT_ID := r.FormValue("gitid")
	person := findClient("id", GIT_ID)
	log.Println("Админ найден")
	token := zacode(person)
	fmt.Fprint(w, token)
}

func checkAbouthandler(w http.ResponseWriter, r *http.Request) {
	Tg_id := r.URL.Query().Get("chatid")
	person := findClient("tg_id", Tg_id)
	if person.Username != "" {
		fmt.Fprint(w, "true")
	} else {
		fmt.Fprint(w, "false")
	}
}

func adminhandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("jwt")
	fmt.Fprintf(w, "http://"+MYIP+":8080/check?Token=%s", token)
} // проверяет есть у пользователя права для перехода в модуль администрирования, если есть то отправляет jwt

func updatehandler(w http.ResponseWriter, r *http.Request) {
	Id := r.URL.Query().Get("Id")
	Key := r.URL.Query().Get("Key")
	Value := r.URL.Query().Get("Value")
	Other1 := r.URL.Query().Get("Other1")
	Other2 := r.URL.Query().Get("Other2")
	token := r.URL.Query().Get("token")
	log.Println(Id, Key, Value, Other1, Other2, token)
	updateClient(Id, Key, Value, Other1, Other2)
	http.Redirect(w, r, "http://"+MYIP+":8080/home?token="+token, http.StatusSeeOther)
} // обновляет данные о пользователе по айди

func reghandler(w http.ResponseWriter, r *http.Request) {
	Id := r.URL.Query().Get("Id")
	var authURL string = "https://github.com/login/oauth/authorize?client_id=" + CLIENT_IDS + "&state=" + Id
	log.Println("send link")
	fmt.Fprint(w, authURL)
} // возвращает ссылку для регистрации через гитхаб

func addhandler(w http.ResponseWriter, r *http.Request) {
	Id := r.URL.Query().Get("Id")
	Username := r.URL.Query().Get("Username")
	Rights := r.URL.Query().Get("Rights")
	Tg_id := r.URL.Query().Get("Tg_id")
	Group := r.URL.Query().Get("Group")
	Subgroup := r.URL.Query().Get("Subgroup")
	var Rightslist right
	a1, _ := strconv.Atoi(strings.Split(Rights, "")[0])
	a2, _ := strconv.Atoi(strings.Split(Rights, "")[1])
	a3, _ := strconv.Atoi(strings.Split(Rights, "")[2])
	Rightslist.Read = a1 != 0
	Rightslist.Write = a2 != 0
	Rightslist.Admin = a3 != 0
	addClient(Id, Username, Tg_id, Group, Subgroup, Rightslist)
} // добавляет в БД пользователя со значениями по умолчанию

func findhandler(w http.ResponseWriter, r *http.Request) {
	Tg_id := r.URL.Query().Get("Tg_id")
	user := findClient("tg_id", Tg_id)
	token := zacode(user)
	fmt.Fprint(w, token)
} // возвращает jwt токен с инфой о пользователе

func zacodeShedule(client clients, action string) string {
	log.Println(client, action)
	tokeExpiresAt := time.Now().Add(time.Minute * time.Duration(5))
	user := jwt.MapClaims{
		"main_group": client.Group,
		"sub_group":  client.Subgroup,
		"action":     action,
		"expires_at": int(tokeExpiresAt.Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenString, err := token.SignedString([]byte(JWTCODE))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	log.Println(tokenString)
	return tokenString
}

func zacodeSheduleComment(action, lesson_number, main_group, sub_group, oddevenweek, comment, weekday string) string {
	log.Println(action, lesson_number, main_group, sub_group, oddevenweek, comment, weekday)
	tokeExpiresAt := time.Now().Add(time.Minute * time.Duration(5))
	user := jwt.MapClaims{
		"action":        action,
		"lesson_number": lesson_number,
		"main_group":    main_group,
		"sub_group":     sub_group,
		"oddevenweek":   oddevenweek,
		"comment":       comment,
		"weekday":       weekday,
		"expires_at":    int(tokeExpiresAt.Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenString, err := token.SignedString([]byte(JWTCODE))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	return tokenString
}

func zacodeSheduleStudloc(action, main_group, sub_group string) string {
	log.Println(action, main_group, sub_group)
	tokeExpiresAt := time.Now().Add(time.Minute * time.Duration(5))
	user := jwt.MapClaims{
		"action":     action,
		"main_group": main_group,
		"sub_group":  sub_group,
		"expires_at": int(tokeExpiresAt.Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenString, err := token.SignedString([]byte(JWTCODE))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	return tokenString
}

func zacodeShedulePrepodloc(action, prepod string) string {
	log.Println(action, prepod)
	tokeExpiresAt := time.Now().Add(time.Minute * time.Duration(5))
	user := jwt.MapClaims{
		"action":     action,
		"prepod":     prepod,
		"expires_at": int(tokeExpiresAt.Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenString, err := token.SignedString([]byte(JWTCODE))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	log.Println(tokenString)
	return tokenString
}

func zacode(client clients) string {
	tokeExpiresAt := time.Now().Add(time.Minute * time.Duration(5))
	user := jwt.MapClaims{
		"name":       client.Username,
		"id":         client.Id,
		"rights":     client.Rights,
		"tg_id":      client.Tg_id,
		"group":      client.Group,
		"subgroup":   client.Subgroup,
		"expires_at": tokeExpiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenString, err := token.SignedString([]byte(JWTCODE))
	if err != nil { // проверяем ошибку если она есть
		log.Fatal(err)
	}
	return tokenString
} // возвращает jwt токен с иинфой о пользователе и временем жизни токена

func handleoauth(w http.ResponseWriter, r *http.Request) {
	var responseHtml = "<html><body><h1>Вы НЕ аутентифицированы!</h1></body></html>"
	code := r.URL.Query().Get("code") // Достаем временный код из запроса
	tg_id := r.URL.Query().Get("state")
	if code != "" {
		authenticating.is_done = true
		authenticating.code = code
		responseHtml = "<html><body><h1>Вы аутентифицированы!</h1></body></html>"
	}
	accessToken := getAccessToken(authenticating.code)
	userinfo := getuserData(accessToken)
	fmt.Fprint(w, responseHtml) // Ответ на запрос
	if is_in_data(strconv.Itoa(userinfo.Id)) {
		fmt.Fprint(w, "<html><body><h1>Уже зарегестрированы</h1></body></html>")
	} else if userinfo.Id != 0 {
		var r right
		r.Read = true
		r.Write = false
		r.Admin = false
		addClient(strconv.Itoa(userinfo.Id), userinfo.Name, tg_id, "", "", r)
		fmt.Fprint(w, "<html><body><h1>Успешно зарегестрированы</h1></body></html>")
	}
	client := http.Client{}

	// Формируем строку запроса вместе с query string
	requestURL := fmt.Sprintf("http://"+KIRILLIP+":6969/gitid?githubid=%s&chatid=%s", strconv.Itoa(userinfo.Id), tg_id)

	// Выполняем запрос на сервер. Ответ попадёт в переменную response
	request, _ := http.NewRequest("GET", requestURL, nil)
	response, _ := client.Do(request)
	defer response.Body.Close()

} // регестрирует пользователя через гитхаб

func getAccessToken(code string) string {
	// Создаём http-клиент с дефолтными настройками
	client := http.Client{}
	requestURL := "https://github.com/login/oauth/access_token"

	// Добавляем данные в виде Формы
	form := url.Values{}
	form.Add("client_id", CLIENT_IDS)
	form.Add("client_secret", CLIENT_SECRETS)
	form.Add("code", code)

	// Готовим и отправляем запрос
	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Accept", "application/json") // просим прислать ответ в формате json
	response, _ := client.Do(request)
	defer response.Body.Close()

	// Достаём данные из тела ответа
	var responsejson struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(response.Body).Decode(&responsejson)
	return responsejson.AccessToken
}

func getuserData(AccessToken string) userData {
	// Создаём http-клиент с дефолтными настройками
	client := http.Client{}
	requestURL := "https://api.github.com/user"

	// Готовим и отправляем запрос
	request, _ := http.NewRequest("GET", requestURL, nil)
	request.Header.Set("Authorization", "Bearer "+AccessToken)
	response, _ := client.Do(request)
	defer response.Body.Close()

	var data userData
	json.NewDecoder(response.Body).Decode(&data)
	return data
}
