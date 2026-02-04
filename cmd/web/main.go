package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lib/pq"
)

type Application struct {
	DB        *sql.DB
	Templates *template.Template
}

type Recipe struct {
	ID                                      int
	Title, Ingredients, Instructions        string
	Categories                              []uint8
}

func main() {
	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create table if not exists
	createTable(db)

	// Parse templates
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	app := &Application{
		DB:        db,
		Templates: tmpl,
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.Get("/", app.homeHandler)
	r.Get("/recipes", app.getRecipes)
	r.Post("/recipes", app.createRecipe)
	r.Delete("/recipes/{id}", app.deleteRecipe)
	r.Get("/health", healthHandler)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func createTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS recipes (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			ingredients TEXT,
			instructions TEXT,
			categories TEXT[]
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	app.Templates.ExecuteTemplate(w, "index.html", nil)
}

func (app *Application) getRecipes(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, title, ingredients, instructions, categories FROM recipes ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Title, &recipe.Instructions, &recipe.Ingredients, &recipe.Categories); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recipes = append(recipes, recipe)
	}

	app.Templates.ExecuteTemplate(w, "recipe-list.html", recipes)
}

func (app *Application) createRecipe(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	ingredients := r.FormValue("ingredients")
	instructions := r.FormValue("instructions")
	categories := []string{}
	if title == "" {
		http.Error(w, "Title required", http.StatusBadRequest)
		return
	}

	var id int
	err := app.DB.QueryRow(
		"INSERT INTO recipes (title, ingredients, instructions, categories) VALUES ($1, $2, $3, $4) RETURNING id",
		title,
		ingredients,
		instructions,
		pq.Array(categories),
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the recipes list
	app.getRecipes(w, r)
}

func (app *Application) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := app.DB.Exec("DELETE FROM recipes WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated list
	app.getRecipes(w, r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}