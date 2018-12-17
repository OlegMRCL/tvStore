package main

import (
	"net/http"
	"log"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
	"database/sql"
	"io/ioutil"
	"time"
	"io"
	"bytes"
)

type Answer struct {
	Ok     bool
	Err    error
	Tv     Tv
	Fields Validated
}

//(GET /tv)
//Выдает из БД список всех телевизоров
func All(w http.ResponseWriter, r *http.Request) {

	app := getAppInstance()
	logRequest(r)
	w.Header().Set("Content-Type", "application/json")

	rows, err := app.db.Query("SELECT * FROM productdb.tv")
	fmt.Println("mysql: SELECT * FROM productdb.tv")
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
	defer rows.Close()
	tvs := []Tv{}

	for rows.Next(){
		tv := Tv{}
		err := rows.Scan(&tv.Id, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)
		if err != nil{
			fmt.Println(err)
			continue
		}
		tvs = append(tvs, tv)
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(tvs)
}

//(GET /tv/{id})
//Выдает 1 телевизор из БД по id
func Get(w http.ResponseWriter, r *http.Request) {
	app := getAppInstance()
	logRequest(r)
	w.Header().Set("Content-Type", "application/json")

	answer := Answer {}

	params := mux.Vars(r)
	id := params ["id"]
	rows := app.db.QueryRow("SELECT * FROM productdb.tv WHERE id = ?", id)
	fmt.Println("mysql: SELECT * FROM productdb.tv WHERE id = " + id)

	var tv Tv
	err := rows.Scan(&tv.Id, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)

	if err == sql.ErrNoRows {
		w.WriteHeader(404)
		http.NotFound(w, r)
	} else if err != nil {
		w.WriteHeader(500)
		log.Println(err)
		answer.Err = err
	}else{
		w.WriteHeader(200)
		answer.Ok = true
		answer.Tv = tv
		_, v := validate(tv)
		answer.Fields = v
	}

	json.NewEncoder(w).Encode(answer)
}

//(POST /tv)
//Добавляет в БД 1 телевизор
func Add(w http.ResponseWriter, r *http.Request) {

	app := getAppInstance()
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")
	answer := Answer{}

	var tv Tv

	err := json.NewDecoder(r.Body).Decode(&tv)
	answer.Tv = tv
	defer r.Body.Close()

	if (err != nil)&&(err != io.EOF) {
		w.WriteHeader(500)
		log.Println(err)
		answer.Err = err

	}else if ok, v := validate(tv);	ok{
		answer.Fields = v
		_, err := app.db.Exec("INSERT INTO productdb.tv (brand, manufacturer, model, year) VALUES (?, ?, ?, ?)",
			tv.Brand, tv.Manufacturer, tv.Model, tv.Year)
		fmt.Printf("mysql: INSERT INTO productdb.tv (brand, manufacturer, model, year) VALUES (%v, %v, %v, %v)\n",
			tv.Brand, tv.Manufacturer, tv.Model, tv.Year)

		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			answer.Err = err
		} else {
			w.WriteHeader(200)
			answer.Ok = true
		}
	}else{
		w.WriteHeader(400)
		formatErrors(v)
		answer.Fields = v
	}
	json.NewEncoder(w).Encode(answer)

}

//(DELETE /tv/{id})
//Удаляет 1 телевизор из БД по id
func Remove (w http.ResponseWriter, r *http.Request) {
	app := getAppInstance()
	logRequest(r)
	w.Header().Set("Content-Type", "application/json")
	answer := Answer{}
	params := mux.Vars(r)
	id := params ["id"]
	rows := app.db.QueryRow("SELECT * FROM productdb.tv WHERE id = ?", id)

	var tv Tv
	err := rows.Scan(&tv.Id, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)
	answer.Tv = tv

	_,v := validate(tv)
	answer.Fields = v

	switch {
	case err == sql.ErrNoRows:
		w.WriteHeader(404)
		http.NotFound(w, r)
	case err != nil:
		w.WriteHeader(500)
		log.Println(err)
		answer.Err = err
	default:
		_, err = app.db.Exec("DELETE FROM productdb.tv WHERE id = ?", id)
		fmt.Printf("mysql: DELETE FROM productdb.tv WHERE id = %v \n", id)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			answer.Err = err
		}else{
			w.WriteHeader(200)
			answer.Ok = true
		}
	}

	json.NewEncoder(w).Encode(answer)
}

//(PUT /tv/{id})
//Изменяет запись 1 телевизора в БД по id
func Update (w http.ResponseWriter, r *http.Request) {
	app := getAppInstance()
	logRequest(r)
	w.Header().Set("Content-Type", "application/json")
	answer := Answer{}
	params := mux.Vars(r)
	id := params ["id"]
	rows := app.db.QueryRow("SELECT * FROM productdb.tv WHERE id = ?", id)

	var tv Tv
	err := rows.Scan(&tv.Id, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)

	switch {
	case err == sql.ErrNoRows:
		w.WriteHeader(404)
		http.NotFound(w, r)
	case (err != nil)&&(err != io.EOF):
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		answer.Err = err
	default:
		err := json.NewDecoder(r.Body).Decode(&tv)
		answer.Tv = tv
		if err != nil {
			w.WriteHeader(400)
			fmt.Println(err)
			answer.Err = err
		} else if ok, v := validate(tv); ok{
			answer.Fields = v
			_, err = app.db.Exec("UPDATE productdb.tv SET brand = ?, manufacturer = ?, model = ?, year = ? WHERE id = ?",
				tv.Brand, tv.Manufacturer, tv.Model, tv.Year, id)
			fmt.Printf("UPDATE productdb.tv SET brand = %v, manufacturer = %v, model = %v, year = %v WHERE id = %v",
				tv.Brand, tv.Manufacturer, tv.Model, tv.Year, id)
			if err != nil {
				w.WriteHeader(500)
				log.Println(err)
				answer.Err = err
			} else {
				w.WriteHeader(200)
				answer.Ok = true
			}
		}else{
			w.WriteHeader(400)
			formatErrors(v)
			answer.Fields = v
		}

	}

	json.NewEncoder(w).Encode(answer)
}


func logRequest(req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	bodyCopy := ioutil.NopCloser(bytes.NewBuffer(body))
	req.Body = bodyCopy
	fmt.Println("\n",time.Now(), "\n", req.Method, req.URL, string(body))
}

func formatErrors(v Validated) {
	if !v.Brand {
		fmt.Println("Wrong format of 'Brand':", v.Brand)
	}
	if !v.Manufacturer {
		fmt.Println("Wrong format of 'Manufacturer':", v.Manufacturer)
	}
	if !v.Model {
		fmt.Println("Wrong format of 'Brand':", v.Model)
	}
	if !v.Year {
		fmt.Println("Wrong format of 'Brand':", v.Year)
	}
}