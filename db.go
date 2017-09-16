package main

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const addr = "postgresql://admin@localhost:26257/broccolibot?sslmode=disable"

type FoodIngredient struct {
	Foodid       string
	Ingredientid int
	Amount       float32
}

func getNutrition(foodID string) {
	db, err := gorm.Open("postgres", addr)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	foodIngredients := new([]FoodIngredient)
	db.Where("foodid = ?", foodID).Find(&foodIngredients)

	log.Println(*foodIngredients)
}
