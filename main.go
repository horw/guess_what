package main

import (
	"bufio"
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

var content string
var clip string
var template string
var mode string

func SetupENV() {

	data, err := f.ReadFile("env/dev")
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	viper.SetConfigType("env")
	viper.ReadConfig(strings.NewReader(string(data)))
}

func GetClient(mode string) GuessWhatClient {

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

	switch mode {
	case singleMode:
		return &SingleClient{client: client}
	case dialogMode:
		return &DialogClient{client: client}
	default:
		return &TemplateClient{
			client: client,
		}
	}
}

const (
	singleMode   = "single"
	templateMode = "template"
	dialogMode   = "dialog"
)

type GuessWhatClient interface {
	Work()
}
type SingleClient struct {
	client *openai.Client
}

func (s *SingleClient) Work() {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	resp, _ := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)
	content := resp.Choices[0].Message.Content
	fmt.Println(content)
}

type DialogClient struct {
	client *openai.Client
}

func (s *DialogClient) Work() {
	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := s.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		fmt.Println(content)
	}
}

type TemplateClient struct {
	client *openai.Client
}

func (s *TemplateClient) Work() {
	temp := ReadTemplate(template)
	formated_content := fmt.Sprintf(temp, content)
	resp, _ := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: formated_content,
				},
			},
		},
	)
	content := resp.Choices[0].Message.Content
	fmt.Println(content)
}

func ReadTemplate(template string) string {

	dat, err := f.ReadFile(fmt.Sprintf("templates/%s", template))
	check(err)
	return string(dat)
}

func run() {
	SetupENV()
	client := GetClient(mode)
	client.Work()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "content",
				Value:       "english",
				Usage:       "content for ChatGPT",
				Destination: &content,
			},
			&cli.StringFlag{
				Name:        "mode",
				Value:       "template",
				Usage:       "single/template/dialog mode for chatting",
				Destination: &mode,
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
			run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
