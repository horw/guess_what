package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func SetupENV() {
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AddConfigPath("./env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
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

func run(content string) {
	SetupENV()
	client := GetClient()

	resp, err := gptRequest(client, content)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Message.Content)
}

func main() {

	var content string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "content",
				Value:       "english",
				Usage:       "content for ChatGPT",
				Destination: &content,
			},
		},
		Action: func(cCtx *cli.Context) error {
			run(content)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
