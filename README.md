# Still In Development

Rawdoggin

To run the app locally, set the followings:
- Database
- Environment Variables

And have these installed:
- Go 1.22.4 or higher
- Node LTS for TailwindCSS (Skip this if you just want to run the web server)
- PostgreSQL

Configure PostgreSQL (Linux with apt):
```sh
sudo apt install postgresql
sudo -u postgres psql
```
```sql
CREATE USER <username> WITH PASSWORD '<password>';
CREATE DATABASE <database_name>;
GRANT ALL PRIVILEGES ON DATABASE <database_name> TO <username>;
\q
```

Install Dependencies:
```sh
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get github.com/go-playground/validator/v10
go get github.com/joho/godotenv
go install github.com/air-verse/air@latest
```

Install Tailwind CSS (and Node with nvm):
```sh
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
nvm install --lts
npm install -D tailwindcss
```

Run database migration and seeding:
```sh
go run . --runTask=all
```

Run the dev server:
```sh
go run .
# Or with air
air -c .air.toml
```

Note:
1. I've only ran this on a debian based machine
