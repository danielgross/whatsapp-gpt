package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/snowflakedb/gosnowflake"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func getConfig() gosnowflake.Config {
	return gosnowflake.Config{
		Account:   "xxxx",
		User:      "xxxx",
		Password:  "xxxx",
		Database:  "xxxx",
		Warehouse: "xxxx",
	}
}

type ExistingSchema struct {
	schema []string
	tables []string
}

func convertToWav(path string) string {
	cmd := exec.Command("/opt/homebrew/bin/ffmpeg", "-y", "-i", path, path+".wav")
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return path + ".wav"
}

/*
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
			return
		}
	} else {
		response = sqlQueryFlow(msg)
	}

	schema := getSchema()

	if len(schema.tables) > 0 {
		message := fmt.Sprintf("This is the list of existing data you already have:\n%s\n%s\n",
			strings.Join(schema.tables, "\n"),
			"Let me process this...",
		)

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
*/
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

	bot, err := tgbotapi.NewBotAPI("5966095889:AAGu_EU88Som66FLvD2TftcnCP_pf1ypUH0")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		if update.Message == nil {
			continue
		}
		text := getAudioText(bot, update)
		if text == "" {
			text = update.Message.Text
		}

		responseFromGPT := talkToGPT(text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseFromGPT.Response)
		uploadVoice(bot, update.Message.Chat.ID, responseFromGPT.Mp3)

		if _, err := bot.Send(msg); err != nil {
			log.Print(err)
		}
	}
}

func uploadVoice(bot *tgbotapi.BotAPI, chatID int64, filePath string) {
	if len(filePath) == 0 {
		return
	}
	// Prepare the voice message
	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(filePath))

	// Send the voice message
	_, err := bot.Send(voice)
	if err != nil {
		log.Println(err)
		return
	}
}

func getAudioText(bot *tgbotapi.BotAPI, update tgbotapi.Update) string {
	file := downloadAudioFile(bot, update)
	if file == nil {
		return ""
	}

	wavFile := convertToWav(file.Name())
	if len(wavFile) == 0 {
		return ""
	}
	script := transcribe(wavFile)

	if len(script.Response) == 0 {
		return ""
	}
	return script.Response
}

func downloadAudioFile(bot *tgbotapi.BotAPI, update tgbotapi.Update) *os.File {
	if update.Message.Audio == nil && update.Message.Voice == nil {
		return nil
	}
	var fileName string
	var fileId string
	if update.Message.Audio != nil {
		fileName = update.Message.Audio.FileName
		fileId = update.Message.Audio.FileID
	} else {
		fileName = update.Message.Voice.FileID
		fileId = update.Message.Voice.FileID
	}
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		log.Println(err)
		return nil
	}
	url := file.Link(bot.Token)
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()
	filePath := fileName
	fileCreated, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer fileCreated.Close()
	_, err = io.Copy(fileCreated, response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	return fileCreated
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
