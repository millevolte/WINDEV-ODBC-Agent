package db

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/alexbrainman/odbc"
	"golang.org/x/text/encoding/charmap"
)

func DecodeISO8859(ba []uint8) []uint8 {
	dec := charmap.ISO8859_1.NewDecoder()
	out, _ := dec.Bytes(ba)
	return out
}

var DB *sql.DB = nil

func OpenDB(DSN string) {
	var err error
	DB, err = sql.Open("odbc", "DSN="+DSN)
	if err != nil {
		os.Exit(1)
	}
}

func SqlSelect(query string) (map[int]map[string]string, error) {
	rows, err := DB.Query(query, nil)
	if err != nil {
		log.Printf("error %v", err)
		return nil, err
	}

	log.Printf("%T", rows)

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	rowNum := 0
	colvals := make([]interface{}, len(cols))
	result := make(map[int]map[string]string)
	for rows.Next() {
		colassoc := make(map[string]interface{}, len(cols))
		colresult := make(map[string]string, len(cols))
		// values we"ll be passing will be pointers, themselves to interfaces
		for i, _ := range colvals {
			colvals[i] = new(interface{})
		}
		if err := rows.Scan(colvals...); err != nil {
			return nil, err
		}

		for i, col := range cols {
			colassoc[col] = *colvals[i].(*interface{})
			t := fmt.Sprintf("%T", colassoc[col])
			v := ""

			switch t {
			case "[]uint8":
				v = string(DecodeISO8859(bytes.Trim(colassoc[col].([]byte), "\x00")))
				v = strings.TrimRight(v, " ")
				break

			case "float64":
				v = fmt.Sprintf("%.2f", colassoc[col])
				break

			case "time.Time":
				var t time.Time = colassoc[col].(time.Time)
				v = fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
				break

			case "int32":
				v = fmt.Sprintf("%d", colassoc[col])
				break

			case "<nil>":
				v = ""
				break

			default:
				v = fmt.Sprintf("undefined-%T", colassoc[col])

			}
			colresult[col] = v
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		result[rowNum] = colresult
		rowNum++
	}
	// duration := time.Since(start)
	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	// log.Println(duration)
	rows.Close()
	return result, nil
}

func SqlQuery(query string) (map[int]map[string]string, error) {
	// start := time.Now()

	var commands = strings.Fields(query)
	if len(commands) < 1 {
		return nil, errors.New("wrong query")
	}
	var command = strings.ToUpper(commands[0])
	// comando SQL acquisito
	if command == "DELETE" || command == "UPDATE" {
		// log.Printf("------\n%s\n------\n", command)
		result, err := DB.Exec(query, nil)
		if err != nil {
			return nil, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		log.Printf("AFFECTED ROWS %v\n", affected)
		data := make(map[int]map[string]string)
		a := make(map[string]string)
		a["affected"] = fmt.Sprintf("%d", affected)
		data[0] = a
		return data, nil
	}

	if command == "INSERT" {
		queries := strings.Split(query, "|")
		_, err := DB.Exec(queries[0], nil)
		if err != nil {
			log.Printf("Error result\n%v\n------\n", err)
			return nil, err
		}
		start := time.Now()
		if len(queries) > 1 {
			result, err := SqlSelect(queries[1])
			duration := time.Since(start)
			log.Printf("\nSQL %s\n", queries[1])
			log.Println(duration)
			return result, err
		}

		data := make(map[int]map[string]string)
		a := make(map[string]string)
		a["lastId"] = "not implemented"
		data[0] = a
		return data, nil
	}

	if command == "SELECT" {
		return SqlSelect(query)
	}

	return nil, errors.New("WRONG SQL")
}
