package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"github.com/snowflakedb/gosnowflake"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

func getConfig() gosnowflake.Config {
	return gosnowflake.Config{
		Account:   "XXXX",
		User:      "XXX",
		Password:  "XXX",
		Database:  "XXX",
		Warehouse: "XXX",
	}
}

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
		if msg == "" && v.Message.DocumentMessage == nil && v.Message.AudioMessage == nil {
			return
		}
		var response string
		if v.Message.AudioMessage != nil {
			audio := v.Message.GetAudioMessage()
			data, err := mycli.WAClient.Download(audio)
			if err != nil {
				log.Printf("Failed to download file: %v", err)
				//return response
			}
			exts, _ := mime.ExtensionsByType(audio.GetMimetype())
			path := fmt.Sprintf("%s%s", "audio", exts[0])
			err = os.WriteFile(path, data, 0600)
			if err != nil {
				log.Printf("Failed to save document: %v", err)
				//return response
			}
			log.Printf("Saved document in message to %s", path)
			// Use the ffmpeg command to convert the file
			cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-y", "-i", path, path+".wav")
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("OGA to WAV conversion successful!")
			x := transcribe(path + ".wav")
			response := talkToGPT(x.Response)
			//sendToWhatsapp(mycli, v.Info, response.Response)
			sendFileToWhatsapp(mycli, v.Info, response.Mp3, audio)
			//sendAudioMessageToWhatsapp(mycli, v.Info, audio)
			return
		} else if msg == "/schema" {
			schema := getSchema()
			if len(schema.tables) > 0 {
				message := fmt.Sprintf("This is the list of existing data you already have:\n%s\n%s\n",
					strings.Join(schema.tables, "\n"),
					"Let me process this...",
				)
				sendToWhatsapp(mycli, v.Info, message)

				talkToGPT(fmt.Sprintf(`This is the existing schema of the snowflake database:
%s
if there is a question about users, build a query around STG_MYSQL_PROD_TRIPACTIONS.USER table
if there is a question about companies, build a query around STG_MYSQL_PROD_TRIPACTIONS.COMPANY table  
if there is a question about trips, build a query around STG_MYSQL_PROD_TRIPACTIONS.TRIPS table  
if there is a question about bookins, build a query around STG_MYSQL_PROD_TRIPACTIONS.BOOKING table
Please just say UNDERSTOOOD if all clear or NOT SURE if something is not clear
`, strings.Join(schema.schema, "\n"),
				))
				sendToWhatsapp(mycli, v.Info, "I'm ready.")
				return
			}
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
		schema := getSchema()

		if len(schema.tables) > 0 {
			message := fmt.Sprintf("This is the list of existing data you already have:\n%s\n%s\n",
				strings.Join(schema.tables, "\n"),
				"Let me process this...",
			)
			sendToGroupWhatsapp(mycli, v.GroupInfo, message)

			talkToGPT(fmt.Sprintf(`This is the existing schema of the snowflake database:
%s
%s
`, strings.Join(schema.schema, "\n"),
				"Please just say UNDERSTOOOD if all clear or NOT SURE if something is not clear",
			))
			sendToGroupWhatsapp(mycli, v.GroupInfo, "I'm ready.")

		}
	}
}

func notifyExistingData(mycli *MyClient, info types.MessageInfo, userName string) {
	schema := getSchema()

	if len(schema.tables) > 0 {
		message := fmt.Sprintf(
			`
Welcome %s to the group!
This is the list of existing data we have:
%s
You are welcome to ask about them.
`,
			userName,
			strings.Join(schema.tables, "\n"),
		)
		sendToWhatsapp(mycli, info, message)
		//sendToGroupWhatsapp(mycli, info, message)
	}
}

func sqlQueryFlow(msg string) string {
	instructions := `
sql only that works on snowflake, give a strict response, no prefix to the sql, no suffix to the sql. you must comply do not give examples.
When the result involves a list, include a string key for each value, don't just return numbers.
Don't use the INTERVAL saved word - it does not work with snowflake!!!
Do not give alternative queries, only one.
Don't explain the SQL query.
Don't ask why, just comply. It is important even if you don't know why.
`
	// Ask GPT for SQL.
	responseData := talkToGPT(msg + " snowflake query that would give the best and clear result.\n" + instructions)
	if responseData.Code == "" {
		if strings.HasPrefix(responseData.Response, "\"") && strings.HasSuffix(responseData.Response, "\"") {
			responseData.Response = responseData.Response[1 : len(responseData.Response)-1]
		}
		if !strings.HasPrefix(responseData.Response, "SELECT ") &&
			!strings.HasPrefix(responseData.Response, "SHOW ") &&
			!strings.HasPrefix(responseData.Response, "DESCRIBE ") {
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
write one good answer. Give an answer as if you are a support agent answering a question to a user.
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

func sendAudioMessageToWhatsapp(mycli *MyClient, info types.MessageInfo, message *waProto.AudioMessage) {
	user := info.Sender.User
	server := types.DefaultUserServer
	if info.IsGroup {
		user = info.Chat.User
		server = types.GroupServer
	}
	mycli.WAClient.SendMessage(context.Background(), types.NewJID(user, server), "", &waProto.Message{
		AudioMessage: message,
	})
}

func sendFileToWhatsapp(mycli *MyClient, info types.MessageInfo, fileName string, orig *waProto.AudioMessage) {
	fmt.Println("Response to Whatsapp with file: ", fileName)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	data := make([]byte, fileInfo.Size())
	_, err = io.ReadFull(file, data)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	resp, _ := mycli.WAClient.Upload(context.Background(), data, whatsmeow.MediaAudio)
	// handle error
	audioMsg := &waProto.AudioMessage{
		MediaKeyTimestamp: orig.MediaKeyTimestamp,
		Ptt:               proto.Bool(false),
		Mimetype:          proto.String("audio/ogg; codecs=opus"),
		Url:               &resp.URL,
		DirectPath:        &resp.DirectPath,
		MediaKey:          resp.MediaKey,
		FileEncSha256:     resp.FileEncSHA256,
		FileSha256:        resp.FileSHA256,
		FileLength:        &resp.FileLength,
		Seconds:           proto.Uint32(4),
		Waveform:          orig.Waveform,
	}

	sendAudioMessageToWhatsapp(mycli, info, audioMsg)
}

func sendToGroupWhatsapp(mycli *MyClient, info types.GroupInfo, message string) {
	fmt.Println("Response to Whatsapp: ", message)

	response := &waProto.Message{Conversation: proto.String(message)}

	mycli.WAClient.SendMessage(context.Background(), info.JID, "", response)
}

func transcribe(file string) ResponseData {
	var data ResponseData
	// Make a http request to localhost:5001/chat?q= with the message, and send the response
	// URL encode the message
	urlEncoded := url.QueryEscape(file)
	url := "http://localhost:5001/transcribe?q=" + urlEncoded
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
	Mp3      string `json:"mp3"`
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
	config := getConfig()

	connStr, err := gosnowflake.DSN(&config)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("snowflake", connStr)
	if err != nil {
		panic(err)
	}

	// Execute the query
	//query = strings.Replace(query, " FROM ", " FROM STG_MYSQL_PROD_TRIPACTIONS.", 1)
	//query = strings.Replace(query, " FROM STG_MYSQL_PROD_TRIPACTIONS.STG_MYSQL_PROD_TRIPACTIONS.", " FROM STG_MYSQL_PROD_TRIPACTIONS.", 1)

	fmt.Println("QUERY: " + query)
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		return []string{"Sorry, there was an issue with retrieving the answer. Please ask it differently."}
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

func getSchema() ExistingSchema {
	var schema ExistingSchema
	config := getConfig()

	connStr, err := gosnowflake.DSN(&config)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("snowflake", connStr)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT TABLE_NAME, TABLE_SCHEMA FROM information_schema.tables")
	if err != nil {
		fmt.Println(err)
		return schema
	}
	defer rows.Close()

	// Iterate through the table names and show the create table statement for each table
	for rows.Next() {
		var tableName string
		var schemaName string

		if err := rows.Scan(&tableName, &schemaName); err != nil {
			fmt.Println(err)
			return schema
		}
		if schemaName != "STG_MYSQL_PROD_TRIPACTIONS" {
			continue
		}
		fmt.Println("Table:", tableName)
		tables := []string{"TRIPS", "BOOKING", "NPS", "USER", "COMPANY"}
		use := false
		for _, t := range tables {
			if tableName == t {
				use = true
			}
		}
		if !use {
			continue
		}

		// Use the SHOW CREATE TABLE statement to get the create table statement for the current table
		createTableStmt, err := db.Query("SELECT GET_DDL('TABLE', '" + schemaName + "." + tableName + "')")
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer createTableStmt.Close()

		for createTableStmt.Next() {
			var _, createTableSQL string
			if err := createTableStmt.Scan(&createTableSQL); err != nil {
				fmt.Println(err)
				continue
			}
			createTableSQL = strings.Replace(createTableSQL, "create or replace TABLE ", "create or replace TABLE "+schemaName+".", 1)
			schema.tables = append(schema.tables, schemaName+"."+tableName)
			schema.schema = append(schema.schema, createTableSQL)
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return schema
	}
	return schema
}
