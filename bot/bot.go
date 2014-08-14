package bot

import (
    //"code.google.com/p/gcfg" // TODO
    "github.com/daneharrigan/hipchat"
    "fmt"
    "strings"
    "errors"
    "log"
)

const (
	conf = "conf.hipchat.com"
    resource = "bot"
)

// TODO enum for status

type Bot struct {
	FullName        string
    MentionName     string

    client          *hipchat.Client
}

type Message struct {
    From    string
    Body    string
}

type Room struct {
    roomJid string
}

func NewBot(user, pass string) (*Bot, error) {

	c, err := hipchat.NewClient(user, pass, resource)
    if err != nil {
        return nil, err
    }

    c.Status("chat")

    log.Println("Connected and available")

	b := &Bot {
		client: c,
	}

    return b, b.init()
}

func (b *Bot) Join(roomId string) (*Room, <-chan *Message) { // TODO add callback to stop blocking, return err?
	log.Println("Joining room: " + roomId)

	roomJid := fmt.Sprintf("%s@%s", roomId, conf)

    b.client.Join(roomJid, b.FullName)

    room := &Room {
        roomJid: roomJid,
    }
    receivedChan := make(chan *Message)

    go func() {
        for message := range b.client.Messages() {
            if strings.HasPrefix(message.Body, "@"+b.MentionName) {
                m := &Message {
                    From: message.From,
                    Body: message.Body[len(b.MentionName):],
                }
                receivedChan <- m
            }
        }
    }()

    return room, receivedChan
}

func (b *Bot) Say(room *Room, msg string) {
    b.client.Say(room.roomJid, b.FullName, msg)
}

func (b *Bot) init() (error) {
	for _, user := range b.client.Users() {
		if user.Id == b.client.Id {
			b.FullName = user.Name
			b.MentionName = user.MentionName
			return nil
		}
	}
	return errors.New("Couldn't initialise the bot")
}