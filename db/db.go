package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Cred struct {
	Id         int
	Domain     string
	Username   string
	Password   string
	Expires_on sql.NullString
	Notes      sql.NullString
	Enabled    int
}

var db *sql.DB

func checkErr(err error) {
	if err != nil {
		//log.Fatalln(err)
		panic(err)
	}
}

func OpenDatabase(sqlfile string) {
	var err error

	db, err = sql.Open("sqlite3", sqlfile)
	checkErr(err)
}

func CloseConnections() {
	db.Close()
}

func QuickTable() {
	createTableSQL := `CREATE TABLE IF NOT EXISTS quick (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"e" TEXT,
	);`

	statement, err := db.Prepare(createTableSQL)
	checkErr(err)

	statement.Exec()
	// log.Println("Quck Cache table created.")
}

func CreateRNDTable(key []byte) []byte {
	createTableSQL := `CREATE TABLE IF NOT EXISTS gen (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"g" BLOB
	);`

	statement, err := db.Prepare(createTableSQL)
	checkErr(err)

	statement.Exec()
	// log.Println("Unikey table created.")
	defer statement.Close()

	getKeySql := `SELECT g FROM gen WHERE id = 1`
	var g []byte
	rows, err := db.Query(getKeySql)
	checkErr(err)
	if rows.Next() {
		rows.Scan(&g)
		return g
	} else {
		replaceSQL := `REPLACE INTO gen (id, g)
			VALUES (1, ?);`
		s, err := db.Prepare(replaceSQL)
		checkErr(err)

		_, err = s.Exec(key)
		checkErr(err)

		defer s.Close()
		return key
	}
}

func CreateCredsTable() {
	createTableSQL := `CREATE TABLE IF NOT EXISTS credentials (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"domain" TEXT,
		"username" TEXT,
		"password" TEXT,
		"expires_on" TEXT NULL,
		"notes" TEXT NULL,
		"enabled" INTEGER
	);`

	statement, err := db.Prepare(createTableSQL)
	checkErr(err)

	statement.Exec()
	defer statement.Close()
	// log.Println("Credentials table created.")
}

func InsertCred(domain string, username string, password string, notes string) {
	if notes == "" {
		insertCredSQL := `INSERT INTO credentials (domain, username, password, enabled)
		VALUES (?, ?, ?, 1)`

		statement, err := db.Prepare(insertCredSQL)
		checkErr(err)

		_, err = statement.Exec(domain, username, password)
		checkErr(err)

		defer statement.Close()
		log.Println("New credentials saved successfully")
	} else {
		insertCredSQL := `INSERT INTO credentials (domain, username, password, notes, enabled)
		VALUES (?, ?, ?, ?, 1)`

		statement, err := db.Prepare(insertCredSQL)
		checkErr(err)

		_, err = statement.Exec(domain, username, password, notes)
		checkErr(err)

		defer statement.Close()
		log.Println("New credentials saved successfully")
	}
}

func DeleteCred(id int) {
	deleteCredSql := `DELETE FROM credentials WHERE id = ?`

	statement, err := db.Prepare(deleteCredSql)
	checkErr(err)

	_, err = statement.Exec(id)
	checkErr(err)
	defer statement.Close()

	log.Printf("Credentials for %d has been correctly deleted", id)
}

func EnableCred(id int, enabled bool) {
	var CredSql string
	if enabled {
		CredSql = `UPDATE credentials set enabled=1 WHERE id = ?`
	} else {
		CredSql = `UPDATE credentials set enabled=0 WHERE id = ?`
	}

	statement, err := db.Prepare(CredSql)
	checkErr(err)

	_, err = statement.Exec(id)
	checkErr(err)
	defer statement.Close()

	log.Printf("Credentials for %d has been correctly modified", id)
}

func dataFromDbRow(rows *sql.Rows) *Cred {
	cred := new(Cred)
	err := rows.Scan(&cred.Id, &cred.Domain, &cred.Username, &cred.Password, &cred.Expires_on, &cred.Notes, &cred.Enabled)
	checkErr(err)
	return cred
}

func FindCredByDomain(domain string) []*Cred {
	findCredSql := `SELECT * FROM credentials WHERE domain LIKE '%'||?||'%'`
	rows, err := db.Query(findCredSql, domain)
	checkErr(err)
	var aRecords []*Cred
	for rows.Next() {
		var cred = new(Cred)
		cred = dataFromDbRow(rows)
		aRecords = append(aRecords, cred)
	}
	rows.Close()
	return aRecords
}

func FindCredById(id int) *Cred {
	findCredSql := `SELECT * FROM credentials WHERE id = ? LIMIT 1`
	rows, err := db.Query(findCredSql, id)
	checkErr(err)
	var cred = new(Cred)
	if rows.Next() {
		cred = dataFromDbRow(rows)
	}
	rows.Close()
	return cred
}

func FindAll() []*Cred {
	findCredSql := `SELECT * FROM credentials`
	rows, err := db.Query(findCredSql)
	checkErr(err)
	var aRecords []*Cred
	for rows.Next() {
		cred := dataFromDbRow(rows)
		// log.Println(cred)
		aRecords = append(aRecords, cred)
	}
	rows.Close()
	return aRecords
}

func UpdateCred(id int, field string, value string) {
	var CredSql string
	if field == "password" {
		CredSql = `UPDATE credentials set password=? WHERE id = ?`
	} else if field == "username" {
		CredSql = `UPDATE credentials set username=? WHERE id = ?`
	} else if field == "domain" {
		CredSql = `UPDATE credentials set domain=? WHERE id = ?`
	} else if field == "expires" {
		CredSql = `UPDATE credentials set expires_on=? WHERE id = ?`
	} else if field == "notes" {
		CredSql = `UPDATE credentials set notes=? WHERE id = ?`
	}

	statement, err := db.Prepare(CredSql)
	checkErr(err)

	_, err = statement.Exec(value, id)
	checkErr(err)
	defer statement.Close()

	log.Printf("Credentials for %d has been correctly modified", id)
}
