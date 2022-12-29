package main

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	godotenv.Load()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalln("Missing API KEY")
	}

	ctx := context.Background()
	client := gpt3.NewClient(apiKey)

	resp, err := client.Completion(ctx, gpt3.CompletionRequest{
		//Prompt:    []string{"The first thing you should know about javascript is"},
		//Prompt: []string{"SELECT SUM(amount) FROM Salary_Payments WHERE date > '2008-01-01'\n# The right widget to show the result of this query is "},
		//Prompt: []string{"SELECT DISTINCT department_name FROM department, salary_payments WHERE department.id = salary_payments.department_id AND salary_payments.amount > 10000 AND salary_payments.date > DATE_SUB(CURDATE(), INTERVAL 3 MONTH) AND salary_payments.department_id IN (SELECT department_id FROM department WHERE department_name LIKE '%Sales%')\n# The right widget to show the results of this query is "},
		Prompt:           []string{"### MySQL tables, with their properties:\\n#\\n# Employee(id, name, department_id)\\n# Department(id, name, address)\\n# Salary_Payments(id, employee_id, amount, date)\\n#\\n### A query to show total salaries paid in the last year is: \\nSELECT"},
		MaxTokens:        gpt3.IntPtr(150),
		Stop:             []string{"#", ";"},
		Echo:             true,
		Temperature:      gpt3.Float32Ptr(0.0),
		TopP:             gpt3.Float32Ptr(1.0),
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(resp.Choices[0].Text)
}
