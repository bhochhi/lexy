package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
)

type Intent struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	SampleUtterances []string `json:"sampleUtterances"`
	Slots            []Slot   `json:"slots,omitempty"`
}

type Slot struct {
	Name                   string                 `json:"name"`
	SlotType               string                 `json:"slotType"`
	SlotConstraint         string                 `json:"slotConstraint"`
	ValueElicitationPrompt ValueElicitationPrompt `json:"valueElicitationPrompt"`
}

type ValueElicitationPrompt struct {
	Messages    []Message `json:"messages"`
	MaxAttempts int       `json:"maxAttempts"`
}

type Message struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

func uploadIntent(svc *lexmodelbuildingservice.LexModelBuildingService, intent Intent) {
	_, err := svc.PutIntent(&lexmodelbuildingservice.PutIntentInput{
		Name:             aws.String(intent.Name),
		Description:      aws.String(intent.Description),
		SampleUtterances: aws.StringSlice(intent.SampleUtterances),
		Slots:            convertSlots(intent.Slots),
		FulfillmentActivity: &lexmodelbuildingservice.FulfillmentActivity{
			Type: aws.String("ReturnIntent"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to upload intent %s: %v", intent.Name, err)
	}
	fmt.Printf("Uploaded intent: %s\n", intent.Name)
}

func convertSlots(slots []Slot) []*lexmodelbuildingservice.Slot {
	var lexSlots []*lexmodelbuildingservice.Slot
	for _, slot := range slots {
		lexSlot := &lexmodelbuildingservice.Slot{
			Name:           aws.String(slot.Name),
			SlotType:       aws.String(slot.SlotType),
			SlotConstraint: aws.String(slot.SlotConstraint),
			ValueElicitationPrompt: &lexmodelbuildingservice.Prompt{
				MaxAttempts: aws.Int64(int64(slot.ValueElicitationPrompt.MaxAttempts)),
				Messages:    convertMessages(slot.ValueElicitationPrompt.Messages),
			},
		}
		lexSlots = append(lexSlots, lexSlot)
	}
	return lexSlots
}

func convertMessages(messages []Message) []*lexmodelbuildingservice.Message {
	var lexMessages []*lexmodelbuildingservice.Message
	for _, message := range messages {
		lexMessage := &lexmodelbuildingservice.Message{
			ContentType: aws.String(message.ContentType),
			Content:     aws.String(message.Content),
		}
		lexMessages = append(lexMessages, lexMessage)
	}
	return lexMessages
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := lexmodelbuildingservice.New(sess)

	file, err := ioutil.ReadFile("../data/intents.json")
	if err != nil {
		log.Fatalf("Failed to read intents.json: %v", err)
	}

	var intents []Intent
	if err := json.Unmarshal(file, &intents); err != nil {
		log.Fatalf("Failed to unmarshal intents.json: %v", err)
	}

	for _, intent := range intents {
		uploadIntent(svc, intent)
	}
}
