package todo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Create will allow a user to create a new todo
// The supported body is {"title": "", "status": ""}
func Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dbUser := "postgres"
	dbHost := "localhost"
	dbPassword := "b4n4n4s"
	dbName := "postgres"

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	var todo CreateTodo

	json.NewDecoder(r.Body).Decode(&todo)

	if todo.Status == "" || todo.Title == "" {
		http.Error(w, "Todo request is missing status and/or title", http.StatusBadRequest)
	} else {

		invalidStatus := true
		for _, status := range allowedStatuses {
			if todo.Status == status {
				invalidStatus = false
				break
			}
		}		

		//double negatives are tricky :-)
		if invalidStatus {
			http.Error(w, "The provided status is not supported", http.StatusBadRequest)
		} else {

			insertStmt := fmt.Sprintf(`INSERT INTO todo (title, status) VALUES ('%s', '%s') RETURNING id`, todo.Title, todo.Status)

			var todoID int

			// Insert and get back newly created todo ID
			if err := db.QueryRow(insertStmt).Scan(&todoID); err != nil {
				fmt.Printf("Failed to save to db: %s", err.Error())
			}

			fmt.Printf("Todo Created -- ID: %d\n", todoID)

			newTodo := Todo{}
			db.QueryRow("SELECT id, title, status FROM todo WHERE id=$1", todoID).Scan(&newTodo.ID, &newTodo.Title, &newTodo.Status)

			jsonResp, _ := json.Marshal(newTodo)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			fmt.Fprintf(w, string(jsonResp))
		}
	}
}

// List will provide a list of all current to-dos
func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		dbUser := "postgres"
	dbHost := "localhost"
	dbPassword := "b4n4n4s"
	dbName := "postgres"

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	todoList := []Todo{}

	rows, err := db.Query("SELECT id, title, status FROM todo")
	defer rows.Close()

	for rows.Next() {
		todo := Todo{}
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Status); err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Failed to build todo list")
		}

		todoList = append(todoList, todo)
	}

	jsonResp, _ := json.Marshal(Todos{TodoList: todoList})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, string(jsonResp))
}
//for update requirements: should the user be able to effectively replace a todo, or rather be limited to only update status on one(not title)?
func Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	dbUser := "postgres"
	dbHost := "localhost"
	dbPassword := "b4n4n4s"
	dbName := "postgres"

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err.Error())
	}
//////////////DB BOILER ABOVE, might refactor into a dbservice, decouple processing from DBAdmin etc
//////////////Input validation boiler below,

	var todo CreateTodo // right model, could address naming for better clarity.
	json.NewDecoder(r.Body).Decode(&todo)

	if todo.Status == "" || todo.Title == "" {
		http.Error(w, "Todo request is missing status and/or title", http.StatusBadRequest)
	} else {

		invalidStatus := true
		for _, status := range allowedStatuses {
			if todo.Status == status {
				invalidStatus = false
				break
			}
		}

		//double negatives are tricky :-)
		if invalidStatus {
			http.Error(w, "The provided status is not supported", http.StatusBadRequest)
		} else {
			var resultingTodo Todo
    		fmt.Printf("updating todoID %s", params.ByName("todoID"))
			updateStmt := fmt.Sprintf(`UPDATE todo SET title = '%s', status = '%s' WHERE id = %s RETURNING id, title, status;`, todo.Title, todo.Status, params.ByName("todoID"))
		    fmt.Printf("update sttm: %s", updateStmt)

			// Insert and get back newly created todo ID
			if err := db.QueryRow(updateStmt).Scan(&resultingTodo.ID, &resultingTodo.Title, &resultingTodo.Status); err != nil {
				fmt.Printf("Failed to save to db: %s", err.Error())
			}

			fmt.Printf("Todo Updated -- id: %s\n", params.ByName("todoID"))

			jsonResp, _ := json.Marshal(resultingTodo)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			fmt.Fprintf(w, string(jsonResp))
		}
	}
}
func Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	dbUser := "postgres"
	dbHost := "localhost"
	dbPassword := "b4n4n4s"
	dbName := "postgres"
	fmt.Println("in delete, delete %s", params.ByName("todoID"))

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err.Error())
	}
//////////////DB BOILER ABOVE
	deleteStmt := fmt.Sprintf(`DELETE FROM todo WHERE id = %s;`, params.ByName("todoID"))

	res, err := db.Exec(deleteStmt)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		fmt.Printf("Failed to save to db: %s", err.Error())
	} 
	
	if count == 0 {
		http.Error(w, "No todo found", 404)
	} else {
	fmt.Printf("Todo Deleted -- id: %s\n record removed %d", params.ByName("todoID"), int(count))
	w.WriteHeader(200)
	fmt.Fprint(w, "OK\n")
	}
}