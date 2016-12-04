package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql",
		"stone:password@/Stone?collation=utf8_general_ci")
	checkErr(err)
	defer db.Close()

	// insert
	stmt, err := db.Prepare(`INSERT INTO Invoice SET
                           CreatedAt=?, ReferenceMonth=?, ReferenceYear=?,
                           Document=?, Description=?, Amount=?,
                           IsActive=?, DeactiveAt=?`)
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(time.Now(), 12, 2016, "DocA", "abc", 16.346, true, nil)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println("Inserted id: ", id)

	// update
	stmt, err = db.Prepare("UPDATE Invoice set Document=? WHERE Id=?")
	checkErr(err)
	defer stmt.Close()

	res, err = stmt.Exec("DocZ", id)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println("Affected by update: ", affect)

	//query
	rows, err := db.Query("SELECT * FROM Invoice")
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var id int
		var createdAt string
		var referenceMonth int
		var referenceYear int
		var document string
		var description string
		var amount float32
		var isActive int8
		var deactiveAt sql.NullString
		err = rows.Scan(&id, &createdAt, &referenceMonth, &referenceYear,
			&document, &description, &amount, &isActive, &deactiveAt)
		checkErr(err)
		fmt.Println(id)
		fmt.Println(createdAt)
		fmt.Println(referenceMonth)
		fmt.Println(referenceYear)
		fmt.Println(document)
		fmt.Println(description)
		fmt.Println(amount)
		fmt.Println(isActive)
		fmt.Println(deactiveAt)
	}
	checkErr(rows.Err())

	// delete
	stmt, err = db.Prepare("DELETE FROM Invoice WHERE Id=?")
	checkErr(err)
	defer stmt.Close()

	res, err = stmt.Exec(id)
	checkErr(err)

	affect, err = res.RowsAffected()
	checkErr(err)

	fmt.Println("Affected by delete: ", affect)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
