package recipe

import (
	"app/modules/db"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Day struct {
	Count   int      `json:"dayCount"`
	Recipes []Recipe `json:"meals"`
}

type Product struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Value    float64 `json:"value"`
	Size     string  `json:"size"`
	Calories int     `json:"calories"`
	Carbs    int     `json:"carbs"`
	Proteins int     `json:"proteins"`
}

type Recipe struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Category     int       `json:"category"`
	Time         int       `json:"time"`
	Image        string    `json:"image"`
	Instructions string    `json:"instructions"`
	Calories     int       `json:"calories"`
	Carbs        int       `json:"carbs"`
	Proteins     int       `json:"proteins"`
	Products     []Product `json:"products"`
}

type Ingredients struct {
	ID        int
	RecipeID  int     `json:"recipeId"`
	ProductID int     `json:"productId"`
	Value     float64 `json:"value"`
}

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/create", createRecipe)
	router.Get("/getAll", getAllRecipes)
	router.Get("/getById/{recipeId}", getRecipeById)
	router.Get("/checkTitle/{title}", checkIfTitleValid)
	return router
}

func createRecipe(w http.ResponseWriter, r *http.Request) {
	//session, _ := auth.Store.Get(r, "cookie")
	//if auth, ok := session.Values["auth"].(bool); !ok || !auth {
	//http.Error(w, "Forbidden", http.StatusForbidden)
	//return
	//}
	var recipe Recipe

	json.NewDecoder(r.Body).Decode(&recipe)

	if !checkIfValid(recipe) {
		http.Error(w, "Bad request", 400)
		return
	}

	db := db.InitDB()

	query, err := db.Prepare("INSERT INTO recipes(title, category, time, image, instructions) VALUES(?,?,?,?,?)")
	if err != nil {
		http.Error(w, "SQL insert error", 400)
		log.Print(err)
		return
	}

	res, err := query.Exec(recipe.Title, recipe.Category, recipe.Time, recipe.Image, recipe.Instructions)
	if err != nil {
		http.Error(w, "SQL insert error", 400)
		log.Print(err)
		return
	}

	recipeID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "SQL insert error", 400)
		log.Print(err)
		return
	}
	recipe.ID = int(recipeID)

	err = createRecipeIngredients(&recipe)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = sumRecipeNutrition(&recipe)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	err = updateRecipeNutrition(&recipe)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	render.JSON(w, r, "Recipe was created")
}

func getAllRecipes(w http.ResponseWriter, r *http.Request) {
	db := db.InitDB()
	result, err := db.Query("SELECT * FROM recipes")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	recipes := []Recipe{}
	for result.Next() {
		var recipe Recipe
		err := result.Scan(&recipe.ID, &recipe.Title, &recipe.Category, &recipe.Time, &recipe.Image, &recipe.Instructions, &recipe.Calories, &recipe.Carbs, &recipe.Proteins)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		recipes = append(recipes, recipe)
	}

	for idx, recipe := range recipes {
		result, er := db.Query("SELECT products.id, products.title, ingredients.value, products.calories, products.proteins, products.carbs, products.size FROM ingredients LEFT JOIN products ON ingredients.productId = products.id WHERE ingredients.recipeId = ?", recipe.ID)
		if er != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		for result.Next() {
			var product Product
			err := result.Scan(&product.ID, &product.Title, &product.Value, &product.Calories, &product.Proteins, &product.Carbs, &product.Size)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			recipes[idx].Products = append(recipes[idx].Products, product)
		}
	}
	render.JSON(w, r, recipes)
}

func getRecipeById(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "recipeId")

	db := db.InitDB()
	result, err := db.Query("SELECT * FROM recipes WHERE id = ?", recipeID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	recipe := Recipe{}
	for result.Next() {
		err = result.Scan(&recipe.ID, &recipe.Title, &recipe.Category, &recipe.Time, &recipe.Image, &recipe.Instructions, &recipe.Calories, &recipe.Carbs, &recipe.Proteins)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	result, err = db.Query("SELECT products.id, products.title, ingredients.value, products.calories, products.proteins, products.carbs, products.size FROM ingredients LEFT JOIN products ON ingredients.productId = products.id WHERE ingredients.recipeId = ?", recipe.ID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	for result.Next() {
		var product Product
		err := result.Scan(&product.ID, &product.Title, &product.Value, &product.Calories, &product.Proteins, &product.Carbs, &product.Size)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		recipe.Products = append(recipe.Products, product)
	}
	render.JSON(w, r, recipe)
}

func updateRecipeNutrition(recipe *Recipe) (err error) {
	db := db.InitDB()
	query, err := db.Prepare("UPDATE recipes SET calories = ?, carbs = ?, proteins = ? WHERE id = ?")
	if err == nil {
		_, err = query.Exec(recipe.Calories, recipe.Carbs, recipe.Proteins, recipe.ID)
	}
	defer db.Close()
	return err
}

func sumRecipeNutrition(recipe *Recipe) (err error) {
	db := db.InitDB()
	result, err := db.Query("SELECT products.title, ingredients.value, products.calories, products.proteins, products.carbs, products.size FROM ingredients LEFT JOIN products ON ingredients.productId = products.id WHERE ingredients.recipeId = ?", recipe.ID)
	defer db.Close()
	if err == nil {
		for result.Next() {
			var product Product
			err := result.Scan(&product.Title, &product.Value, &product.Calories, &product.Proteins, &product.Carbs, &product.Size)
			if err != nil {
				return err
			}
			recipe.Calories += product.Calories
			recipe.Carbs += product.Carbs
			recipe.Proteins += product.Proteins
		}
	}
	return err
}

func createRecipeIngredients(recipe *Recipe) (err error) {
	db := db.InitDB()
	for _, product := range recipe.Products {
		query, err := db.Prepare("INSERT INTO ingredients(recipeid, productid, value) VALUES(?,?,?)")
		if err == nil {
			_, err := query.Exec(recipe.ID, product.ID, product.Value)
			defer db.Close()
			if err != nil {
				return err
			}
		}
	}
	defer db.Close()
	return err
}

func checkIfTitleValid(w http.ResponseWriter, r *http.Request) {
	title := chi.URLParam(r, "title")
	var ifExists bool

	db := db.InitDB()
	err := db.QueryRow("SELECT IF(COUNT(*),'true','false') FROM recipes WHERE title = ?", title).Scan(&ifExists)
	defer db.Close()
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	if ifExists {
		http.Error(w, "Title exists", 400)
		return
	}

	render.JSON(w, r, "Title is valid")
}

func checkIfValid(r Recipe) bool {
	if len(r.Title) < 4 {
		return false
	}
	if r.Category == 0 {
		return false
	}
	if r.Time == 0 {
		return false
	}
	if len(r.Image) < 4 {
		return false
	}
	if len(r.Instructions) < 10 {
		return false
	}
	if len(r.Products) < 2 {
		return false
	}
	return true
}
