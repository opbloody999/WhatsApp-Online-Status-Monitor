package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
)

// mainHandler: always show all contacts, no selection, prepare all data in Go
func mainHandler(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Contacts []ContactData
	}

	// Fetch all contacts
	var contactsMap map[types.JID]types.ContactInfo
	var err error
	for i := 0; i < 3; i++ {
		contactsMap, err = client.Store.Contacts.GetAllContacts(context.Background())
		if err != nil || contactsMap == nil || len(contactsMap) == 0 {
			log.Printf("[ERROR] Failed to fetch contacts (attempt %d): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	var contacts []ContactData
	mu.RLock()
	for jid, info := range contactsMap {
		name := info.FullName
		if name == "" {
			name = info.PushName
		}
		status := userStatus[jid.String()]
		if status == "" || name == "" {
			status = "Hidden"
		}
		contacts = append(contacts, ContactData{
			CurrentStatus: status,
			Username:      name,
			JID:           jid.String(),
		})
		userNames[jid.String()] = name
		parsedJID, err := types.ParseJID(jid.String())
		if err != nil {
			log.Printf("[ERROR] Failed to parse JID %s: %v", jid.String(), err)
			continue
		}
		if err := client.SubscribePresence(parsedJID); err != nil {
			log.Printf("[ERROR] Failed to subscribe presence for %s: %v", jid.String(), err)
		}
	}
	mu.RUnlock()
	sort.Slice(contacts, func(i, j int) bool {
		return strings.ToLower(contacts[i].Username) < strings.ToLower(contacts[j].Username)
	})

	tmpl := template.Must(template.ParseFiles(filepath.Join("templates", "main.html")))
	err = tmpl.Execute(w, PageData{
		Contacts: contacts,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// statusUpdateHandler returns live status updates as JSON (for future AJAX/live polling if needed)
func statusUpdateHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	updates := []StatusUpdate{}
	for jid, logs := range userStatusLog {
		onlineRanges := calculateOnlineRanges(logs)
		isOnline := false
		if len(logs) > 0 {
			lastLog := logs[len(logs)-1]
			isOnline = lastLog.Status == "Online"
		}
		username := userNames[jid]
		updates = append(updates, StatusUpdate{
			JID:          jid,
			Username:     username,
			OnlineRanges: onlineRanges,
			IsOnline:     isOnline,
		})
	}

	type ContactItem struct {
		JID    string
		Name   string
		Notify string
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}

// historyHandler returns all status history from SQLite as JSON
func historyHandler(w http.ResponseWriter, r *http.Request) {
	jid := r.URL.Query().Get("jid")
	logs := getStatusLogFromDB(jid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}


