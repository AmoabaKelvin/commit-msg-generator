package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type CommitMessage struct {
	CommitMessage string `json:"commit_message" jsonschema_description:"The commit message"`
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var commitMessageSchema = GenerateSchema[CommitMessage]()

func GenerateCommitMessage(diff string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("commit_message"),
		Description: openai.F("The commit message"),
		Schema:      openai.F(commitMessageSchema),
		Strict:      openai.Bool(true),
	}

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(`You are a helpful assistant that generates commit messages for git commits. 
			The commit messages should be in the active voice and no more than 50 characters.
			The commit message should be detailed enough to understand the changes made to the code.
			The commit messages should be in the format of a conventional commit message.
			If there are several additions and are all not related to the same thing, make sure you add
			a separate commit message for each addition in the description section.
			Here is the diff of the files that are staged for commit:
			%s`, diff)),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
	})
	if err != nil {
		panic(err.Error())
	}

	commitMessage := CommitMessage{}
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &commitMessage)
	if err != nil {
		return "", err
	}
	return commitMessage.CommitMessage, nil
}
