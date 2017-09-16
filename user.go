package main

import (
	// "log"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Name string   `json:"name"`
	Data Nutrient `json:"data"`
	Goal Nutrient `json:"goal"`
}

type Nutrient struct {
	Calories      int `json:"calories"`
	Fat           int `json:"fat"`
	Cholesterol   int `json:"cholesterol"`
	Sodium        int `json:"sodium"`
	Carbohydrates int `json:"carbohydrates"`
	Protein       int `json:"protein"`
}

/*
  ========================================
  Insert User
  ========================================
*/

func insertUserDB() error {
	// create new MongoDB session
	collection, session := initMongoDB("user")
	defer session.Close()

	user := new(User)
	user.Name = userName
	user.Data = *new(Nutrient)
	user.Goal = *new(Nutrient)

	return collection.Insert(user)
}

/*
  ========================================
  Reset Data
  ========================================
*/

func resetDataDB() error {
	// create new MongoDB session
	collection, session := initMongoDB("user")
	defer session.Close()

	selector := bson.M{"name": userName}
	change := bson.M{"data": new(Nutrient)}
	update := bson.M{"$set": &change}

	return collection.Update(selector, update)
}

/*
  ========================================
  Set Goal
  ========================================
*/

func setGoalDB() error {
	// create new MongoDB session
	collection, session := initMongoDB("user")
	defer session.Close()

	nutr := new(Nutrient)
	nutr.Calories = 2000
	nutr.Fat = 65
	nutr.Cholesterol = 300
	nutr.Sodium = 2400
	nutr.Carbohydrates = 300
	nutr.Protein = 50

	selector := bson.M{"name": userName}
	change := bson.M{"goal": nutr}
	update := bson.M{"$set": &change}

	return collection.Update(selector, update)
}

/*
  ========================================
  Get User
  ========================================
*/

func getUserDB(user *User) error {
	// create new MongoDB session
	collection, session := initMongoDB("user")
	defer session.Close()

	selector := bson.M{"name": user.Name}

	return collection.Find(selector).One(user)
}

/*
  ========================================
  Update Data
  ========================================
*/

func updateDataDB(user *User, nutr *Nutrient) error {
	// create new MongoDB session
	collection, session := initMongoDB("user")
	defer session.Close()

	selector := bson.M{"name": user.Name}
	change := bson.M{
		"data.calories": nutr.Calories,
		"data.fat": nutr.Fat,
		"data.cholesterol": nutr.Cholesterol,
		"data.sodium": nutr.Sodium,
		"data.carbohydrates": nutr.Carbohydrates,
		"data.protein": nutr.Protein,
	}
	update := bson.M{"$inc": &change}

	return collection.Update(selector, update)
}
