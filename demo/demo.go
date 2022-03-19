package main

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const MAXRANDOM = 25

func main() {

	createTable := flag.Bool("create", false, "create table in the database")
	dropTable := flag.Bool("drop", false, "Drop current table from the database")
	insertRandomValues := flag.Int("insert", 0, "insert N random values into the database")
	poisonRecordToInsert := flag.String("insert_poison", "", "insert poison record (should be in BASE64 format)")
	selectAllFromTable := flag.Bool("select", false, "select all stored values from database")

	flag.Parse()

	db, err := sql.Open("postgres", "sslmode=disable dbname=test user=test password=test host=127.0.0.1 port=9393")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return
	}

	if *createTable {
		_, err := db.Exec("DROP TABLE IF EXISTS test_table;")
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = db.Exec("DROP SEQUENCE IF EXISTS test_table_seq;")
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = db.Exec("CREATE SEQUENCE test_table_seq START 1;")
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS test_table(id INTEGER PRIMARY KEY DEFAULT nextval('test_table_seq'), username BYTEA, password BYTEA, email BYTEA);")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Table has been successfully created")
	}

	if *dropTable {
		_, err := db.Exec("DROP TABLE IF EXISTS test_table;")
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = db.Exec("DROP SEQUENCE IF EXISTS test_table_seq;")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Table has been successfully dropped")
	}

	if *insertRandomValues > 0 {
		if *insertRandomValues > MAXRANDOM {
			log.Fatal("Too much to insert. Use value from range [1 .. " + fmt.Sprint(MAXRANDOM) + "]")
			return
		}

		emails, err := loadFile("demo/emails")
		if err != nil {
			log.Fatal(err)
		}
		passwords, err := loadFile("demo/passwords")
		if err != nil {
			log.Fatal(err)
		}
		usernames, err := loadFile("demo/usernames")
		if err != nil {
			log.Fatal(err)
		}

		s1 := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s1)

		for i := 0; i < *insertRandomValues; i++ {
			userName := getRandomInput(r, usernames)
			password := getRandomInput(r, passwords)
			email := getRandomInput(r, emails)
			_, err = db.Exec(`insert into test_table(username, password, email) values ('\x` + hex.EncodeToString(userName) + `', '\x` + hex.EncodeToString(password) + `', '\x` + hex.EncodeToString(email) + `')`)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		log.Println("Insert has been successful")
	}

	if *poisonRecordToInsert != "" {
		value, err := base64.StdEncoding.DecodeString(*poisonRecordToInsert)
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = db.Exec(`insert into test_table(username, password, email) values ('poison_record', '\x` + hex.EncodeToString(value) + `', '\x` + hex.EncodeToString(value) + `')`)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Poison record insert has been successful")
	}

	if *selectAllFromTable {
		result, err := db.Query("select * from test_table")
		if err != nil {
			log.Fatal(err)
			return
		}
		result.Close()
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Select has been successful")
	}
}

func getRandomInput(r *rand.Rand, from []string) []byte {
	return []byte(from[r.Intn(MAXRANDOM)])
}

func loadFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

const (
	// TagSymbol used in begin tag in AcraStruct
	TagSymbol byte = '"'
)

const (
	// length of EC public key
	PublicKeyLength = 45
	// length of 32 byte of symmetric key wrapped to smessage
	SMessageKeyLength = 84
	KeyBlockLength    = PublicKeyLength + SMessageKeyLength
	DataLengthSize    = 8
)

// TagBegin represents begin sequence of bytes for AcraStruct.
var TagBegin = []byte{TagSymbol, TagSymbol, TagSymbol, TagSymbol, TagSymbol, TagSymbol, TagSymbol, TagSymbol}

var (
	ErrIncorrectAcraStructTagBegin   = errors.New("AcraStruct has incorrect TagBegin")
	ErrIncorrectAcraStructLength     = errors.New("AcraStruct has incorrect length")
	ErrIncorrectAcraStructDataLength = errors.New("AcraStruct has incorrect data length value")
)
