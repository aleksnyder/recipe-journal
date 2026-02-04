# Recipe Journal

**A recipe journal app using Htmx, Go, and PostgreSQL.**

## 🚀 Quick Start

### Local Development

**Prerequisites:**
- Go 1.21 or higher
- PostgreSQL (or use Docker)

**Steps:**
```bash
# Clone the repository
git clone https://github.com/aleksnyder/recipes-app.git
cd recipes-app

# Install dependencies
go mod download

# Set up PostgreSQL (or use Docker)
docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
export PORT=8080

# Run the application
go run cmd/web/main.go

# Open browser to http://localhost:8080
```

## 📄 License

MIT License - see LICENSE file for details
