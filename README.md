Task
This project aims to provide a Go utility/library for fetching currencies using exchangeratesapi.io. Please register on this site to obtain an API Key.

Dependencies
github.com/mattn/go-sqlite3: Go SQLite3 driver.
github.com/pressly/goose/v3/cmd/goose: Goose command-line tool for database migrations.
github.com/lib/pq: PostgreSQL driver for Go's database/sql package.
Setting Up
Ensure Go is installed on your system.

Copy the .env.example file to .env and replace API_KEY with your own API Key obtained from the exchangeratesapi.io website.

Install goose globally to run migrations:

bash

go install github.com/pressly/goose/v3/cmd/goose@latest
Start the PostgreSQL database using Docker. Make sure Docker is allowed to create the /var/db/data directory. Create the directory /var/db/data using:

bash

sudo mkdir -p /var/db/data
sudo chown -R $USER:$USER /var/db/data
docker compose up -d
Migrating the Database
To migrate the database, run the following command:

bash

make migrate-db
Running the Application
To fetch currency details from the API, save them to the database, and display the results, run:

bash

make run
This command executes the utility and provides the desired output.
