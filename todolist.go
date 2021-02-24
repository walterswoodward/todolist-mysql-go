package main

import (
	"io"
	"net/http"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql" // To import a package solely for its side-effects (initialization), use the blank identifier
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db, _ = gorm.Open("mysql", "root:abracadabra@/todolist?charset=utf8&parseTime=True&loc=Local")

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
	
	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/ping", Ping).Methods("GET")
	http.ListenAndServe(":8000", router)
}
