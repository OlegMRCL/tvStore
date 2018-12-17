package main
import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"log"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"sync"
)


type Tv struct{
	Id int
	Brand string
	Manufacturer string
	Model string
	Year int
}


type Cfg struct{
	ListenHost string
	ListenPort string
	DbName string
	DbUser string
	DbPassword string

}

type Application struct {
	config Cfg
	db *sql.DB
}

var appInstance *Application
var once sync.Once

func getAppInstance() *Application {
	once.Do(func(){
		appInstance = new(Application)

		//Подгружаем конфиги из toml
		appInstance.config = loadConfigs()

		//Подключаем БД
		appInstance.db = connectMysql(&appInstance.config)
	})
	return appInstance
}

func loadConfigs() Cfg {
	cfg := Cfg{}
	_, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		fmt.Println(err)
	}
	return cfg
}

func connectMysql(cfg *Cfg) *sql.DB{
	dbData := cfg.DbUser + ":" + cfg.DbPassword + "@/" + cfg.DbName
	db, err := sql.Open("mysql", dbData)
	if err != nil {
		log.Println(err)
	}
	return db
}


type Validated struct {
	Brand bool
	Manufacturer bool
	Model bool
	Year bool
}

func validate(tv Tv) (ok bool, result Validated){

	result = Validated{
		Brand: true,
		Manufacturer: len(tv.Manufacturer)>2,
		Model: len(tv.Model)>1,
		Year: tv.Year >= 2010,
	}

	ok = result.Brand && result.Manufacturer && result.Model && result.Year

	return
}

func (app *Application) run() {

	r := mux.NewRouter()
	defer app.db.Close()
	r.HandleFunc("/tv", All).Methods("GET")
	r.HandleFunc("/tv/{id}", Get).Methods("GET")
	r.HandleFunc("/tv", Add).Methods("POST")
	r.HandleFunc("/tv/{id}", Remove).Methods("DELETE")
	r.HandleFunc("/tv/{id}", Update).Methods("PUT")


	listen := app.config.ListenHost + ":" + app.config.ListenPort
	fmt.Println(listen)
	fmt.Println("Server is listening...")
	http.ListenAndServe(listen, r)
}


func main() {

	app := getAppInstance()
	app.run()

}
