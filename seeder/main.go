package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	names    [5163]string
	surnames [88799]string
)

const (
	namesCap    = len(names) - 1
	surnamesCap = len(surnames) - 1
)

type randomUser struct {
	Username  string
	Firstname string
	Lastname  string
	Password  string
	Gender    string
}

var password string

func newRandomUser() randomUser {
	firstName := names[rand.Intn(namesCap)]
	lastName := surnames[rand.Intn(surnamesCap)]
	s := time.Now().UnixNano()
	username := fmt.Sprint(s) + "." + strings.ToLower(lastName) + "." + strings.ToLower(firstName)
	if len(username) > 30 {
		username = username[0:30]
	}
	gender := "f"
	if s%2 == 1 {
		gender = "m"
	}

	return randomUser{
		username,
		firstName,
		lastName,
		password,
		gender,
	}
}

func populateArrays(arr []string, filename string) error {
	file, err := os.Open("data/" + filename + ".txt")
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var (
		i   int
		max = len(arr)
	)
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		arr[i] = sc.Text()
		i++
		if i >= max {
			break
		}
	}

	return sc.Err()
}

func init() {
	rand.Seed(time.Now().UnixNano())
	pass, err := bcrypt.GenerateFromPassword([]byte("1234567890"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	password = string(pass)
}

const rowsPerInsert = 5000

func exec(db *sql.DB, placeholders []string, params []interface{}) {
	stmt := fmt.Sprintf("INSERT INTO `users` (`username`, `first_name`, `last_name`, `gender`, `password_hash`) VALUES %s",
		strings.Join(placeholders, ","))
	_, err := db.Exec(stmt, params...)
	if err != nil {
		fmt.Println("failed to execute batch", err)
	}
}
func insert(db *sql.DB, qty int, wg *sync.WaitGroup) {
	placeholders := make([]string, 0, rowsPerInsert)
	params := make([]interface{}, 0, rowsPerInsert*5)
	for i := 0; i < qty; i++ {
		u := newRandomUser()
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		params = append(params, u.Username, u.Firstname, u.Lastname, u.Gender, u.Password)
		if len(placeholders) >= rowsPerInsert {
			exec(db, placeholders, params)
			placeholders = make([]string, 0, rowsPerInsert)
			params = make([]interface{}, 0, rowsPerInsert*5)
		}
	}
	if len(placeholders) > 0 {
		exec(db, placeholders, params)
	}
	fmt.Printf("%d rows where been inserted\n", qty)
	wg.Done()
}

func main() {

	qty := flag.Int("q", 10, "Quantity of users to generate")
	dbUser := flag.String("u", "root", "username to connect with DB")
	dbPassword := flag.String("p", "", "password to connect with DB")
	dbDSN := flag.String("dbHost", "127.0.0.1:3336", "host name with port")
	dbName := flag.String("dbName", "otus_ha", "database name")
	flag.Parse()

	db, err := sql.Open("mysql", *dbUser+":"+*dbPassword+"@tcp("+*dbDSN+")/"+*dbName+"?parseTime=true")
	if err != nil {
		panic(err)
	}

	err = populateArrays(names[:], "firstnames")
	if err != nil {
		panic(err)
	}
	err = populateArrays(surnames[:], "surnames")
	if err != nil {
		panic(err)
	}

	fmt.Printf("populated names(%d), surnames(%d)\n", len(names), len(surnames))

	batchSize := int(math.Ceil(float64(*qty) / float64(runtime.NumCPU())))
	fmt.Printf("have %d cores with batch size %d\n", runtime.NumCPU(), batchSize)
	var wg sync.WaitGroup
	for *qty > 0 {
		if *qty < batchSize {
			batchSize = *qty
		}
		wg.Add(1)
		go insert(db, batchSize, &wg)
		*qty -= batchSize
	}
	wg.Wait()

	fmt.Println("Done")
}
