package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func main() {
	lambda.Start(bookitCheck)
}

func bookitCheck() {

	request := os.Getenv("REQUEST_URL")
	recipients := strings.Split(os.Getenv("RECIPIENTS_CSV"), ",")

	if len(recipients) == 0 {
		panic("No recipients have been found")
	}

	if len(request) == 0 {
		panic("No request Url has been set")
	}
	
	// make http request
	resp, err := http.Get(request)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	response := string(body)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "callback=jQuery211003281978113731543_157021192534(")
	response = strings.TrimSuffix(response, ");")

	var slots Slots

	json.Unmarshal([]byte(response), &slots)
	log.Println(slots)
	firstSlot, err := getFirstAvailableSlot(slots)
	if err == nil {
		log.Println("found a slot. sending sms")

		for _, recipient := range recipients {
			// send sns event
			sendSNSEvent(firstSlot, recipient)
		}
	} else {
		log.Println("No slot found")
	}
}

func sendSNSEvent(slot Slot, recipient string) {
	slotStr, err := json.Marshal(slot)
	if err != nil {
		log.Fatalln(err)
	}

	client := sns.New(session.New())
	input := &sns.PublishInput{
		Message:     aws.String("New Slot available " + string(slotStr) + " click url asap - https://app.bookitit.com/en/hosteds/widgetdefault/21a8d76163e6f2dc0e5ca528c922d37c3"),
		PhoneNumber: aws.String(recipient),
	}

	result, err := client.Publish(input)
	if err != nil {
		log.Fatalln("Publish error:", err)
		return
	}
	log.Println(result)
}

func getFirstAvailableSlot(slots Slots) (Slot, error) {
	var ret Slot
	for _, slot := range slots.Slots {
		if len(slot.Times) > 0 {
			return slot, nil
		}
	}
	return ret, errors.New("No available slot")
}

type Slots struct {
	Slots []Slot `json:"Slots"`
}

type Slot struct {
	Agenda string                 `json:"agenda"`
	Date   string                 `json:"date"`
	Times  map[string]interface{} `json:"times"`
	State  int                    `json:"state"`
}
