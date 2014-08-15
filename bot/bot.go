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

type Keywords map[string]func(*Room, *Message)

type Plugin interface {
    Keywords() Keywords
}

type pluginData struct {
    keywords    Keywords
    sink        chan *Message
}

type UserInfo struct {
    Id          string
    FullName    string
    MentionName string
}

type Room struct {
    bot         *Bot
    roomJid     string
    plugins     []*pluginData
}

type Message struct {
    From    *UserInfo
    Keyword string
    Body    []string
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

func (b *Bot) Join(roomId string) (*Room) {
	log.Println("Joining room: " + roomId)

	roomJid := fmt.Sprintf("%s@%s", roomId, conf)

    b.client.Join(roomJid, b.UserInfo.FullName)

    room := &Room {
        bot:        b,
        roomJid:    roomJid }

    go func() {
        for message := range b.client.Messages() {
            if strings.HasPrefix(message.Body, "@"+b.UserInfo.MentionName) {
                userInfo := b.GetUserInfo(strings.Split(message.From, "/")[1])
                if userInfo == nil {
                    b.dunno(room)
                    continue
                }

                msg := buildMessage(userInfo, message.Body[len(b.UserInfo.MentionName)+1:])

                for _, plugin := range room.plugins {
                    plugin.sink <- msg
                }
            }
        }
    }()

    return room
}

func buildMessage(userInfo *UserInfo, msg string) (*Message) {
    newMsg := strings.Split(strings.TrimSpace(msg)," ")

    if len(newMsg) < 1 {
        return &Message { From: userInfo }
    }

    return &Message {
        From:   userInfo,
        Keyword:    newMsg[0],
        Body:       newMsg[1:],
    }
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

func (r *Room) AttachPlugin(plugin Plugin) error {
    kw := plugin.Keywords()

    if len(kw) == 0 {
        return errors.New("Empty keywords")
    }

    p := &pluginData {
        keywords:   kw,
        sink:       make(chan *Message)}
    r.plugins = append(r.plugins, p)

    go func() {
        for msg := range p.sink {
            p.sendMessage(r, msg)
        }
    }()

    return nil
}

func (r *Room) Say(msg string) {
    r.bot.Say(r, msg)
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

func (p *pluginData) sendMessage(room *Room, msg *Message) {
    kwFun, exists := p.keywords[msg.Keyword]
    if !exists {
        kwFun, exists = p.keywords[""]
        if !exists {
            return
        }
    }
    kwFun(room, msg)
}

// TODO pulllllllllllllll out

type HelloPlugin struct {}

func (p HelloPlugin) Keywords() Keywords {
    return map[string]func(*Room,*Message) {
        "": func(room *Room, msg *Message) {
            room.Say("Hello @" + msg.From.MentionName)
        },
    }
}