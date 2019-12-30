package main

import (
	// "fmt"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/expectedsh/go-sonic/sonic"
	"github.com/gin-gonic/gin"
)

// CachedPageHandler returns a browser rendered React app
func CachedPageHandler(c *gin.Context) {
	/*
		In this handler we will load in a html file as a string
		and return the contents to the browser

		This allows use to edit the index file and serve it
		with out having to make changes to the server
	*/
	content, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal(err)
	}
	yourHTMLString := string(content)
	//Write your 200 header status (or other status codes, but only WriteHeader once)
	c.Writer.WriteHeader(http.StatusOK)
	//Convert your cached html string to byte array
	c.Writer.Write([]byte(yourHTMLString))
	return
}

type incommingInput struct {
	Input string `json:"input"`
}

type indexedData struct {
	Textbook []map[string]string `json:"textbook"`
	Syllabus []map[string]string `json:"syllabus"`
	Faq      []map[string]string `json:"faq"`
}

type pseudoDatabase struct {
	Textbook []outputResult `json:"textbook"`
	Syllabus []outputResult `json:"syllabus"`
	Faq      []outputResult `json:"faq"`
}

type outputResult struct {
	Key         string `json:"key"`
	Location    string `json:"location"`
	Value       string `json:"value"`
	IndexedText string `json:"indexed_text"`
}

func main() {
	r := gin.Default()
	r.GET("/", CachedPageHandler)

	r.GET("/flush", func(c *gin.Context) {
		ingester, err := sonic.NewIngester("localhost", 1491, "SecretPassword")
		if err != nil {
			panic(err)
		}

		ingester.FlushBucket("db", "textbook")
		ingester.FlushBucket("db", "syllabus")
		ingester.FlushBucket("db", "faq")

	})

	r.GET("/ingester", func(c *gin.Context) {
		ingester, err := sonic.NewIngester("localhost", 1491, "SecretPassword")
		if err != nil {
			panic(err)
		}

		// read formatted file to struct
		file, _ := ioutil.ReadFile("to_ingest.json")
		data := indexedData{}
		_ = json.Unmarshal([]byte(file), &data)

		// add each item to sonic
		for _, toAdd := range data.Textbook {
			key := toAdd["key"]
			val := toAdd[key]
			// fmt.Println(key, val)
			ingester.Push("db", "textbook", key, val)

		}

		// add each item to sonic
		for _, toAdd := range data.Syllabus {
			key := toAdd["key"]
			val := toAdd[key]
			// fmt.Println(key, val)
			ingester.Push("db", "syllabus", key, val)
			// fmt.Println(resp)
		}

		// add each item to sonic
		for _, toAdd := range data.Faq {
			key := toAdd["key"]
			val := toAdd[key]
			// fmt.Println(key, val)
			ingester.Push("db", "faq", key, val)
			// fmt.Println(resp)
		}

		c.JSON(200, gin.H{
			"message": "ingested",
		})
	})

	r.POST("/search", func(c *gin.Context) {
		var loginCmd incommingInput
		c.BindJSON(&loginCmd)

		search, err := sonic.NewSearch("localhost", 1491, "SecretPassword")
		if err != nil {
			panic(err)
		}
		textbookResults, err := search.Query("db", "textbook", loginCmd.Input, 10, 0)
		syllabusResults, err := search.Query("db", "syllabus", loginCmd.Input, 10, 0)
		faqResults, err := search.Query("db", "faq", loginCmd.Input, 10, 0)
		if err != nil {
			panic(err)
		}

		// fetch results based on ids
		// read formatted file to struct
		file, _ := ioutil.ReadFile("to_ingest.json")
		data := pseudoDatabase{}
		_ = json.Unmarshal([]byte(file), &data)
		// fmt.Println(data.Textbook)
		// fmt.Println(faqResults)
		// fmt.Println(textbookResults)
		// fmt.Println(syllabusResults)

		// var faqResultsFinal = make([]outputResult)
		var faqResultsFinal []outputResult
		for _, faqRes := range faqResults {
			for _, faq := range data.Faq {
				if faq.Key == faqRes {

					faqResultsFinal = append(faqResultsFinal, faq)
				}
			}
		}

		// spew.Dump(faqResultsFinal)
		// var textbookResultsFinal = []outputResult{}
		var textbookResultsFinal []outputResult
		for _, textRes := range textbookResults {
			for _, txt := range data.Textbook {
				if txt.Key == textRes {
					// spew.Dump(txt)
					textbookResultsFinal = append(textbookResultsFinal, txt)
					// textbookResultsFinal = append(newVal, textbookResultsFinal)
				}
			}
		}

		// var syllabusResultsFinal = []outputResult{}
		var syllabusResultsFinal []outputResult
		for _, sylRes := range syllabusResults {
			for _, syl := range data.Syllabus {
				if syl.Key == sylRes {
					// spew.Dump(syl)
					// syllabusResultsFinal = append(newVal, syllabusResultsFinal)
					syllabusResultsFinal = append(syllabusResultsFinal, syl)
				}
			}
		}

		c.JSON(200, gin.H{
			"textbook": textbookResultsFinal,
			"syllabus": syllabusResultsFinal,
			"faq":      faqResultsFinal,
		})
	})
	r.POST("/suggest", func(c *gin.Context) {

		var loginCmd incommingInput
		c.BindJSON(&loginCmd)

		search, err := sonic.NewSearch("localhost", 1491, "SecretPassword")
		if err != nil {
			panic(err)
		}

		results, err := search.Suggest("db", "textbook", loginCmd.Input, 10)

		if err != nil {
			panic(err)
		}

		// fmt.Println(results)

		c.JSON(200, gin.H{
			"message": results,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
