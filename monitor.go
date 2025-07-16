package main

import (
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Presence:
		jid := v.From.String()
		status := "Online"
		if v.Unavailable {
			status = "Offline"
		}

		mu.Lock()
		userStatus[jid] = status
		userStatusLog[jid] = append(userStatusLog[jid], StatusLog{
			Time:   time.Now().UTC(),
			Status: status,
		})
		name := userNames[jid]
		mu.Unlock()

		// Log to SQLite database
		logStatusToSQLite(jid, name, status)
	}
}

