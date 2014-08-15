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
	UserInfo   *UserInfo
    Dunno      func(*Room)

    client     *hipchat.Client
}

type UserInfo struct {
    Id          string
    FullName    string
    MentionName string
}

type Message struct {
    From    *UserInfo
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

    b.client.Join(roomJid, b.UserInfo.FullName)

    room := &Room {
        roomJid: roomJid,
    }
    receivedChan := make(chan *Message)

    go func() {
        for message := range b.client.Messages() {
            if strings.HasPrefix(message.Body, "@"+b.UserInfo.MentionName) {
                userInfo := b.GetUserInfo(strings.Split(message.From, "/")[1])
                if userInfo == nil {
                    b.dunno(room)
                    continue
                }
                m := &Message {
                    From: userInfo,
                    Body: message.Body[len(b.UserInfo.MentionName):],
                }
                receivedChan <- m
            }
        }
    }()

    return room, receivedChan
}

func (b *Bot) Say(room *Room, msg string) {
    log.Printf(`Saying: "%s" @ %#v`, msg, room)
    b.client.Say(room.roomJid, b.UserInfo.FullName, msg)
}

func (b *Bot) Disconnect() {
    log.Println("Disconnecting")
    b.client.Status("unavailable")
}

func (b *Bot) GetUserInfo(idToken string) (*UserInfo) {
    for _, user := range b.client.Users() {
        if user.Id == idToken || user.Name == idToken || user.MentionName == idToken {
            return &UserInfo {
                Id: user.Id,
                FullName: user.Name,
                MentionName: user.MentionName}
        }
    }

    return nil
}

func (b *Bot) init() (error) {
    go b.client.KeepAlive()

    botInfo := b.GetUserInfo(b.client.Id)
	
    if botInfo != nil {
        b.UserInfo = botInfo
        return nil
    } else {
	   return errors.New("Couldn't initialise the bot")
    }
}

func (b *Bot) dunno(room *Room) {
    log.Println("dunno()")
    if b.Dunno != nil {
        b.Dunno(room)
    }
}