package main

import (
    "fmt"
    "github.com/slack-go/slack/socketmode"
    "log"
    "os"
    "strings"

	"github.com/joho/godotenv"
    "github.com/slack-go/slack"
    "github.com/slack-go/slack/slackevents"
)

func envLoad() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func main() {
	envLoad()

	webAPI := slack.New(
        os.Getenv("SLACK_BOT_TOKEN"),
        slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
        slack.OptionDebug(true),
        slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
    )
    socketMode := socketmode.New(
        webAPI,
        socketmode.OptionDebug(true),
        socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
    )
    authTest, authTestErr := webAPI.AuthTest()
    if authTestErr != nil {
        fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
        os.Exit(1)
    }
    selfUserID := authTest.UserID

    go func() {
        for envelope := range socketMode.Events {
            switch envelope.Type {
            case socketmode.EventTypeEventsAPI:
                socketMode.Ack(*envelope.Request)

                eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)
                switch eventPayload.Type {
                case slackevents.CallbackEvent:
                    switch event := eventPayload.InnerEvent.Data.(type) {
                    case *slackevents.MessageEvent:
                        if event.User != selfUserID && strings.Contains(event.Text, "こんにちは") {
                            _, _, err := webAPI.PostMessage(
                                event.Channel,
                                slack.MsgOptionText(
                                    fmt.Sprintf(":wave: こんにちは <@%v> さん！", event.User),
                                    false,
                                ),
                            )
                            if err != nil {
                                log.Printf("Failed to reply: %v", err)
                            }
                        }
                    default:
                        socketMode.Debugf("Skipped: %v", event)
                    }
                default:
                    socketMode.Debugf("unsupported Events API eventPayload received")
                }
            default:
                socketMode.Debugf("Skipped: %v", envelope.Type)
            }
        }
    }()

    socketMode.Run()
}
