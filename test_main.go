package main

import (
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"
)

func TestCreateTable(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		t.Errorf("createTable failed: %v", err)
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='currency_data'")
	if err != nil {
		t.Fatal("Failed to query database:", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Error("Table 'currency_data' does not exist in the database")
	}
}

func TestSaveToDB(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	testData := Response{
		Data: map[string]float32{
			"EUR": 1.2,
			"USD": 1.0,
		},
		Time:       1616599029,
		Currencies: nil,
	}

	err = saveToDB(db, testData)
	if err != nil {
		t.Errorf("saveToDB failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM currency_data").Scan(&count)
	if err != nil {
		t.Fatal("Failed to query database:", err)
	}
	if count != len(testData.Data) {
		t.Errorf("Expected %d rows in currency_data table, got %d", len(testData.Data), count)
	}
}

func TestGetExchangeRate(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}
	_, err = db.Exec("INSERT INTO currency_data(currency, rate, updated_at) VALUES('USD', 1.0, ?)", time.Now())
	if err != nil {
		t.Fatal("Failed to insert test data:", err)
	}

	rate, err := getExchangeRate(db, "USD")
	if err != nil {
		t.Errorf("getExchangeRate failed: %v", err)
	}
	if rate != 1.0 {
		t.Errorf("Expected exchange rate of 1.0 for USD, got %f", rate)
	}
}

func TestUpdateExchangeRates(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	ch := make(chan struct{})

	go updateExchangeRates(db, ch)

	time.Sleep(1 * time.Second)

	ch <- struct{}{}

}

func TestGetUserInput(t *testing.T) {

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal("Failed to create pipe:", err)
	}
	os.Stdin = r

	input := "USD\n"

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ch := make(chan string)
		go getUserInput(ch, nil, &wg)

		_, err := w.WriteString(input)
		if err != nil {
			t.Fatal("Failed to write to pipe:", err)
		}

		w.Close()

		currency := <-ch

		if currency != "USD" {
			t.Errorf("Expected user input 'USD', got '%s'", currency)
		}
	}()

	wg.Wait()
}
