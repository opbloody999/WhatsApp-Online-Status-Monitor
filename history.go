package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// getStatusLogFromDB fetches status history for a specific JID from SQLite
func getStatusLogFromDB(jid string) []StatusLog {
	db, err := sql.Open("sqlite3", "status_history.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return []StatusLog{}
	}
	defer db.Close()

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS status_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		jid TEXT NOT NULL,
		name TEXT,
		status TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		return []StatusLog{}
	}

	query := `SELECT status, timestamp FROM status_history WHERE jid = ? ORDER BY timestamp DESC LIMIT 50`
	rows, err := db.Query(query, jid)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return []StatusLog{}
	}
	defer rows.Close()

	var logs []StatusLog
	for rows.Next() {
		var status string
		var timestamp string
		
		err := rows.Scan(&status, &timestamp)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		
		// Parse timestamp
		t, err := time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			t = time.Now()
		}
		
		logs = append(logs, StatusLog{
			Time:   t,
			Status: status,
		})
	}

	return logs
}

// logStatusToSQLite logs status changes to SQLite database
func logStatusToSQLite(jid, name, status string) {
	db, err := sql.Open("sqlite3", "status_history.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS status_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		jid TEXT NOT NULL,
		name TEXT,
		status TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		return
	}

	insertSQL := `INSERT INTO status_history (jid, name, status) VALUES (?, ?, ?)`
	_, err = db.Exec(insertSQL, jid, name, status)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
	}
}

// calculateOnlineRanges calculates online time ranges from status logs
func calculateOnlineRanges(logs []StatusLog) []OnlineRange {
	var ranges []OnlineRange
	var currentStart *time.Time

	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]
		if log.Status == "Online" && currentStart == nil {
			currentStart = &log.Time
		} else if log.Status == "Offline" && currentStart != nil {
			ranges = append(ranges, OnlineRange{
				Start: *currentStart,
				End:   log.Time,
			})
			currentStart = nil
		}
	}

	// If still online, add range until now
	if currentStart != nil {
		ranges = append(ranges, OnlineRange{
			Start: *currentStart,
			End:   time.Now(),
		})
	}

	return ranges
}

