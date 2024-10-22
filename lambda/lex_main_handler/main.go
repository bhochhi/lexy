// lambda/lex_main_handler/main.go
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LexDialogActionMessage struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type LexDialogAction struct {
	Type             string                  `json:"type"`
	FulfillmentState string                  `json:"fulfillmentState"`
	Message          *LexDialogActionMessage `json:"message"`
}

type LexResponse struct {
	SessionAttributes map[string]string `json:"sessionAttributes"`
	DialogAction      LexDialogAction   `json:"dialogAction"`
}

func handleRequest(ctx context.Context, request events.LexEvent) (LexResponse, error) {

	fmt.Printf("Hello Request: %+v\n", request)
	// Default response for unhandled intents
	return LexResponse{
    	SessionAttributes: request.SessionAttributes,
    	DialogAction: LexDialogAction{
    		Type:             "Close",
    		FulfillmentState: "Fulfilled",
    		Message: &LexDialogActionMessage{
    			ContentType: "PlainText",
    			Content:     "Thank you for using our service.",
    		},
    	},
    }, nil
}

func main() {
	lambda.Start(handleRequest)
}
