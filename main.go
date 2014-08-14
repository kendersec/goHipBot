package main

import (
    //"code.google.com/p/gcfg" // TODO
    "flag"
    "log"
    "./bot"
)

func main() {
    user := flag.String("u", "", "username")
    pass := flag.String("p", "", "password")
    roomId := flag.String("r", "", "room XMPP name")

    flag.Parse()

    if *user == "" || *pass == "" || *roomId == "" {
        flag.Usage()
        return
    }

    bot, err := bot.NewBot(*user, *pass)
    if err != nil {
        log.Fatal(err)
    }

    room, roomchan := bot.Join(*roomId)

    for msg := range roomchan {
        bot.Say(room, "Hello #" + msg.From)
    }
}