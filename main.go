package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/husobee/vestigo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Page struct {
	Viewer string
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type FoodIngredient struct {
	Foodid       string
	Ingredientid int
	Amount       float32
}

const (
	errorCode  = 555
	serverName = "GWS"
	userName   = "User"
	imageName  = "food.jpg"
	fileName   = "response.json"
	addr       = "postgresql://admin@localhost:26257/broccolibot?sslmode=disable"
)

func main() {
	router := vestigo.NewRouter()

	// set up router global CORS policy
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{"*"},
		AllowCredentials: false,
		MaxAge:           3600 * time.Second,
	})

	fileServerAssets := http.FileServer(http.Dir("assets"))
	router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Server", serverName)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/assets")
		fileServerAssets.ServeHTTP(w, r)
	})

	fileServerImages := http.FileServer(http.Dir("images"))
	router.Get("/images/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Server", serverName)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/images")
		fileServerImages.ServeHTTP(w, r)
	})

	router.Post("/postImage", postImage)
	router.Post("/calculateNutrition/:foodID/:amount", calculateNutrition)
	router.Post("/updateUser", updateUser)
	router.Get("/getUser", getUser)
	router.Get("/", viewIndex)

	log.Println("Listening...")
	if err := http.ListenAndServe(":4444", router); err != nil {
		log.Println(err)
	}
}

/*
  ========================================
  1. View Index
  ========================================
*/

func viewIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("=== view index ===")
	returnCode := 0

	setHeader(w)
	var page Page

	layout := path.Join("assets/html", "index.html")
	content := path.Join("assets/html", "content.html")

	t, err := template.ParseFiles(layout, content)
	if err != nil {
		returnCode = 1
		log.Println("view index 1:", err)
	}

	if returnCode == 0 {
		if err := t.ExecuteTemplate(w, "my-template", page); err != nil {
			returnCode = 2
			log.Println("view index 2:", err)
		}
	}

	// error handling
	if returnCode != 0 {
		handleError(returnCode, errorCode, "Error: view index", w)
	}
}

/*
  ========================================
  2. Post Image
  ========================================
*/

func postImage(w http.ResponseWriter, r *http.Request) {
	log.Println("=== post image ===")
	returnCode := 0

	// fileName := time.Now().Format("20060102150405") + ".jpg"

	file, _, err := r.FormFile("uploadFile")
	defer file.Close()
	if err != nil {
		returnCode = 1
		log.Println("post image 1:", err)
	}

	originalImage, _, err := image.Decode(file) // decode file
	if returnCode == 0 {
		if err != nil {
			returnCode = 2
			log.Println("post image 2:", err)
		}
	}

	// Step 4: Resize image to width = 600 px.
	imageCompressed := imaging.Resize(originalImage, 600, 0, imaging.Linear)

	// Step 5: Save image in directory.
	imageFile, err := os.Create(imageName)
	defer imageFile.Close()
	if returnCode == 0 {
		if err != nil {
			log.Println("post image 5:", err)
		}
	}

	// Step 6: Compress image.
	if returnCode == 0 {
		if err := jpeg.Encode(imageFile, imageCompressed, &jpeg.Options{100}); err != nil {
			log.Println("process image 6:", err)
		}
	}

	classifyImage(imageName)

	if returnCode == 0 {
		if err := json.NewEncoder(w).Encode("post image: true"); err != nil {
			returnCode = 2
			log.Println("post image 2:", err)
		}
	}

	// error handling
	if returnCode != 0 {
		handleError(returnCode, errorCode, "Error: post image", w)
	}
}

/*
  ========================================
  3. Classify Image
  ========================================
*/

func classifyImage(fileName string) {
	log.Println("== classify image ===")

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("./script.sh")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Println("classify image 3:", err)
	}
}

/*
  ========================================
  4. Calculate Nutrition
  ========================================
*/

func calculateNutrition(w http.ResponseWriter, r *http.Request) {
	log.Println("=== calculate nutrition ===")

	foodID := vestigo.Param(r, "foodID")
	amount, err := strconv.Atoi(vestigo.Param(r, "amount"))

	db, err := gorm.Open("postgres", addr)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("total serving amount:", amount)

	foodIngredients := new([]FoodIngredient)
	db.Table("ingredients").Select("sum(calories * amount)").Joins("JOIN food_ingredients on id = ingredientid").Where("foodid = ?", foodID).Find(&foodIngredients)
	// SELECT SUM(calories * amount) FROM ingredients JOIN food_ingredients on id = ingredientid WHERE foodid = $foodid

	log.Println(*foodIngredients)
}

/*
  ========================================
  5. Update User
  ========================================
*/

func updateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("=== update user ===")
	returnCode := 0

	user := new(User)
	user.Name = userName
	nutr := new(Nutrient)

	if err := json.NewDecoder(r.Body).Decode(nutr); err != nil {
		returnCode = 1
		log.Println("update user 1:", err)
	}

	if returnCode == 0 {
		if err := updateDataDB(user, nutr); err != nil {
			returnCode = 2
			log.Println("update user 2:", err)
		}
	}

	if returnCode == 0 {
		if err := json.NewEncoder(w).Encode(user); err != nil {
			returnCode = 3
			log.Println("update user 3:", err)
		}
	}

	// error handling
	if returnCode != 0 {
		handleError(returnCode, errorCode, "Error: update user.", w)
	}
}

/*
  ========================================
  6. Get User
  ========================================
*/

func getUser(w http.ResponseWriter, r *http.Request) {
	log.Println("=== get user ===")
	returnCode := 0

	user := new(User)
	user.Name = userName

	if err := getUserDB(user); err != nil {
		returnCode = 1
		log.Println("get user err 1:", err)
	}

	if returnCode == 0 {
		if err := json.NewEncoder(w).Encode(user); err != nil {
			returnCode = 2
			log.Println("get user err 2:", err)
		}
	}

	// error handling
	if returnCode != 0 {
		handleError(returnCode, errorCode, "User could not be gotten.", w)
	}
}

/*
  ========================================
  Basics
  ========================================
*/

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-control", "no-cache, no-store, max-age=0, must-revalidate")
	w.Header().Set("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
	w.Header().Set("Server", serverName)
}

func handleError(returnCode, statusCode int, message string, w http.ResponseWriter) {
	error := new(Error)
	error.Code = returnCode
	error.Message = message

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(error); err != nil {
		log.Println(err)
	}
}
