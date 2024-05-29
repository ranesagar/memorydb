package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type DB struct {
	data   map[string]int
	count  map[int]int
	root   bool // Just for debugging. Not used.
	level  int  // Just for debugging. Not used.
	parent *DB
}

func newDB() *DB {
	return &DB{
		data:   make(map[string]int),
		count:  make(map[int]int),
		root:   true,
		level:  0,
		parent: nil,
	}
}

func (db *DB) Get(name string) (int, error) {

	if _, exists := db.data[name]; exists {
		return db.data[name], nil
	}

	return 0, errors.New("KEY NOT FOUND-NULL")
}

func (db *DB) Delete(name string) {
	if v, exists := db.data[name]; exists {
		// v := db.data[name]
		delete(db.data, name)
		db.count[v]--
		if db.count[v] == 0 {
			delete(db.count, v)
		}
	}
}

func (db *DB) Set(name string, value int) {
	// check if exists, decrement old value, increment new value
	// and update the value
	if _, exists := db.data[name]; exists {
		oldVal := db.data[name]
		db.count[oldVal]--
		if db.count[oldVal] == 0 {
			delete(db.count, oldVal)
		}
	}
	// increment or add new value
	db.count[value]++
	db.data[name] = value
}

func (db *DB) Count(c int) int {
	if _, exists := db.count[c]; exists {
		return db.count[c]
	}
	return 0
}

func main() {
	fmt.Println("In-memory DB")
	db := newDB()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter command: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		switch command {
		case "SET":
			if len(parts) != 3 {
				fmt.Println("Usage: SET name value")
				continue
			}
			name := parts[1]
			var value int
			fmt.Sscanf(parts[2], "%d", &value)
			db.Set(name, value)
			db.Log()
			fmt.Println("OK")
		case "GET":
			if len(parts) != 2 {
				fmt.Println("Usage: GET name")
				continue
			}
			name := parts[1]
			value, err := db.Get(name)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(value)
			}
			db.Log()
		case "DELETE":
			if len(parts) != 2 {
				fmt.Println("Usage: DELETE name")
				continue
			}
			name := parts[1]
			db.Delete(name)
			db.Log()
			fmt.Println("OK")
		case "COUNT":
			if len(parts) != 2 {
				fmt.Println("Usage: COUNT value")
				continue
			}
			var value int
			fmt.Sscanf(parts[1], "%d", &value)
			count := db.Count(value)
			fmt.Println(count)
		case "BEGIN":
			db = db.Begin()
			db.Log()
			fmt.Println("OK")
		case "ROLLBACK":
			var err error
			db, err = db.Rollback()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("OK")
			}
		case "COMMIT":
			db = db.Commit()
			db.Log()
			fmt.Println("OK")
		case "END":
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func (db *DB) Log() {
	for db != nil {
		fmt.Println(db)
		db = db.parent
	}
}

func (db *DB) Begin() *DB {
	newData := make(map[string]int)
	for k, v := range db.data {
		newData[k] = v
	}

	newCount := make(map[int]int)
	for k, v := range db.count {
		newCount[k] = v
	}

	return &DB{
		data:   newData,
		count:  newCount,
		level:  db.level + 1,
		parent: db,
	}
}

func (db *DB) Rollback() (*DB, error) {
	if db.parent != nil {
		return db.parent, nil
	}
	return nil, errors.New("NO TRANSACTION")
}

func (db *DB) Commit() *DB {
	curr := db
	for curr.parent != nil {
		curr = curr.parent
	}
	newData := make(map[string]int)
	for k, v := range db.data {
		newData[k] = v
	}

	newCount := make(map[int]int)
	for k, v := range db.count {
		newCount[k] = v
	}
	curr.data = newData
	curr.count = newCount
	return curr
}
