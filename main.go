package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	openai "github.com/sashabaranov/go-openai"
)

type Request struct {
	Url    string                     `json:"url"`
	Schema map[string]json.RawMessage `json:"schema"`
}

func main() {
	OPEN_AI_KEY := os.Getenv("OPEN_AI_KEY")
	wsURL := os.Getenv("WS_URL")

	if OPEN_AI_KEY == "" {
		log.Fatal("OPEN_AI_KEY is not set")
	}

	if wsURL == "" {
		log.Fatal("WS_URL is not set")
	}

	log.Println("Starting server...")
	log.Println("WS_URL: ", wsURL)

	openAIclient := openai.NewClient(OPEN_AI_KEY)

	r := gin.Default()

	browser := rod.New().ControlURL(wsURL).MustConnect()

	r.POST("/lookup", func(c *gin.Context) {
		var requestBody Request

		if err := c.BindJSON(&requestBody); err != nil {
			log.Fatal(err)
		}

		log.Println("Request URL: ", requestBody.Url)

		var pageText string
		page := browser.MustPage(requestBody.Url)
		page.MustWaitLoad()
		page.MustWaitRequestIdle()
		page.MustWaitDOMStable()
		page.MustWaitStable()
		pageText = page.MustEval(`() => {
      function textNodesUnder(el){
        var n, a=[], walk=document.createTreeWalker(el,NodeFilter.SHOW_TEXT,null,false);
        while(n=walk.nextNode()) a.push(n);
        return a;
      }

      return textNodesUnder(document).filter(element => element.parentElement.tagName !== 'SCRIPT' && element.parentElement.tagName !== 'STYLE').map(v => v.nodeValue).map(s => s.trim()).join(' ')
    }`).Str()

		prompt := `
  I have text data that was extracted from a webpage. The text data is as follows:

  {{.PageText}}

  I need to extract information from this data. I will provide JSON schema for the data. Return me the data in JSON format.
  `

		t := template.Must(template.New("prompt").Parse(prompt))
		data := struct {
			PageText string
		}{
			PageText: pageText,
		}

		buf := &bytes.Buffer{}
		err := t.Execute(buf, data)
		if err != nil {
			log.Fatal(err)
		}

		req := openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo16K0613,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: buf.String(),
				},
			},
			Functions: []openai.FunctionDefinition{
				{
					Name:        "ParseDataToJSON",
					Description: "Parses text data from the webpage to JSON format",
					Parameters:  requestBody.Schema,
				},
			},
		}

		resp, err := openAIclient.CreateChatCompletion(c, req)
		if err != nil {
			log.Printf("Completion error: %v\n", err)
			return
		}

		if err != nil {
			log.Fatal(err)
		}

		log.Println(resp.Choices[0].Message.FunctionCall.Arguments)

		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &jsonMap)

		c.JSON(http.StatusOK, gin.H{
			"result": jsonMap,
		})
	})

	r.Run(":8080")
}
