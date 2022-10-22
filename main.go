package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const uri = "mongodb+srv://CalcData:r5p3Gwuhn7ELIm3z@cluster0.vif5nkw.mongodb.net/?retryWrites=true&w=majority"

type User struct {
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	Verefy     string   `json:"verefy"`
	Times      []string `json:"times"`
	Companets  []string `json:"companets"`
	Token      string   `json:"token"`
}

type Token struct {
	Token string `json:"token"`
}

func main() {
	r := gin.Default()
	r.POST("register", register)
	r.GET("login", login)
	r.POST("cheskverefy", cheskverefy)
	r.POST("verefyuser", verefyUser)
	r.GET("getuser", getUser)
	r.GET("getusers", getAllUsers)
	r.POST("addtime", addTime)
	r.POST("updatetime", updateTime)
	r.Run(":8080")
}
func passwordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Println(err)
	}
	return string(hash)
}

func register(c *gin.Context) {
	//chesk email data base if exist return error if not create new user and return token to client
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: user.Email}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)

	if result.Email == user.Email {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exist"})
		return
	}
	user = User{
		Email:      user.Email,
		Password:   passwordHash(user.Password),
		Verefy:     "false",
		Times:      []string{},
		Companets:  []string{},
		Token:      createToken(user.Email),
	}
	user.Token = createToken(user.Email)
	collection.InsertOne(context.Background(), user)
	c.JSON(http.StatusOK, user)
}

func login(c *gin.Context) {
	//if verfy is false return error if true return token
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	//all users in data base
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result User
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		if result.Email == user.Email {
			if result.Verefy == "false" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "email is not verifed"})
				return
			}
			err := bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "password is incorrect"})
				return
			}
			c.JSON(http.StatusOK, Token{Token: createToken(user.Email)})
			return
		}
	}
}

func cheskverefy(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: user.Email}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	if result.Email == user.Email {
		c.JSON(http.StatusOK, gin.H{"verefy": result.Verefy})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "email is incorrect"})
}

func verefyUser(c *gin.Context) {
	//post authorization bearer token user db update verefy to true and return token to client 

	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: user.Email}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	if result.Email == user.Email {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "verefy", Value: "true"},
			}},
		}
		collection.UpdateOne(context.Background(), filter, update)
		c.JSON(http.StatusOK, Token{Token: createToken(user.Email)})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "email is incorrect"})
}

func getUser(c *gin.Context) {
	//get authorization bearer token return all user data
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is empty"})
		return
	}
	
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "token", Value: token}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	if result.Token == token {
		c.JSON(http.StatusOK, result)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "token is incorrect"})

}

func getAllUsers(c *gin.Context) {
	//get authorization bearer token user db return all users all data
	token := c.Request.Header.Get("Authorization")
	token = token[7:len(token)]
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: claims["email"]}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	if result.Email == claims["email"] {
		var results []*User
		cur, err := collection.Find(context.Background(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}
		for cur.Next(context.Background()) {
			var elem User
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, &elem)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		cur.Close(context.Background())
		//results = results[1:]

		c.JSON(http.StatusOK, results)
		return
	}
}

func addTime(c *gin.Context) {
	//get authorization bearer token add array time db user
	token := c.Request.Header.Get("Authorization")
	token = token[7:len(token)]
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: claims["email"]}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	if result.Email == claims["email"] {
		update := bson.D{
			{Key: "$push", Value: bson.D{
				{Key: "times", Value: user.Times},
			}},
		}
		collection.UpdateOne(context.Background(), filter, update)
		c.JSON(http.StatusOK, gin.H{"message": "time added"})
		return
	}
}

func updateTime(c *gin.Context) {
	//{"_id":{"$oid":"6353b71356649875aca01589"},"email":"Gani@gmail.com","password":"$2a$10$UCjP8dAiyCdR7DPo9V0I2.it2IikdrdxP73WVoZ7w2pCJpjb/Oi7W","verefy":"true","times":[],"companets":[],"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJlbWFpbCI6IkdhbmlAZ21haWwuY29tIn0.I8RA_hRIY_kfj65ZtNGtmpYsarLwFQIDC5xBZbxFnGY"} update index time db
	token := c.Request.Header.Get("Authorization")
	token = token[7:len(token)]
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	var user User
	c.BindJSON(&user)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	collection := client.Database("CalcData").Collection("users")
	filter := bson.D{{Key: "email", Value: claims["email"]}}
	var result User
	collection.FindOne(context.Background(), filter).Decode(&result)
	insex := 0
	if result.Email == claims["email"] {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "insex", Value: insex},
				{Key: "times." + strconv.Itoa(insex), Value: user.Times},
			}},
		}
		collection.UpdateOne(context.Background(), filter, update)
		c.JSON(http.StatusOK, gin.H{"message": "time updated"})
		return
	}
}

func createToken(username string) string {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["email"] = username
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("SECRET")))
	return tokenString
}
