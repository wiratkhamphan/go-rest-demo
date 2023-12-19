package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

// Recipe is a struct representing a recipe.
type Recipe struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type recipeStore interface {
	Add(name string, recipe Recipe) error
	Get(name string) (Recipe, error)
	List() (map[string]Recipe, error)
	Update(name string, recipe Recipe) error
	Remove(name string) error
}

// MemStore is an in-memory implementation of the recipeStore interface.
type MemStore struct {
	recipes map[string]Recipe
}

// NewMemStore creates a new instance of the in-memory store.
func NewMemStore() recipeStore {
	return &MemStore{
		recipes: make(map[string]Recipe),
	}
}

// Implement the recipeStore interface methods for MemStore.

func (m *MemStore) Add(name string, recipe Recipe) error {
	if _, exists := m.recipes[name]; exists {
		return errors.New("recipe already exists")
	}
	m.recipes[name] = recipe
	return nil
}

func (m *MemStore) Get(name string) (Recipe, error) {
	recipe, exists := m.recipes[name]
	if !exists {
		return Recipe{}, NotFoundErr
	}
	return recipe, nil
}

func (m *MemStore) List() (map[string]Recipe, error) {
	return m.recipes, nil
}

func (m *MemStore) Update(name string, recipe Recipe) error {
	if _, exists := m.recipes[name]; !exists {
		return NotFoundErr
	}
	m.recipes[name] = recipe
	return nil
}

func (m *MemStore) Remove(name string) error {
	if _, exists := m.recipes[name]; !exists {
		return NotFoundErr
	}
	delete(m.recipes, name)
	return nil
}

// Define a custom NotFoundErr error.
var NotFoundErr = errors.New("not found")

// RecipesHandler is a handler for recipe-related operations.
type RecipesHandler struct {
	store recipeStore
}

// NewRecipesHandler creates a new instance of the RecipesHandler.
func NewRecipesHandler(store recipeStore) *RecipesHandler {
	return &RecipesHandler{store: store}
}

// Your existing Gin and route handling code...

func main() {
	// Create Gin router
	router := gin.Default()

	// Instantiate recipe Handler and provide a data store implementation
	store := NewMemStore() // Assuming you have a NewMemStore() function
	recipesHandler := NewRecipesHandler(store)

	// Register Routes
	router.GET("/", homePage)
	router.GET("/recipes", recipesHandler.ListRecipes)
	router.POST("/recipes", recipesHandler.CreateRecipe)
	router.GET("/recipes/:id", recipesHandler.GetRecipe)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipe)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipe)

	// Start the server
	router.Run(":8080")
}

func (h RecipesHandler) CreateRecipe(c *gin.Context) {
	// Get request body and convert it to Recipe (not recipes.Recipe)
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a URL-friendly name
	id := slug.Make(recipe.Name)

	// Add to the store
	h.store.Add(id, recipe)

	// Return success payload
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
func (h RecipesHandler) ListRecipes(c *gin.Context) {
	// Call the store to get the list of recipes
	r, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// Return the list, JSON encoding is implicit
	c.JSON(200, r)
}
func (h RecipesHandler) GetRecipe(c *gin.Context) {
	// Retrieve the URL parameter
	id := c.Param("id")

	// Get the recipe by ID from the store
	recipe, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	// Return the recipe, JSON encoding is implicit
	c.JSON(200, recipe)
}
func (h RecipesHandler) UpdateRecipe(c *gin.Context) {
	// Get request body and convert it to Recipe (not recipes.Recipe)
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve URL parameter
	id := c.Param("id")

	// Call the store to update the recipe
	err := h.store.Update(id, recipe)
	if err != nil {
		if err == NotFoundErr {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success payload
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h RecipesHandler) DeleteRecipe(c *gin.Context) {
	// Retrieve URL parameter
	id := c.Param("id")

	// Call the store to delete the recipe
	err := h.store.Remove(id)
	if err != nil {
		if err == NotFoundErr {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success payload
	c.JSON(http.StatusOK, gin.H{"status": "success"})

}
func homePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the home page"})
}
