package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)


var response Response

type ModuleArgs struct {
	Host     string `json:"db_host"`
	Port     string `json:"port"`
	Database string `json:"db_name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Query    string `json:"query"`
}

type Response struct {
	Count int `json:"count,omitempty"`
	Results interface{} `json:"query_results,omitempty"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
}

func ExitJson(responseBody Response) {
	returnResponse(responseBody)
}

func FailJson(err error) {
	response.Failed = true
	response.Results = err.Error()
	returnResponse(response)
}

func returnResponse(responseBody Response) {
	var response []byte
	var err error
	response, err = json.Marshal(responseBody)
	if err != nil {
		response, _ = json.Marshal(Response{Results: "Invalid response object"})
	}
	fmt.Println(string(response))
	if responseBody.Failed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func createSQLClient(args ModuleArgs) *sql.DB {
	intPort, err := strconv.Atoi(args.Port)

	if err != nil {
		FailJson(err)
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", args.Username, args.Password, args.Host, intPort, args.Database)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		FailJson(err)
	}

	return db
}

func getCountResult(rows *sql.Rows) int {
	var count int
	var data []byte

	for rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			FailJson(err)
		}
		count, err = strconv.Atoi(string(data))
		if err != nil {
			FailJson(err)
		}
	}

	return count
}

func executeSQLQuery(db *sql.DB, args ModuleArgs) {
	rows, err := db.Query(args.Query)
	if err != nil {
		response.Failed = true
		response.Results = fmt.Sprintf("Query: %s  Error: %s", args.Query, err.Error())
		returnResponse(response)
	}

	defer rows.Close()

	if strings.Contains(args.Query, "count") {
		response.Count = getCountResult(rows)
		response.Changed = true
		returnResponse(response)
	}

	queryResultsToJSON(rows)

}

func queryResultsToJSON(rows *sql.Rows)  {

	columns, err := rows.Columns()
	if err != nil {
		FailJson(err)
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}


	var results []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			FailJson(err)
		}

		rowMap := make(map[string]interface{})
		for i, v := range values {
			x := v.([]byte)

			if nx, ok := strconv.ParseFloat(string(x), 64); ok == nil {
				rowMap[columns[i]] = nx
			} else if b, ok := strconv.ParseBool(string(x)); ok == nil {
				rowMap[columns[i]] = b
			} else if "string" == fmt.Sprintf("%T", string(x)) {
				rowMap[columns[i]] = string(x)
			} else {
				fmt.Printf("Failed on if for type %T of %v\n", x, x)
			}

		}

		results = append(results, rowMap)
	}

	if err != nil {
		FailJson(err)
	}

	response.Results = results
	response.Changed = true
	response.Failed = false
	ExitJson(response)
}



func main() {
	response = Response{}

	if len(os.Args) != 2 {
		FailJson(fmt.Errorf("No argument file provided"))
	}

	argsFile := os.Args[1]

	text, err := ioutil.ReadFile(argsFile)
	if err != nil {
		FailJson(fmt.Errorf("Could not read configuration file: " + argsFile))
	}

	var moduleArgs ModuleArgs
	err = json.Unmarshal(text, &moduleArgs)
	if err != nil {
		FailJson(fmt.Errorf("Configuration file not valid JSON: " + argsFile + " Data: %s Error: %s", string(text), err.Error()))
	}

	db := createSQLClient(moduleArgs)

	executeSQLQuery(db, moduleArgs)
}
