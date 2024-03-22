package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	baseURL           = "https://api.freecurrencyapi.com/v1/latest"
	defaultBase       = "USD"
	defaultCurrencies = "EUR,USD,RUB,GBP,JPY"
	dbPath            = "./currency.db"
	updateInterval    = 5 * time.Second
)

var apiKey = os.Getenv("API_KEY")

type Response struct {
	Data       map[string]float32 `json:"data"`
	Time       int64              `json:"time"`
	Currencies []CurrencyData     `json:"currencies"`
}

type CurrencyData struct {
	ID        int
	Currency  string
	Rate      float32
	UpdatedAt time.Time
}

func Latest(params map[string]string) (Response, error) {
	if params == nil {
		params = make(map[string]string)
	}
	if params["apikey"] == "" {
		params["apikey"] = apiKey
	}
	if params["base_currency"] == "" {
		params["base_currency"] = defaultBase
	}
	if params["currencies"] == "" {
		params["currencies"] = defaultCurrencies
	}

	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
	}
	queryString := strings.Join(queryParams, "&")

	response, err := http.Get(baseURL + "?" + queryString)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()

	var respData Response
	err = json.NewDecoder(response.Body).Decode(&respData)
	if err != nil {
		return Response{}, err
	}

	return respData, nil
}

func createTable(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS currency_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		currency TEXT,
		rate REAL,
		updated_at TIMESTAMP
	);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func saveToDB(db *sql.DB, data Response) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO currency_data(currency, rate, updated_at) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for currency, rate := range data.Data {
		_, err = stmt.Exec(currency, rate, time.Now())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func getExchangeRate(db *sql.DB, currency string) (float32, error) {
	var rate float32
	err := db.QueryRow("SELECT rate FROM currency_data WHERE currency = ?", currency).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func updateExchangeRates(db *sql.DB, ch chan<- struct{}) {
	for {
		<-time.After(updateInterval)

		latest, err := Latest(nil)
		if err != nil {
			log.Println("Error getting latest rates:", err)
			continue
		}

		err = saveToDB(db, latest)
		if err != nil {
			log.Println("Error saving data to database:", err)
			continue
		}

		ch <- struct{}{}
	}
}

func getUserInput(ch chan<- string, db *sql.DB, wg *sync.WaitGroup) {
	for {
		var currency string
		fmt.Print("Enter currency code (e.g., EUR): ")
		fmt.Scanln(&currency)
		currency = strings.TrimSpace(currency) // Trim whitespace

		// If the currency is empty, skip further processing
		if currency == "" {
			fmt.Println("Currency code cannot be empty. Please try again.")
			continue
		}

		ch <- currency

		// Wait for the goroutine to finish
		wg.Wait()

		// Ask user if they want to check another currency
		fmt.Print("Enter another currency code? (yes/no): ")
		var choice string
		fmt.Scanln(&choice)
		if strings.ToLower(choice) != "yes" {
			break
		}
	}
	fmt.Println("Latest rates updated successfully at", time.Now().Format("2006/01/02 15:04:05"))
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	ch := make(chan string)
	wg := &sync.WaitGroup{}
	go getUserInput(ch, db, wg)

	fmt.Println("Press Ctrl+C to exit.")
	for {
		select {
		case currency := <-ch:
			wg.Add(1)
			go func(curr string) {
				rate, err := getExchangeRate(db, curr)
				if err != nil {
					fmt.Printf("Error getting exchange rate for %s: %v\n", curr, err) // Use curr instead of currency
					return
				}
				fmt.Printf("Exchange rate for %s: %f\n", curr, rate)                     // Print exchange rate
				fmt.Printf("Updated at: %s\n", time.Now().Format("2006/01/02 15:04:05")) // Print update time

				// Signal that the update is complete
				wg.Done()
			}(currency)
		}
	}
}
