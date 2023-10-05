package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

type Request struct {
	Url    string                     `json:"url"`
	Schema map[string]json.RawMessage `json:"schema"`
}

func main() {
	OPEN_AI_KEY := os.Getenv("OPEN_AI_KEY")
	openAIclient := openai.NewClient(OPEN_AI_KEY)

	allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:3000")
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	r := gin.Default()

	r.POST("/test", func(c *gin.Context) {
		var requestBody Request

		if err := c.BindJSON(&requestBody); err != nil {
			log.Fatal(err)
		}

		var pageText string
		err := chromedp.Run(chromeCtx,
			chromedp.Navigate(requestBody.Url),
			chromedp.Evaluate(`
    function textNodesUnder(el){
      var n, a=[], walk=document.createTreeWalker(el,NodeFilter.SHOW_TEXT,null,false);
      while(n=walk.nextNode()) a.push(n);
      return a;
    }
    textNodesUnder(document).filter(element => element.parentElement.tagName !== 'SCRIPT' && element.parentElement.tagName !== 'STYLE').map(v => v.nodeValue).map(s => s.trim()).join(' ')
    `, &pageText),
		)

		if err != nil {
			log.Fatal(err)
		}

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
		err = t.Execute(buf, data)
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

		resp, err := openAIclient.CreateChatCompletion(chromeCtx, req)
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

	r.Run(":8081")
}
