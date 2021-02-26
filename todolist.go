package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // To import a package solely for its side-effects (initialization), use the blank identifier
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id int `gorm:"primary_key"`
	Description string
	Completed bool
}

// Helpers
func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func GetItemByID(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

// GET Functions
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"active": true}`)
}

func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Incomplete TodoItems")
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)
}

func GetCompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Completed TodoItems");
	CompleteTodoItems := GetTodoItems(true);
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CompleteTodoItems)
}

// POST Functions
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
	w.Header().Set("Content-Type", "application/json")
	// Return JSON response
	json.NewEncoder(w).Encode(result.Value)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameters from mux
	vars := mux.Vars(r)
	// Get value for "id", convert to integer, short assign to id
	id, _ := strconv.Atoi(vars["id"])
	// Test if the TodoItem exists in the DB
	err := GetItemByID(id)

	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		// strconv.ParseBool returns 2 values, the second of which being an
		// error if passed an invalid value so we use the blank identifier `_`,
		// otherwise at compilation time you'll get a assignment mismatch error 
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
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
	// GET
	router.HandleFunc("/", HealthCheck).Methods("GET")
	router.HandleFunc("/check", HealthCheck).Methods("GET")
	router.HandleFunc("/todo", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/complete", GetCompleteItems).Methods("GET")
	// POST
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	http.ListenAndServe(":8000", router)
}
