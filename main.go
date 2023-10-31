package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("dev")   // name of config file (without extension)
	viper.SetConfigType("env")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./env") // path to look for the config file in
	err := viper.ReadInConfig()  // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	token := viper.Get("OPENAI_TOKEN")
	config := openai.DefaultConfig(token)
	proxyUrl, err := url.Parse("http://127.0.0.1:1095")
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
	}

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: os.Args[1],
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
