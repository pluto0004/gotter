package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"golang.org/x/exp/utf8string"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var keyword = map[string]string{
    "なんて日だ":     "なんてGoな日だ！！！！",
    "oh...":  "ﾌｧｧｧｧｧｧｧｧｧ",
    "努力":     "GO",
	"はい":     "ふぁーい",
	"します":	"しますだふぁー",
    "まだ":     "まだだふぁー",
    "した":     "しただふぁー",
    "です":     "ですだふぁー",
    "よう":     "ようだふぁー",
    "ゴーファー":      "GOOOOOOOOOOO!",
    "な":      "ふぁ",
}


func main(){
	err := godotenv.Load(fmt.Sprintf("./%s.env", os.Getenv("GO_ENV")))
	if err != nil {
        fmt.Println("Env file cannot be read!")
    }

	dbInit()
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}


	e.Renderer = t
	e.Static("/", "src/images")
	e.Static("/css", "src/css")

	// router
    e.GET("/hello", Hello)
    e.GET("/logs", ShowLog)
	e.POST("/gotweets", GoTweet)
	e.POST("/tweets", tweets)
	e.Logger.Fatal(e.Start(":8000"))

}

// DB related
type Log struct{
	gorm.Model
	Tweet	string
	User	string
}

// DB migration
func dbInit() {
	db, err := gorm.Open(GetDBConfig())
    if err != nil {
        panic("You can't open DB (dbInit())")
    }
	db.AutoMigrate(&Log{})
	defer db.Close()
}

func GetDBConfig()(string, string){
	DBMS     := "postgres"
	PORT 	 := "5432"
	USER     := os.Getenv("DB_USER")
	PASS     := os.Getenv("DB_PASS")
	DBNAME   := os.Getenv("DB")
	HOST 	 := os.Getenv("ENV")
	
	CONNECT := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", HOST, PORT, USER, DBNAME, PASS) //Build connection string

	return DBMS, CONNECT
}

// Add data to DB
func dbInsert(tweet string, user string) {
	db, err := gorm.Open(GetDBConfig())
    if err != nil {
        panic("You can't open DB (dbInsert())")
    }
	db.Create(&Log{Tweet: tweet, User: user})
	fmt.Println("inserted!!")
    defer db.Close()
}

// DB All Get
func dbGetAll() []Log {
	db, err := gorm.Open(GetDBConfig())
    if err != nil {
        panic("You can't open DB (dbGetAll())")
    }
    defer db.Close()
    var logs []Log
	db.Order("created_at desc").Find(&logs)
    return logs
}



func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func Hello(c echo.Context) error {
    return c.Render(http.StatusOK, "hello", Keyword{})
}

func ShowLog(c echo.Context) error {
	logs := dbGetAll()
    return c.Render(http.StatusOK, "show", logs)
}


func connectTwitterApi() *anaconda.TwitterApi{
	// read Json
	raw, error := ioutil.ReadFile("path/to/twitterAccount.json")
	if error != nil {
		fmt.Println((error.Error()))
		return nil
	}
	var twitterAccount TwitterAccount

	json.Unmarshal(raw, &twitterAccount)

	// auth
	return anaconda.NewTwitterApiWithCredentials(twitterAccount.AccessToken, twitterAccount.AccessTokenSecret, twitterAccount.ConsumerKey, twitterAccount.ConsumerSecret)
}

func GoTweet(c echo.Context) error {
	api := connectTwitterApi()
  
	message := c.FormValue("message")

	//Convert
	utf8Message := utf8string.NewString(message)
    size := utf8Message.RuneCount()
    if size > 120 {
        return c.JSON(http.StatusBadRequest, "string length over")
    }
    for key, value := range keyword {
        message = strings.ReplaceAll(message, key, value)
    }
 
    if strings.HasSuffix(message, "ご") {
        message += "ﾌｧｰ"
    }
 
    message = "GO!GO!GO! " + message
    message += "\n#ごーふぁったー"
 
    if utf8string.NewString(message).RuneCount() >= 140 {
        return c.JSON(http.StatusBadRequest, "string length over")
	}
	
	keyword := Keyword{
		Result: message,
	}
	tweet, err := api.PostTweet(message, nil)
	if(err != nil){
	  log.Fatal(err)
	}
	dbInsert(message, tweet.User.Name)
	fmt.Println(tweet.Text,tweet.User.Name)
	return c.Render(http.StatusOK, "hello", keyword)
}


func tweets(c echo.Context) error {
	keyword := Keyword{
		Text: c.FormValue("text"),
	}
	value := keyword.Text
	api := connectTwitterApi()

	searchResult, _ := api.GetSearch(`"` + value + `"`, nil)
	tweets := make([]*TweetTemplate, 0)


	for _, data := range searchResult.Statuses{
		tweet := new(TweetTemplate)
        tweet.Text = data.FullText
        tweet.User = data.User.Name
        tweet.Id = data.User.IdStr
        tweet.ScreenName = data.User.ScreenName
        tweet.Date = data.CreatedAt
		tweet.TweetId = data.IdStr

		tweets = append(tweets, tweet)
	}

    return c.Render(http.StatusOK, "tweets.html", tweets)

}

type TwitterAccount struct{
	AccessToken string `json:"accessToken"`
	AccessTokenSecret string `json:"accessTokenSecret"`
	ConsumerKey string `json:"consumerKey"`
	ConsumerSecret string `json:"consumerSecret"`
}

type Tweet struct{
	User string `json:"user"`
	Text string `json:"text"`
}

type Tweets *[]Tweet

type TweetTemplate struct{
	User string `json:"user"`
	Text string `json:"text"`
	ScreenName string `json:""screenName`
	Id string `json:"id"`
	Date string `json:"date"`
	TweetId string `json:"tweetId`
}

type Template struct {
    templates *template.Template
}

type Keyword struct {
	Text string
	Message string
	Result string
}


