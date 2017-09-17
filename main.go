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

	calculate("pancakes", 2)
	// runScraper("pancakes")

	router.Post("/postImage", postImage)
	router.Post("/calculateNutrition/:foodID/:amount", calculateNutrition)
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
  4. Run Scraper
  ========================================
*/

func runScraper(arguments string) {
	cmd := "/scrapper/BroccoliScrapper.exe"
	args := []string{arguments}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		log.Println(os.Stderr, err)
		os.Exit(1)
	}
	log.Println("Scrapper done")
}

/*
  ========================================
  5. Calculate Nutrition
  ========================================
*/

func calculate(foodID string, amount float32) {
	log.Println("=== calculate nutrition ===")

	db, err := gorm.Open("postgres", addr)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("=== total amount", amount, "===")

	foodIngredients := new([]FoodIngredient)
	nutrient := new(Nutrient)

	db.Table(
		"broccolibot.ingredients").Select(
		`sum(calories * amount) AS "calories",
		 sum(fat * amount) AS "fat",
		 sum(cholesterol * amount) AS "cholesterol",
		 sum(sodium * amount) AS "sodium",
		 sum(carbohydrates * amount) AS "carbohydrates",
		 sum(protein * amount) AS "protein"`).Joins(
		"JOIN broccolibot.food_ingredients on id = ingredientid").Where(
		"foodid = ?", foodID).Find(&nutrient)
	// SELECT SUM(calories * amount) AS "Calories" FROM ingredients JOIN food_ingredients on id = ingredientid WHERE foodid = $foodid

	log.Println(*foodIngredients)
	log.Println(*nutrient)
}

func calculateNutrition(w http.ResponseWriter, r *http.Request) {
	log.Println("=== calculate nutrition ===")
	returnCode := 0

	foodID := vestigo.Param(r, "foodID")
	// amount, err := strconv.Atoi(vestigo.Param(r, "amount"))

	runScraper(foodID)

	/*
		db, err := gorm.Open("postgres", addr)
		defer db.Close()

		if err != nil {
			log.Fatal(err)
		}

		log.Println("=== total amount", amount, "===")
		// **** Do stuff with total amount ****

		nutrient := new(Nutrient)

		db.Table(
			"broccolibot.ingredients").Select(
			`sum(calories * amount) AS "calories",
			 sum(fat * amount) AS "fat",
			 sum(cholesterol * amount) AS "cholesterol",
			 sum(sodium * amount) AS "sodium",
			 sum(carbohydrates * amount) AS "carbohydrates",
			 sum(protein * amount) AS "protein"`).Joins(
			"JOIN broccolibot.food_ingredients on id = ingredientid").Where(
			"foodid = ?", foodID).Find(&nutrient)
		// SELECT SUM(calories * amount) AS "Calories" FROM ingredients JOIN food_ingredients on id = ingredientid WHERE foodid = $foodid

		log.Println(*nutrient)
	*/

	if returnCode == 0 {
		if err := json.NewEncoder(w).Encode("nutrient"); err != nil {
			returnCode = 3
			log.Println("calculate nutrition 3:", err)
		}
	}

	// error handling
	if returnCode != 0 {
		handleError(returnCode, errorCode, "Error: calculate nutrition.", w)
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
