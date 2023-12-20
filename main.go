package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Recipe คือโครงสร้างที่แทนสูตรอาหาร
type Recipe struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// recipeStore คือ interface ที่กำหนดวิธีการจัดการกับข้อมูลของ Recipe
type recipeStore interface {
	Add(name string, recipe Recipe) error
	Get(name string) (Recipe, error)
	List() (map[string]Recipe, error)
	Update(name string, recipe Recipe) error
	Remove(name string) error
}

// MySQLStore เป็น implement ของ recipeStore ที่ใช้ MySQL
type MySQLStore struct {
	db *sql.DB
}

// DBConnection ทำการเชื่อมต่อกับฐานข้อมูล MySQL
func DBConnection() (*sql.DB, error) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "web_lek"

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		return nil, err
	}

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewMySQLStore สร้าง instance ใหม่ของ MySQL store
func NewMySQLStore(db *sql.DB) recipeStore {
	return &MySQLStore{db: db}
}

// นิยาม method ของ interface recipeStore สำหรับ MySQLStore

// Add เพิ่ม Recipe เข้าสู่ฐานข้อมูล
func (m *MySQLStore) Add(name string, recipe Recipe) error {
	_, err := m.db.Exec("INSERT INTO recipe (name, description) VALUES (?, ?)", name, recipe.Description)
	return err
}

// Get ดึงข้อมูล Recipe จากฐานข้อมูล
func (m *MySQLStore) Get(name string) (Recipe, error) {
	var recipe Recipe
	err := m.db.QueryRow("SELECT name, description FROM recipe WHERE name = ?", name).Scan(&recipe.Name, &recipe.Description)
	if err != nil {
		return Recipe{}, ErrNotFound
	}
	return recipe, nil
}

// List ดึงรายการ Recipe ทั้งหมดจากฐานข้อมูล
func (m *MySQLStore) List() (map[string]Recipe, error) {
	rows, err := m.db.Query("SELECT name, description FROM recipe")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := make(map[string]Recipe)
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.Name, &recipe.Description)
		if err != nil {
			return nil, err
		}
		recipes[recipe.Name] = recipe
	}

	return recipes, nil
}

// Update อัพเดตข้อมูล Recipe ในฐานข้อมูล
func (m *MySQLStore) Update(name string, recipe Recipe) error {
	result, err := m.db.Exec("UPDATE recipe SET description = ? WHERE name = ?", recipe.Description, name)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Remove ลบ Recipe จากฐานข้อมูล
func (m *MySQLStore) Remove(name string) error {
	result, err := m.db.Exec("DELETE FROM recipe WHERE name = ?", name)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// นิยาม error ที่ custom ชื่อ NotFoundErr
var ErrNotFound = errors.New("not found")

// RecipesHandler เป็น handler สำหรับตัวดำเนินการที่เกี่ยวกับ recipe
type RecipesHandler struct {
	store recipeStore
}

// NewRecipesHandler สร้าง instance ใหม่ของ RecipesHandler
func NewRecipesHandler(store recipeStore) *RecipesHandler {
	return &RecipesHandler{store: store}
}

// main เป็นฟังก์ชันหลักที่ทำการสร้างเซิร์ฟเวอร์และกำหนด route
func main() {
	// สร้าง Gin router
	router := gin.Default()

	// สร้าง MySQL store และให้ implement ข้อมูลของ store
	db, err := DBConnection()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	recipesHandler := NewRecipesHandler(store)

	// ลงทะเบียน Routes
	router.GET("/", homePage)
	router.GET("/recipes", recipesHandler.ListRecipes)
	router.POST("/recipes", recipesHandler.CreateRecipe)
	router.GET("/recipes/:id", recipesHandler.GetRecipe)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipe)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipe)

	// เริ่มเซิร์ฟเวอร์
	router.Run(":8080")
	if err != nil {
		panic(err)
	}
}

// homePage คือ handler สำหรับ route หน้าแรก
func homePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the home page"})
}

// ListRecipes คือ handler สำหรับดึงรายการสูตรอาหารทั้งหมด
func (h *RecipesHandler) ListRecipes(c *gin.Context) {
	// เรียกใช้ store เพื่อดึงรายการสูตรอาหาร
	recipes, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งรายการสูตรอาหารกลับไป
	c.JSON(http.StatusOK, recipes)
}

// CreateRecipe คือ handler สำหรับเพิ่มสูตรอาหารใหม่
func (h *RecipesHandler) CreateRecipe(c *gin.Context) {
	// ดึง request body และแปลงเป็นโครงสร้าง Recipe
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// เพิ่มสูตรอาหารใหม่
	err := h.store.Add(recipe.Name, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งผลลัพธ์สำเร็จกลับ
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetRecipe คือ handler สำหรับดึงข้อมูลสูตรอาหารจาก ID
func (h *RecipesHandler) GetRecipe(c *gin.Context) {
	// ดึงพารามิเตอร์ URL
	id := c.Param("id")

	// ดึงข้อมูลสูตรอาหารจาก store ด้วย ID
	recipe, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลสูตรอาหารกลับไป
	c.JSON(http.StatusOK, recipe)
}

// UpdateRecipe คือ handler สำหรับอัปเดตข้อมูลสูตรอาหาร
func (h *RecipesHandler) UpdateRecipe(c *gin.Context) {
	// ดึงพารามิเตอร์ URL
	id := c.Param("id")

	// ดึง request body และแปลงเป็นโครงสร้าง Recipe
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// เรียกใช้ store เพื่ออัปเดตสูตรอาหาร
	err := h.store.Update(id, recipe)
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งผลลัพธ์สำเร็จกลับ
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeleteRecipe คือ handler สำหรับลบสูตรอาหาร
func (h *RecipesHandler) DeleteRecipe(c *gin.Context) {
	// ดึงพารามิเตอร์ URL
	id := c.Param("id")

	// เรียกใช้ store เพื่อลบสูตรอาหาร
	err := h.store.Remove(id)
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งผลลัพธ์สำเร็จกลับ
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
