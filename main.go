package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fydhfzh/url-shortener/db"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var urlCollection *db.UrlCollection
var baseUrl string

type shortenBody struct {
	LongUrl string `json:"long_url"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	urlCollection = db.InitDB()
	baseUrl = os.Getenv("LOCAL_BASE_URL")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "Go url shortener"})
	})
	r.GET("/:code", redirect)
	r.POST("/shorten", shorten)

	r.Run(":5000")
}

func shorten(c *gin.Context) {
	var body shortenBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, urlError := url.ParseRequestURI(body.LongUrl)

	if urlError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": urlError.Error()})
		return
	}

	urlCode, idErr := shortid.Generate()

	if idErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	result, queryErr := urlCollection.FindOne(urlCode)

	if queryErr != nil {
		if queryErr != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
	}

	if len(result) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Code in use: %s", urlCode)})
		return
	}

	date := time.Now()
	expires := date.AddDate(0, 0, 5)
	newUrl := baseUrl + urlCode
	docId := primitive.NewObjectID()

	newDoc := &db.UrlDoc{
		ID:        docId,
		UrlCode:   urlCode,
		LongUrl:   body.LongUrl,
		ShortUrl:  newUrl,
		CreatedAt: time.Now(),
		ExpiresAt: expires,
	}

	insertErr := urlCollection.InsertOne(newDoc)

	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"newUrl":  newUrl,
		"expires": expires.Format("2006-01-02 15:04:05"),
		"db_id":   docId,
	})
}

func redirect(c *gin.Context) {
	code := c.Param("code")

	var result bson.M
	result, queryErr := urlCollection.FindOne(code)

	if queryErr != nil {
		if queryErr == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No URL with code: %s", code)})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
	}

	log.Print(result["longUrl"])

	longUrl := fmt.Sprint(result["longUrl"])
	c.Redirect(http.StatusPermanentRedirect, longUrl)
}
