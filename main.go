package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

//go:embed templates/* env/*
var f embed.FS

func SetupENV() {

	data, err := f.ReadFile("env/dev")
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	viper.SetConfigType("env")
	viper.ReadConfig(strings.NewReader(string(data)))
}

func GetClient() *openai.Client {

	token := viper.Get("OPENAI_TOKEN").(string)
	config := openai.DefaultConfig(token)

	proxy := viper.Get("PROXY")
	if proxy != nil {
		proxyUrl, err := url.Parse(proxy.(string))
		if err != nil {
			panic(err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		config.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	client := openai.NewClientWithConfig(config)
	return client

}

func gptRequest(client *openai.Client, content string) (response openai.ChatCompletionResponse, err error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	return resp, err
}

func ReadTemplate(template string) string {

	dat, err := f.ReadFile(fmt.Sprintf("templates/%s", template))
	check(err)
	return string(dat)
}

func run(content string, template string) {
	SetupENV()
	temp := ReadTemplate(template)
	formated_content := fmt.Sprintf(temp, content)

	client := GetClient()

	resp, err := gptRequest(client, formated_content)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Message.Content)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	var content string
	var clip string
	var template string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "content",
				Value:       "english",
				Usage:       "content for ChatGPT",
				Destination: &content,
			},
			&cli.StringFlag{
				Name:        "template",
				Value:       "check_grammar",
				Usage:       "template for ChatGPT",
				Destination: &template,
			},
			&cli.StringFlag{
				Name:        "c",
				Value:       "n",
				Usage:       "use clipboard instead input? y/n. (defaut n)",
				Destination: &clip,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if clip == "y" {
				err := clipboard.Init()
				if err != nil {
					panic(err)
				}
				content = string(clipboard.Read(clipboard.FmtText))
			}
			fmt.Printf("Your msg: %s\n", content)
			run(content, template)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
