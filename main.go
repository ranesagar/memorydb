package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type DB struct {
	data         map[string]int
	count        map[int]int
	deletedKey   map[string]bool
	deletedValue map[int]int
	root         bool // Just for debugging. Not used.
	level        int  // Just for debugging. Not used.
	parent       *DB
}

func newDB() *DB {
	return &DB{
		data:         make(map[string]int),
		count:        make(map[int]int),
		deletedKey:   make(map[string]bool),
		deletedValue: make(map[int]int),
		root:         true,
		level:        0,
		parent:       nil,
	}
}

func (db *DB) Get(name string) (int, error) {
	// recursively lookup upto the base node
	curr := db
	for db != nil {
		if _, exists := db.data[name]; exists {
			if !curr.deletedKey[name] {
				return db.data[name], nil
			}

		}
		db = db.parent
	}

	return 0, errors.New("KEY NOT FOUND-NULL")
}

func (db *DB) Delete(name string) {
	valueFromBase, err := db.Get(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	db.deletedKey[name] = true
	db.deletedValue[valueFromBase]++
}

func (db *DB) Set(name string, value int) {
	// 1.check if it was deleted in current transaction
	// 2. Check if exists in ANY transaction, if yes:
	//    decrement old value, increment new value, and update the value

	if db.deletedKey[name] {
		delete(db.deletedKey, name)
	}

	oldVal, err := db.Get(name)
	if err == nil { // there is an old value
		db.deletedValue[oldVal]++
	}

	db.count[value]++
	db.data[name] = value
}

func (db *DB) Count(c int) int {
	// recursively lookup upto the base node and keep adding
	// while adding, make sure to subtract # of times the value was delete if db.delete() was called
	totalCount := 0
	for db != nil {
		fmt.Println(totalCount, db.count[c], db.deletedValue[c])
		totalCount = totalCount + db.count[c] - db.deletedValue[c]

		db = db.parent
	}
	if totalCount < 0 {
		return 0
	}
	return totalCount
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
	return &DB{
		data:         make(map[string]int),
		count:        make(map[int]int),
		deletedKey:   make(map[string]bool),
		deletedValue: make(map[int]int),
		root:         false,
		level:        db.level + 1,
		parent:       db,
	}
}

func (db *DB) Rollback() (*DB, error) {
	if db.parent != nil {
		return db.parent, nil
	}
	return nil, errors.New("NO TRANSACTION")
}

func (db *DB) Commit() *DB {
	if db.parent == nil {
		return db
	}
	curr := db
	for curr.parent != nil {
		// 1. Deep merge all data to parent node
		for k, v := range curr.data {
			curr.parent.data[k] = v
		}
		// 2. Increment count of parent
		for k, v := range db.count {
			curr.parent.count[k] += v
		}
		// 3. Delete deleted keys. Also deep merge current deleted map
		for k, v := range curr.deletedKey {
			curr.parent.deletedKey[k] = v
			delete(curr.parent.data, k)
		}
		// 4. Decrement count of parent
		for k, v := range curr.deletedValue {
			curr.parent.count[k] -= v
			if curr.parent.count[k] <= 0 {
				delete(curr.parent.count, k)
			}
		}
		// fmt.Println(curr.parent)
		curr = curr.parent
	}

	return curr
}
