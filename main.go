package main

import (
    //"code.google.com/p/gcfg" // TODO
    "flag"
    "log"
    "./bot"
    "os"
    "os/signal"
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

    b, err := bot.NewBot(*user, *pass)
    if err != nil {
        log.Fatal(err)
    }

    b.Dunno = func(room *bot.Room) {
        b.Say(room, "I dunno who you are :(")
    }

    r := b.Join(*roomId)
    err = r.AttachPlugin(new(bot.HelloPlugin))

    if err != nil {
        log.Fatal(err)
    }

    handleControlC(b)
}

func handleControlC(bot *bot.Bot) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    for _ = range c {
        log.Println("Disconnecting")
        bot.Disconnect()
        os.Exit(0)
    }
}