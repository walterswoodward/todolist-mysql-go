package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/cors"
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

func GetAllTodoItems() interface{} {
	var todos []TodoItemModel
	TodoItems := db.Find(&todos).Value
	return TodoItems
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

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get All TodoItems");
	TodoItems := GetAllTodoItems();
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TodoItems)
}

// POST Functions
func CreateItem(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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

func UpdateItems(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// Get URL parameters from mux
	params := mux.Vars(r)
	// Get value for "id", convert to integer, short assign to id
	id, _ := strconv.Atoi(params["id"])
	// Test if the TodoItem exists in the DB
	err := GetItemByID(id)

	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		todo := &TodoItemModel{}
		db.First(&todo, id)

		completed := r.FormValue("completed")
		if completed != "" {
			completed, err := strconv.ParseBool(completed);
			if err == nil {
				todo.Completed = completed
			} else {
				log.Info(err)
			}
		}

		description := r.FormValue("description")
		if description != "" {
			todo.Description = description
		}
		log.WithFields(log.Fields{"Id": id, "Completed": completed, "Description": description}).Info("Updating TodoItem")
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

// DELETE functions
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameters from mux
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	// Test if the TodoItem exists in the database
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
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
	router.HandleFunc("/incomplete", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/complete", GetCompleteItems).Methods("GET")
	router.HandleFunc("/all", GetAllItems).Methods("GET")
	// POST
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItems).Methods("POST")
	// DELETE
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	http.ListenAndServe(":8000", handler)
}
