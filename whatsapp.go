package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3" // Required for sqlite3 driver
	"github.com/skip2/go-qrcode"
)

// Setup WhatsApp client and handle session loading, connection, etc.
func setupWhatsAppClient(dbPath string) error {
	// Load session data from environment variable if available
	sessionDataB64 := os.Getenv("WHATSAPP_SESSION_DATA_B64")
	if sessionDataB64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(sessionDataB64)
		if err != nil {
			fmt.Printf("Error decoding WHATSAPP_SESSION_DATA_B64: %v\n", err)
		} else {
			// Write the decoded session data to the temporary file
			err = os.WriteFile(dbPath, decoded, 0600)
			if err != nil {
				fmt.Printf("Error writing session data to %s: %v\n", dbPath, err)
			}
			fmt.Println("Loaded WhatsApp session from WHATSAPP_SESSION_DATA_B64.")
		}
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	storeContainer, err := sqlstore.New(context.Background(), "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL", dbPath), dbLog)
	if err != nil {
		log.Printf("Failed to create SQL store: %v", err)
		return err
	}

	device, err := storeContainer.GetFirstDevice(context.Background())
	if err != nil {
		log.Printf("Failed to get device: %v", err)
		return err
	}
	client = whatsmeow.NewClient(device, waLog.Stdout("Client", "DEBUG", true))
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new session
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code (scan with your phone):")
				// Print QR code as ASCII for easier scanning
				printQRCodeASCII(evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			return err
		}
	}

	return nil
}

// printQRCodeASCII prints the QR code as ASCII art in the terminal for easy scanning
func printQRCodeASCII(qrString string) {
	qr, err := qrcode.New(qrString, qrcode.Medium)
	if err != nil {
		fmt.Println("Failed to generate QR code:", err)
		return
	}
	// Print ASCII QR code
	fmt.Println(qr.ToString(false))

	// Save QR code as PNG file
	fileName := "whatsapp_login_qr.png"
	err = qr.WriteFile(256, fileName)
	if err != nil {
		fmt.Println("Failed to save QR code as PNG:", err)
	} else {
		fmt.Printf("QR code also saved as %s\n", fileName)
	}
}
