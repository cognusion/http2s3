package main

import (
	"fmt"
	"github.com/andybons/hipchat"
	"time"
)

// This was a proper callback, but now is just a function called to
// help send a message.
// TODO: Make this a chan listener
func message_callback(sender, to, url, sizeStr string) {
	message := fmt.Sprintf("A %s file from %s has been uploaded to %s", sizeStr, sender, url)
	if to != "0" {
		message = message + " for @" + to
	}
	sendHipchat(GlobalConfig.Get("hipChatToken"), GlobalConfig.Get("hipChatFrom"), GlobalConfig.Get("hipChatRoom"), message, "purple")
}

// Errors go to hipchat too
func hipchatErrorMessage(message string) {
	sendHipchat(GlobalConfig.Get("hipChatToken"), GlobalConfig.Get("hipChatFrom"), GlobalConfig.Get("hipChatRoom"), message, "red")
}

// Send a message to hipchat, retrying every 10s for 120s if it cannot
func sendHipchat(token, from, room, message, color string) (err error) {
	defer Track("sendHipchat", Now(), debugOut)

	c := hipchat.NewClient(token)
	req := hipchat.MessageRequest{
		RoomId:        room,
		From:          from,
		Message:       message,
		Color:         color,
		MessageFormat: hipchat.FormatText,
		Notify:        true,
	}

	// we will make 12 attempts, 10s apart (120s total)
	// to deliver the message before simply giving up.
	// TODO: Deadletter handling
	for i := 0; i < 12; i++ {
		if err = c.PostMessage(req); err == nil {
			// Huzzah
			break
		} else {
			// Fail!
			time.Sleep(10 * time.Second)
		}
	}
	return
}
