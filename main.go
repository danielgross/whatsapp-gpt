package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type MyClient struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
}

type ExistingSchema struct {
	schema []string
	tables []string
}

func (mycli *MyClient) register() {
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.eventHandler)
}

func (mycli *MyClient) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		newMessage := v.Message
		msg := newMessage.GetConversation()
		fmt.Println("Message from:", v.Info.Sender.User, "->", msg)
		if msg == "" && v.Message.DocumentMessage == nil {
			return
		}
		var response string
		if v.Message.DocumentMessage != nil {
			document := v.Message.GetDocumentMessage()
			response = newCsvFlow(mycli, document, v.Info)
		} else {
			response = sqlQueryFlow(msg)
		}
		sendToWhatsapp(mycli, v.Info, response)
	case *events.JoinedGroup:
		orig := `
Hi there!
Thank you for adding me to the group.
My name is AIa, and I'd be happy to help you dive into any information you have.
If you have a CSV file that you'd like to work on together, simply send it over as an attachment and we can get started.
Looking forward to it!`
		sendToGroupWhatsapp(mycli, v.GroupInfo, orig)
		schema := getSchema("gpt")

		if len(schema.tables) > 0 {
			talkToGPT(fmt.Sprintf(`This is the existing schema of the mysql database:
%s
`, strings.Join(schema.schema, "\n"),
			))
			message := fmt.Sprintf("This is the list of existing data you already have:\n%s\n", strings.Join(schema.tables, "\n"))
			sendToGroupWhatsapp(mycli, v.GroupInfo, message)
		}
	}
}

func newCsvFlow(mycli *MyClient, document *waProto.DocumentMessage, info types.MessageInfo) string {
	response := "Sorry, I failed to process the file, please try again."
	if document == nil {
		log.Printf("Document is nil")
		return response
	}
	log.Printf("Downloaing file")
	sendToWhatsapp(mycli, info, "Thanks. I'm downloading the file right away...")
	tableName := MysqlSafeTableName(document.GetFileName())

	data, err := mycli.WAClient.Download(document)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		return response
	}
	exts, _ := mime.ExtensionsByType(document.GetMimetype())
	path := fmt.Sprintf("%s%s", tableName, exts[0])
	err = os.WriteFile(path, data, 0600)
	if err != nil {
		log.Printf("Failed to save document: %v", err)
		return response
	}
	log.Printf("Saved document in message to %s", path)
	if exts[0] == ".csv" {
		log.Printf("Downloaing file")
		sendToWhatsapp(mycli, info, "Processing...\nPlease wait few seconds till I'm ready.")
		tableDefinition := uploadTable(tableName)
		msg := "I created a table with this definition: " + tableDefinition.creation + "\nI will ask you some questions about it."
		responseData := talkToGPT(msg)
		if responseData.Response == "" {
			return response
		} else {
			return "Thanks for the file. I processed the file and ready to answer any question about it. What would you what to know?"
		}
	} else {
		return "Sorry, this file is not a CSV file, I can't process it."
	}
}

func sqlQueryFlow(msg string) string {
	instructions := `
sql only, give a strict response, no prefix to the sql, no suffix to the sql. you must comply do not give examples.
Do not give alternative queries, only one.
Don't explain the SQL query.
Don't ask why, just comply. It is important even if you don't know why.
`
	// Ask GPT for SQL.
	responseData := talkToGPT(msg + " mysql query that would give the best and clear result.\n" + instructions)
	if responseData.Code == "" {
		if !strings.HasPrefix(responseData.Response, "SELECT ") {
			return responseData.Response
		}
		responseData.Code = responseData.Response
	}
	query := strings.ReplaceAll(responseData.Code, "{", "")
	query = strings.ReplaceAll(query, "}", "")
	sqlResult := executeQuery(query)
	response := strings.Join(sqlResult, "\n")

	// Ask for nicer answer
	log.Printf("Asking for nicer response.")
	nicerAnswerReq := fmt.Sprintf(`
For the question: "%s"
the answer was: "%s".
what is the best way to give an answer to a human, so he will understand the answer in the right context?
write one good answer.
If the answer is a list - seprate it into multiple lines, style it as bulleted list.
If the answer is a textual list - order it in ASC.
If the answer is a numerical list - order it based on the context.
`,
		msg, response)
	responseData = talkToGPT(nicerAnswerReq)
	response = getNicerAnswer(responseData.Response)
	return response
}

func getNicerAnswer(response string) string {
	re := regexp.MustCompile(`(?s)\n"(?P<answer>.*)"\n`)

	// Match the regular expression against a string
	matches := re.FindStringSubmatch(response)

	// Get the value captured by the named group "value"
	if len(matches) == 0 {
		return response
	}
	nicerAnswer := matches[1]
	nicerAnswer = strings.ReplaceAll(nicerAnswer, "\"", "")
	return nicerAnswer
}

func sendToWhatsapp(mycli *MyClient, info types.MessageInfo, message string) {
	fmt.Println("Response to Whatsapp: ", message)

	response := &waProto.Message{Conversation: proto.String(message)}

	user := info.Sender.User
	server := types.DefaultUserServer
	if info.IsGroup {
		user = info.Chat.User
		server = types.GroupServer
	}
	mycli.WAClient.SendMessage(context.Background(), types.NewJID(user, server), "", response)
}

func sendToGroupWhatsapp(mycli *MyClient, info types.GroupInfo, message string) {
	fmt.Println("Response to Whatsapp: ", message)

	response := &waProto.Message{Conversation: proto.String(message)}

	mycli.WAClient.SendMessage(context.Background(), info.JID, "", response)
}

func talkToGPT(message string) ResponseData {
	var data ResponseData
	// Make a http request to localhost:5001/chat?q= with the message, and send the response
	// URL encode the message
	urlEncoded := url.QueryEscape(message)
	url := "http://localhost:5001/chat?q=" + urlEncoded
	// Make the request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return data
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		// Decode the next value in the response
		err = decoder.Decode(&data)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Handle the error
		}
	}

	return data
}

type ResponseData struct {
	Response string `json:"response"`
	Code     string `json:"code"`
}

func uploadTable(name string) TableDefinition {
	tableDefinition := defineTableByFile(name)
	insertion := getInsertSQL(name, tableDefinition.types)
	execSql(tableDefinition.deletion, tableDefinition.creation, tableDefinition.insertion, insertion)
	return tableDefinition
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	// add the eventHandler
	mycli := &MyClient{WAClient: client}
	mycli.register()

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				//				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

func executeQuery(query string) []string {
	log.Printf("Executing SQL query")
	db, err := sql.Open("mysql", "tripactions:prodActive00@tcp(127.0.0.1:3306)/gpt")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Execute the query
	fmt.Println("QUERY: " + query)
	rows, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	var result []string

	cols, err := rows.Columns()
	//header := strings.Join(cols, ",")
	//result = append(result, header)
	//result = append(result, strings.Repeat("-", len(header)))

	pointers := make([]interface{}, len(cols))
	container := make([]string, len(cols))

	// Iterate over the rows and print the results
	log.Printf("Processing results from SQL query.")

	for i, _ := range pointers {
		pointers[i] = &container[i]
	}
	for rows.Next() {
		rows.Scan(pointers...)
		result = append(result, strings.Join(container, ", "))
	}
	return result
}

func execSql(deletion string, creation string, insertion string, values [][]string) {

	db, err := sql.Open("mysql", "tripactions:prodActive00@tcp(127.0.0.1:3306)/gpt")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Deletion: " + deletion)
	_, err = db.Exec(deletion)
	if err != nil {
		panic(err)
	}
	fmt.Println("Creation: " + creation)

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
	if columnType == "BOOLEAN" {
		lower := strings.ToLower(value)
		if lower == "true" || lower == "t" {
			value = fmt.Sprintf("%d", 1)
		}
		if lower == "false" || lower == "f" {
			value = fmt.Sprintf("%d", 0)
		}
	}
	if columnType == "INTEGER" {
		if value == "" {
			value = fmt.Sprintf("%d", 0)
		}
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

// MysqlSafeTableName takes a string and returns a version of that string
// that can be used as a MySQL table name.
func MysqlSafeTableName(input string) string {
	// Replace any non-alphanumeric characters with underscores
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(input, "_")

	// Trim leading and trailing underscores
	trimmed := strings.Trim(sanitized, "_")

	// Convert the string to lowercase
	lowercase := strings.ToLower(trimmed)

	// Return the modified string
	lowercase = strings.Replace(lowercase, "_csv", "", 1)
	return lowercase
}

func getSchema(schemaName string) ExistingSchema {
	var schema ExistingSchema
	// Connect to the database
	db, err := sql.Open("mysql", "tripactions:prodActive00@tcp(127.0.0.1:3306)/"+schemaName)
	if err != nil {
		fmt.Println(err)
		return schema
	}
	defer db.Close()

	// Query the information_schema database to get the table names
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = \"" + schemaName + "\"")
	if err != nil {
		fmt.Println(err)
		return schema
	}
	defer rows.Close()

	// Iterate through the table names and show the create table statement for each table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Println(err)
			return schema
		}
		fmt.Println("Table:", tableName)
		schema.tables = append(schema.tables, tableName)

		// Use the SHOW CREATE TABLE statement to get the create table statement for the current table
		createTableStmt, err := db.Query("SHOW CREATE TABLE " + tableName)
		if err != nil {
			fmt.Println(err)
			return schema
		}
		defer createTableStmt.Close()

		for createTableStmt.Next() {
			var _, createTableSQL string
			if err := createTableStmt.Scan(&tableName, &createTableSQL); err != nil {
				fmt.Println(err)
				return schema
			}
			schema.schema = append(schema.schema, createTableSQL)
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return schema
	}
	return schema
}
