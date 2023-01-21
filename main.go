package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/snowflakedb/gosnowflake"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var conversation Conversation

type Conversation struct {
	fullMessage  string
	lastRequest  string
	lastResponse string
	schema       string
}

func getConfig() gosnowflake.Config {
	return gosnowflake.Config{
		Account:   "xxxxx",
		User:      "xxxxx",
		Password:  "xxxxx",
		Database:  "xxxxx",
		Warehouse: "xxxxx",
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

func record(message string) ResponseData {
	var data ResponseData
	// Make a http request to localhost:5001/chat?q= with the message, and send the response
	// URL encode the message
	urlEncoded := url.QueryEscape(message)
	url := "http://localhost:5001/record?q=" + urlEncoded
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

	bot, err := tgbotapi.NewBotAPI("XXXXX")
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
		conversation.fullMessage += "\n" + text

		sqlQuery := promptSqlExpert(conversation.fullMessage)

		sqlQuery = strings.ReplaceAll(sqlQuery, "tbooking", "dwh.mart_reporting.mart_booking")
		sqlQuery = strings.ReplaceAll(sqlQuery, "tcompany", "DWH.MART_REPORTING.MART_COMPANY")
		resultData := executeQuery(sqlQuery)
		csvFile := saveToCsv(resultData)
		sqlResult := strings.Join(resultData, "\n")
		answer := sqlResult

		conversationExpertTextRaw := explainSqlQueryExpert(sqlQuery)

		lastAnswer := conversationExpertTextRaw + "\n" + sqlResult

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, lastAnswer)
		toMp3 := explainQuestionAndAnswer(conversation.fullMessage, answer)
		mp3 := record(toMp3)
		uploadVoice(bot, update.Message.Chat.ID, mp3.Mp3)
		graph := createGraph(csvFile)
		fmt.Println("Graph file name is ", graph)

		if _, err := bot.Send(msg); err != nil {
			log.Print(err)
		}
	}
}

func createGraph(csvFileName string) string {
	// Create a new gnuplot command
	cmd := exec.Command("gnuplot", "-p")

	// Create a new pipe for the command's standard input
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ""
	}

	// Set the gnuplot command's script
	script := `set datafile separator ","
set term png
set output "graph.png"
plot "` + csvFileName + `" using 1:2 with lines
exit`

	// Start the command
	err = cmd.Start()
	if err != nil {
		return ""
	}

	// Write the script to the command's standard input
	_, err = stdin.Write([]byte(script))
	if err != nil {
		return ""
	}

	// Close the command's standard input
	stdin.Close()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return ""
	}

	return "graph.png"
}

func saveToCsv(data []string) string {
	// Create a new CSV file
	file, err := os.Create("data.csv")
	if err != nil {
		return ""
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the data to the CSV file
	for _, str := range data {
		err := writer.Write([]string{str})
		if err != nil {
			return ""
		}
	}
	return file.Name()
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
	header := strings.Join(cols, ",")
	result = append(result, header)
	result = append(result, strings.Repeat("-", len(header)))

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

func buildPrompt(question string) string {
	promptTables := `
	You are a snowflake DBA.  Follow these guidelines:
	1. Always use ilike when comparing text.  user name, CSM manager, company name...
	2. Never use email unless the field name contains 'EMAIL'
	3. When asked about hotel, flight, car bookings for a specific user always user BOOKING_USER
	4. When asked about Navan, add IS_REED_MACKAY_BOOKING='FALSE'
	5. Always remove duplicates.
	6. When counting, always use DISTINCT on what you are about to count
	7. When you are asked to add something, you need to add it to the select statement
	8. When dividing, always use NULLIFZERO
	9. Always end an SQL SELECT query with a ; char
    10. Always limit the select query to maximum of 20 records or less, as needed by the question, but never more than 20 records.
	INPUT: Ilan Twig's bookings
	OUTPUT: BOOKING_USER ilike '%Ilan Twig%'
		###Snowflake tables,with their properties:
		#
		#table tbooking, columns=[COMPANY_UUID use DISTINCT for counting,AGENCY_UUID,AGENCY_PCC,AGENCY_NAME,REPORT_DATETIME use this for booking date,REPORT_DATETIME_PST use this for booking date in pacific timezone,ACTIVE_BOOKING,BAGGAGE_FEE,BASE_PRICE use this for total calculating travel spend.  it is in USD,BASE_PRICE_LOCAL,BOOKING_CURRENCY,BOOKING_ID,BOOKING_IMPERSONATOR_PERSON_UUID,BOOKING_STATUS,CAR_VEHICLE_TYPE,CAR_DROPOFF_DATETIME,CHECK_IN_DATETIME,CHECK_OUT_DATETIME,CONFIRMATION_NUMBER,CREDIT_CARD_LAST4_DIGITS,CREDIT_CARD_TYPE,DATE_CANCELLED,DATE_MODIFIED,DESTINATION_CITY,DESTINATION_COUNTRY,DESTINATION_STATE,DIM_HOTEL_CITY,END_DATE,ESTIMATED_SPEND,EXCHANGE_FEE,FAIR_PRICE,FEE_TYPE,FLIGHT_OUTBOUND_ARRIVAL_DATETIME,FLIGHT_CABIN_CLASS use this for flight class,FLIGHT_CARRIER_LOCATOR_LIST,FLIGHT_CARRIER,FLIGHT_OUTBOUND_DEPARTURE_DATETIME,FLIGHT_MAIN_FARE_CLASS,FLIGHT_MAJOR_AIRLINE_CODE,FLIGHT_NUMBER_LIST,GDS_FEES,GROSS_SAVING_PERCENTAGE,HAS_HOTEL_BREAKFAST_INCLUDED,HAS_VALID_START_DATE,HOTEL_CORPORATE_DISCOUNT,HOTEL_INVENTORY_PROVIDER,HOTEL_NAME,HOTEL_PARENT_CHAIN_NAME,EXPEDIA_HOTEL_CHAIN,EXPEDIA_PARENT_CHAIN_NAME,HOTEL_PAYMENT_SCHEDULE,HOTEL_SABRE_CHAIN_CODE,HOTEL_SABRE_RATE_ACCESS_CODE,HOTEL_STARS_RATING,HOTEL_UUID,HOTEL_ROOM_GROUP_NAME,IS_AFFILIATE_PURCHASE_WITH_POINTS,IS_AGENT_MADE_BOOKING,IS_BOOKED_FOR_SELF,IS_BOOKING_WITH_REWARD,IS_BOOKING_WITH_SAVING,IS_OUT_OF_POLICY,IS_PERSONAL,IS_REFUNDABLE,IS_REED_MACKAY_BOOKING,IS_REED_MACKAY_ME_BOOKING,IS_REED_MACKAY_VIP_BOOKING,IS_TRANSACTED_ON_TRIPACTIONS,IS_TRIPACTIONS_LODGING_COLLECTION_RATE,IS_NEGOTIATED_DISCOUNT,NEGOTIATED_DISCOUNT_TYPE,LONG_HAUL_DURATION_THRESHOLD,LONGEST_FLIGHT_DURATION,NET_SAVING_PERCENTAGE,NET_SAVING,ORIGIN_CITY,ORIGIN_COUNTRY,ORIGIN_STATE,POLICY_VIOLATION_REASON,PRICE_BEFORE_FEE_LOCAL,PRICE_BEFORE_FEE,PRICE_TO_BEAT,PROVIDER,RESORT_FEE,REVENUE_PROVIDER,REVENUE,REWARD_VALUE,SAVING,SEATS_FEES_BASE,SEATS_FEES_TAX_LOCAL,SEATS_FEES_TAX,SOURCE_CURRENCY_CONVERSION,SOURCE_CURRENCY,START_DATE,TAX_LOCAL,TAX,TOTAL_PRICE_LOCAL,TOTAL_PRICE,TRAVEL_DURATION,TRIP_FEE_LOCAL,TRIP_FEE,TRIP_PURPOSE_GROUP,TRIP_ROUTE_TYPE,VALID_BOOKING,VENDOR,VIP_FEE,BOOKING_TRIP_PURPOSE,BOOKING_TRIP_NAME,BOOKING_TYPE can be flight/hotel/car/black car. always use ilike with this field,FRONT_END_REVENUE,ESTIMATE_PERCENTAGE_YIELD_OF_BOOKING_VALUE,ESTIMATED_TOTAL_BOOKING_REVENUE,TOTAL_ROOM_NIGHTS,LEAD_TIME_DAYS,IS_TRIP_INVITE,IS_TEAM_EVENT_INVITE indicates that this was a team event booking,TIME_TO_BOOK_MINUTES,DEPARTMENT,BOOKING_USER_EMAIL,BOOKING_USER always use ilike with this field.  Use it for questions about booking by user name,BOOKING_USER_GENDER,IS_ADMIN use to know if the user or booker is admin,IS_BUYER,IS_COMPANY_DELEGATE,BOOKING_USER_IS_EXECUTIVE_ASSISTANT,IS_INFLUENCER,IS_PROGRAM_MANAGER,IS_ROAD_WARRIOR,IS_SUPER_ADMIN,IS_VIP,POLICY_LEVEL,REGION,SUBSIDIARY,BOOKING_USER_TITLE,TRAVEL_FREQUENCY,BOOKING_USER_MANAGER,booking_user_is_executive_assistant,TRAVELLER_LEGAL_ENTITY_NAMES,TRAVELLER_LEGAL_ENTITY_UUIDS,TRAVELLER_MANAGER_EMAILS,TRAVELLER_OFFICE_COUNTRIES,TRAVELLER_OFFICE_LEGAL_NAMES,TRAVELLER_USERS_EMAILS,TRAVELLER_USERS,ACCOUNT_EXECUTIVE_MANAGER,ACCOUNT_EXECUTIVE,ACCOUNT_ID,ACCOUNT_OWNER,ACCOUNT_OWNER_EMAIL,ACCOUNT_STATUS,COMPANY_NAME always use ilike with this field,COMPANY_STATUS,GLOBAL_CSM_DIRECTOR,GLOBAL_CSM_DIRECTOR_EMAIL,GLOBAL_CSM_DIVISION,GLOBAL_CSM_MANAGER,GLOBAL_CSM_MANAGER_EMAIL,GLOBAL_CSM,GLOBAL_CSM_EMAIL,INDUSTRY,IS_REWARDS_ENABLED,MARKET_SEGMENT,PREVIOUS_TRAVEL_MGMT,REFERRER,SALESFORCE_ACCOUNT_TYPE,AFFILIATE_PRICING_MODEL,DOMAIN,IS_AFFILATE_PROGRAM_ENABLED,IS_COMPANY_LHG_SELF_ONBOARDING,LAUNCH_DATE,NUM_EMPLOYEES,REGIONAL_CSM_UUID,REGIONAL_CSM,REGIONAL_CSM_EMAIL,TRIPACTIONS_REGION,AGENCY_COUNTRY_NAME always use ilike,SABRE_FLIGHT_EXTRAS_TOTAL_PRICE,SABRE_NUMBER_OF_FLIGHT_CHANGES,ORIGIN_CONTINENT,ORIGIN_REGION,DESTINATION_CONTINENT,DESTINATION_REGION,POLICY_TYPE,MAX_ROOM_PRICE_PER_NIGHT_ALLOWED,MAX_PRICE_ALLOWED,COMPANY_POLICY_TYPE,COMPANY_BILLING_COUNTRY_CODE,COMPANY_OPPORTUNITY_CLOSE_DATE,IS_RM_ACCOUNT,COMPANY_BILLING_COUNTRY_NAME,IS_LIQUID,IS_VIP_BOOKING,REGIONAL_CSM_MANAGER,GLOBAL_CSM_MANAGER,flight_booking_distance_miles,airline_route];
		#
		### `

	return promptTables + question + "\nSELECT"
}

func promptSqlExpert(question string) string {
	fmt.Println("Full questions:\n", question)
	prompt := buildPrompt(question)
	ctx := context.Background()
	client := gpt3.NewClient(os.Getenv("API_KEY"))
	response, err := client.CompletionWithEngine(ctx, "code-davinci-002", gpt3.CompletionRequest{
		Prompt:      []string{prompt},
		Temperature: gpt3.Float32Ptr(0),
		MaxTokens:   gpt3.IntPtr(300),
		Stop:        []string{"#", ";"},
	})

	if err != nil {
		log.Fatalln(err)
	}

	return "SELECT " + response.Choices[0].Text
}

func explainSqlQueryExpert(whatToExplain string) string {
	trimmed := whatToExplain[:int(math.Min(float64(len(whatToExplain)), float64(300)))]
	prompt := "You are a high school teacher.  Explain what the blow query means:" +
		" in the most simple way.\nQuery:" + trimmed + "\n"
	return explain(prompt)
}

func explainQuestionAndAnswer(question string, answer string) string {
	prompt := fmt.Sprintf(`
You are a high school teacher.
User asked the question: "%s"
The answer is: "%s"
Please give a simple explanation: 
`, question, answer)
	return explain(prompt)
}

func explain(prompt string) string {
	fmt.Println("Prompt to GPT3: ", prompt)
	ctx := context.Background()
	client := gpt3.NewClient(os.Getenv("API_KEY"))
	response, err := client.CompletionWithEngine(ctx, "text-davinci-003", gpt3.CompletionRequest{
		Prompt:           []string{prompt},
		Temperature:      gpt3.Float32Ptr(0.5),
		TopP:             gpt3.Float32Ptr(1.0),
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		MaxTokens:        gpt3.IntPtr(300),
	})

	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Answer from GPT3: ", response.Choices[0].Text)

	return response.Choices[0].Text
}
