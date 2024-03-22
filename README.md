![image](https://github.com/gauss2302/task_plata/assets/144123738/249942e5-b7f8-4985-b055-585c4ed8b05e)


**Task** <br />
This project aims to provide a Go utility/library for fetching currencies using exchangeratesapi.io. Please register on this site to obtain an API Key.
The requests for update go independently with goroutines.

**Dependencies** <br />
_github.com/mattn/go-sqlite3_: Go SQLite3 driver.
_github.com/pressly/goose/v3/cmd/goose_: Goose command-line tool for database migrations.
_github.com/lib/pq_: PostgreSQL driver for Go's database/sql package.

**Setting Up** <br />
Ensure Go is installed on your system.

Copy the .env.example file to .env and replace API_KEY with **your own API** Key obtained from the exchangeratesapi.io website.

Install goose globally to run migrations:

_go install github.com/pressly/goose/v3/cmd/goose@latest_

Start the PostgreSQL database using Docker. Make sure Docker is allowed to create the /var/db/data directory. Create the directory /var/db/data using:

_sudo mkdir -p /var/db/data
sudo chown -R $USER:$USER /var/db/data
docker compose up -d_

**Migrating the Database** <br />
To migrate the database, run the following command:

_make migrate-db_

_Running the Application_

To fetch currency details from the API, save them to the database, and display the results, run:

_go run main.go_

This command executes the utility and provides the desired output.
