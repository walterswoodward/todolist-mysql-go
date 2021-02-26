package main

import (
	"io"
	"net/http"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql" // To import a package solely for its side-effects (initialization), use the blank identifier
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"encoding/json"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id int `gorm:"primary_key"`
	Description string
	Completed bool
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	// Obtain POST request value for description
	description := r.FormValue("description")
	// Log that the addition of the new todo item is about to be saved to the database
	log.WithFields(log.Fields{"description": description}).Info("Add a new TodoItem. Saving to database.")
	// Create the todo object to be saved to the database and short assign it to `todo` 
	todo := &TodoItemModel{Description: description, Completed: false}
	// Pass reference of reference of TodoItemModel object (via pointer) to db.Create(...) to add it to the database
	db.Create(&todo)
	// Use db.Last(...) to get the newly added row in todo_item_models
	result := db.Last(&todo)
	// Set response header
	w.Header().Set("Content-Type", "application/json")
	// Return JSON response
	json.NewEncoder(w).Encode(result.Value)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"species": "Felis catus",
		"count": 1000,
	}).Info("A group of cats emerge from the tree line after hearing an unfamiliar tone...")
	
	// set header + write to view
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `pong`)
}

func init() {
	// With the default log.SetFormatter(&log.TextFormatter{}) when a TTY 
	// is not attached, the output is compatible with the logfmt format
	log.SetFormatter(&log.TextFormatter{})
	// add the calling method as a field
	log.SetReportCaller(true)
}

func main() {
	// close db connection once main is returned
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})
	
	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/ping", Ping).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	http.ListenAndServe(":8000", router)
}
