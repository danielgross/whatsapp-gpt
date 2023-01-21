package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	tableDefinition := defineTableByFile("data")
	insertion := getInsertSQL("data", tableDefinition.types)
	execSql(tableDefinition.deletion, tableDefinition.creation, tableDefinition.insertion, insertion)
}

func execSql(deletion string, creation string, insertion string, values [][]string) {

	db, err := sql.Open("mysql", "tripactions:prodActive00@tcp(127.0.0.1:3306)/gpt")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(deletion)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(creation)
	if err != nil {
		panic(err)
	}
	for _, rec := range values {
		args := make([]interface{}, len(rec))
		for i, v := range rec {
			args[i] = strings.ReplaceAll(v, "\"", "")
		}

		query := fmt.Sprintf(insertion, args...)
		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("SQL executed successfully")
}

func getInsertSQL(filename string, columnTypes []string) [][]string {
	file, err := os.Open(filename + ".csv")
	if err != nil {
		// handle the error
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		// handle the error
	}
	content, err := reader.ReadAll()
	if err != nil {
		// handle the error
	}
	values := [][]string{}
	for _, line := range content {
		lineValue := []string{}
		for i, columnValue := range line {
			lineValue = append(lineValue, convert(columnTypes[i], columnValue))
		}
		values = append(values, lineValue)
	}
	return values
}

func convert(columnType string, value string) string {
	if columnType == "DATETIME2" {
		value = strings.Replace(value, "T", " ", 1)
		value = strings.Replace(value, "Z", "", 1)
	}
	return value
}

type TableDefinition struct {
	creation  string
	deletion  string
	insertion string
	types     []string
}

func defineTableByFile(filename string) TableDefinition {
	file, err := os.Open(filename + ".csv")
	if err != nil {
		// handle the error
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		// handle the error
	}
	firstRow, err := reader.Read()
	if err != nil {
		// handle the error
	}
	var types []string
	var columnTypes []string
	var empties []string
	for i, columnName := range header {
		columnType := mysqlDataType(firstRow[i])
		columnTypes = append(columnTypes, columnType)
		if strings.HasPrefix(columnType, "DATETIME") {
			types = append(types, columnName+" DATETIME")
		} else {
			types = append(types, columnName+" "+columnType)
		}
		empties = append(empties, "\"%s\"")
	}
	deletion := "DROP TABLE IF EXISTS " + filename + ";"
	creation := "CREATE TABLE " + filename + " (\n" + strings.Join(types, ",\n") + "\n);"
	insertion := "INSERT INTO " + filename + " (" + strings.Join(header, ", ") + ") VALUES (" + strings.Join(empties, ", ") + ");"
	tableDefinition := TableDefinition{deletion: deletion, creation: creation, types: columnTypes, insertion: insertion}
	return tableDefinition
}

func mysqlDataType(s string) string {
	// check if the string can be parsed to an integer
	_, errInt := strconv.ParseInt(s, 10, 0)
	if errInt == nil {
		return "INTEGER"
	}

	// check if the string can be parsed to a floating-point number
	_, errFloat := strconv.ParseFloat(s, 64)
	if errFloat == nil {
		return "FLOAT"
	}

	// check if the string can be parsed to a boolean
	_, errBool := strconv.ParseBool(s)
	if errBool == nil {
		return "BOOLEAN"
	}

	// check if the string matches the 'YYYY-MM-DD' date format
	dateRegexp := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if dateRegexp.MatchString(s) {
		return "DATE"
	}

	// check if the string matches the 'YYYY-MM-DD HH:MM:SS' datetime format
	datetimeRegexp := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)
	if datetimeRegexp.MatchString(s) {
		return "DATETIME"
	}

	// check if the string matches the 'YYYY-MM-DD HH:MM:SS' datetime format
	datetimeRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if datetimeRegexp.MatchString(s) {
		return "DATETIME2"
	}

	// check if the string matches the 'HH:MM:SS' time format
	timeRegexp := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`)
	if timeRegexp.MatchString(s) {
		return "TIME"
	}

	// check if the string matches the 'YYYY' year format
	yearRegexp := regexp.MustCompile(`^\d{4}$`)
	if yearRegexp.MatchString(s) {
		return "YEAR"
	}
	// if none of the above checks passed, return TEXT as the default data type
	return "TEXT"
}
