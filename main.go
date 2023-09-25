package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gokyle/readpass"
	"github.com/kisom/filecrypt/crypto"
	"github.com/yang3yen/xxtea-go/xxtea"

	"robs/passmgr/db"
	"robs/passmgr/genkey"
	"robs/passmgr/makeapwd"
)

var version = "1.0.0"
var new_pass []byte
var klight []byte

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sealString(input string) string {
	b := []byte(input)
	data, err := crypto.Seal(new_pass, b)
	checkErr(err)

	j := base64.StdEncoding.EncodeToString(data)
	return string(j)
}

func openString(input string) string {
	data, err := base64.StdEncoding.DecodeString(input)
	checkErr(err)

	ret, err := crypto.Open(new_pass, data)
	checkErr(err)

	return string(ret)
}

func encodeLight(input string) string {
	b := []byte(input)
	data, err := xxtea.Encrypt(b, klight, true, 0)
	checkErr(err)

	j := base64.StdEncoding.EncodeToString(data)
	return string(j)
}

func decodeLight(input string) string {
	data, err := base64.StdEncoding.DecodeString(input)
	checkErr(err)

	ret, err := xxtea.Decrypt(data, klight, true, 0)
	checkErr(err)
	return string(ret)
}

func printRows(results *db.Cred) {
	var expires_date string = "none"
	if results.Expires_on.Valid {
		expires_date = results.Expires_on.String
	}

	dname := decodeLight(results.Domain)
	uname := decodeLight(results.Username)
	pname := openString(results.Password)

	fmt.Printf("ID#%d Domain/System:%s, Username:%s, Password:%s, Expires_on:%s, Enabled %d", results.Id, dname, uname, pname, expires_date, results.Enabled)
	if results.Notes.Valid {
		nname := decodeLight(results.Notes.String)
		fmt.Println("\n" + nname)
	}
	fmt.Println("...")
}

func main() {
	help := flag.Bool("h", false, "Help Page")
	generateKey := flag.String("g", "", "key file")
	weekweb := flag.Bool("weekweb", false, "Generate a week web random Password")
	web := flag.Bool("web", false, "Generate a random Password")
	gen := flag.Bool("gen", false, "Generate a strong random Password")
	crazy := flag.Bool("crazy", false, "Generate a crazy strong random Password")
	addEntry := flag.Bool("add", false, "Add an entry")
	deleteEntry := flag.Bool("delete", false, "Delete an entry")
	editEntry := flag.Bool("edit", false, "Edit an entry")
	enableEntry := flag.Bool("enable", false, "Enable an entry")
	disableEntry := flag.Bool("disable", false, "Disable an entry")
	list := flag.Bool("list", false, "List the entries")
	getEntryID := flag.Bool("getByID", false, "Get an entry by id #")
	getEntryDomain := flag.Bool("getByDomain", false, "Get an entry by Domain name")
	dbfile := flag.String("dbfile", "", "Database file")
	keyfile := flag.String("k", "", "Key file")
	id := flag.Int("id", 0, "Use id#")
	domain := flag.String("domain", "", "Website Domain/Server Name")
	username := flag.String("username", "", "Username")
	expires := flag.String("expires", "", "Expires on")
	notes := flag.String("notes", "", "System Notes")
	pwd := flag.String("pwd", "", "Password")
	flag.Parse()

	if *help {
		usage()
		return
	}

	var password string

	if *weekweb {
		password = makeapwd.MakePass(8, 13)
		fmt.Println(password)
	} else if *web {
		password = makeapwd.MakePass(13, 18)
		fmt.Println(password)
	} else if *gen {
		password = makeapwd.MakePass(15, 24)
		fmt.Println(password)
	} else if *crazy {
		password = makeapwd.MakePass(32, 48)
		fmt.Println(password)
	}

	// Just return now as it displayed a new RND password and that was the only Argument given!
	if password != "" && len(os.Args) == 2 {
		return
	}

	if *generateKey != "" {
		genkey.MakeKey(*generateKey)
		return
	}

	var pass []byte
	var err error
	pass, err = readpass.PasswordPromptBytes("Database password: ")
	checkErr(err)

	if *keyfile != "" {
		// Open keyfile for reading
		file, err := os.Open(*keyfile)
		checkErr(err)
		data, err := ioutil.ReadAll(file)
		checkErr(err)
		new_pass = append(data, pass...)
	} else {
		new_pass = pass
	}

	if *dbfile != "" {
		db.OpenDatabase(*dbfile)
		db.CreateCredsTable()
		key, _ := xxtea.URandom(16)
		klight = db.CreateRNDTable(key)
	}

	if *deleteEntry && *id > 0 {
		db.DeleteCred(*id)
	}

	if *enableEntry && *id > 0 {
		db.EnableCred(*id, true)
	}

	if *disableEntry && *id > 0 {
		db.EnableCred(*id, false)
	}

	if *editEntry && *id > 0 {
		fmt.Println("Please wait...!")
		if *pwd != "" {
			pdata := sealString(*pwd)
			db.UpdateCred(*id, "password", pdata)
		}
		if *domain != "" {
			ddata := sealString(*domain)
			db.UpdateCred(*id, "domain", ddata)
		}
		if *username != "" {
			udata := sealString(*username)
			db.UpdateCred(*id, "username", udata)
		}
		if *expires != "" {
			db.UpdateCred(*id, "expires", *expires)
		}
		if *notes != "" {
			ndata := sealString(*notes)
			db.UpdateCred(*id, "notes", ndata)
		}
	}

	if *addEntry && (*domain != "" || *username != "") {
		if *pwd != "" {
			password = *pwd
		} else if *pwd == "" && password == "" {
			password = makeapwd.MakePass(32, 48)
		}

		fmt.Println("Please wait...!")

		ddata := encodeLight(*domain)
		udata := encodeLight(*username)
		pdata := sealString(password)
		var ndata string = ""
		if *notes != "" {
			ndata = encodeLight(*notes)
		}

		db.InsertCred(ddata, udata, pdata, ndata)
	}

	if *getEntryID && *id > 0 {
		result := db.FindCredById(*id)
		printRows(result)
	}

	if *getEntryDomain && *domain != "" {
		results := db.FindCredByDomain(*domain)
		for _, slice := range results {
			printRows(slice)
		}
	}

	if *list {
		results := db.FindAll()
		for _, slice := range results {
			printRows(slice)
		}
	}

	if *dbfile != "" {
		db.CloseConnections()
	}
}

func usage() {
	progName := filepath.Base(os.Args[0])
	fmt.Printf(`%s version %s, (c) 2022 Bob S. (TechnoWizardBob.com)

Usage:
%s [-h] [-dbfile filename] [-k keyfile] [-list] [-add] [...]
	-h 		Print this help message.
For use with displaying random passwords:
	-weekweb	Make a week web page password for screen.
	-web 		Make a web page password for screen.
	-gen		Make a password for screen.
	-crazy		Make a crazy password for screen.
Accessing the password store:
	-dbfile		Encrypted SQLite3 Database.
	-g		Generate a keyfile in path/file.
	-k		(Optional) Uses this keyfile.	
For use with editing entries:
	-delete		Delete an entry.
	-edit		Edit an entry.
	-enable		Mark as enabled.
	-disable	Mark as disabled.
	-expires	When password is no longer valid.
	-id		Use only the entry for ID#.
Creating new entries:
	-add		Add a new entry.
	-domain 	Add this domain name.
	-username	Add this username.
	-pwd		Add this password.
For use with listing:
	-list 		Show all the entries.
	-getByID	Show the entry with ID#.
	-getByDoamin	Show the entries with domain name that matches.

Examples:
%s -dbfile ~/.main.db -k ~/.my.key -add -domain "gmail.com" -username "msmith" -pwd "@@S*fg8s^g8ds#afg!78g"
	Add a new password.

%s -dbfile ~/.main.db -k ~/.my.key -list
	List all passwords.

%s -dbfile ~/.main.db -k ~/.my.key -getByDomain -domain "youtube"
	View entries for a domain name.

%s -dbfile ~/.main.db -k ~/.my.key -disable -id 2
	Disable the entry with the ID of #2.

%s -dbfile ~/.main.db -k ~/.my.key -delete -id 2
	Delete an entry by ID of #2.

%s -dbfile ~/.main.db -k ~/.my.key -edit -id 1 -pwd "NewP@330rd"
	Edit an entry's password with the ID of #1.

`, progName, version, progName, progName, progName, progName, progName, progName, progName)
}
