package database

import (
    "database/sql"
    "fmt"
    "math/rand"
    "time"
)

type User struct {
    UID  int
    Nick string
    Language string
}

type Room struct {
    Name        string
    Description string
    OwnerUID    int
    UserCount   int // Ini akan diisi dari Hub
}

func InitDB(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }
    return db, nil
}

func GetOrCreateUser(db *sql.DB, fingerprint string) (User, error) {
    var u User
    err := db.QueryRow("SELECT uid, nickname, language FROM users WHERE pubkey_fingerprint = ?", fingerprint).Scan(&u.UID, &u.Nick, &u.Language)
    
    if err == sql.ErrNoRows {
        // Generate UID Random (1000 - 9999)
        rand.Seed(time.Now().UnixNano())
        newUID := rand.Intn(9000) + 1000
        
        // Generate Suffix Random untuk Anon (misal: 4 karakter hex)
        anonSuffix := fmt.Sprintf("%x", rand.Intn(0xffff))
        newNick := fmt.Sprintf("Anon-%s", anonSuffix)

        _, err = db.Exec("INSERT INTO users (uid, pubkey_fingerprint, nickname, language) VALUES (?, ?, ?, ?)", 
            newUID, fingerprint, newNick, "")
        
        if err != nil {
            return u, err
        }
	return User{UID: newUID, Nick: newNick, Language: ""}, nil
    }
    return u, err
}

func UpdateNickname(db *sql.DB, uid int, newNick string) error {
    _, err := db.Exec("UPDATE users SET nickname = ? WHERE uid = ?", newNick, uid)
    return err
}

func GetActiveRooms(db *sql.DB) ([]Room, error) {
    rows, err := db.Query("SELECT name, description, owner_uid FROM rooms")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var rooms []Room
    for rows.Next() {
        var r Room
        if err := rows.Scan(&r.Name, &r.Description, &r.OwnerUID); err == nil {
            rooms = append(rooms, r)
        }
    }
    return rooms, nil
}

func CreateRoom(db *sql.DB, name string, ownerUID int, desc string) error {
    _, err := db.Exec("INSERT INTO rooms (name, owner_uid, description) VALUES (?, ?, ?)", 
        name, ownerUID, desc)
    return err
}

func DeleteRoom(db *sql.DB, name string) error {
    _, err := db.Exec("DELETE FROM rooms WHERE name = ?", name)
    return err
}

// Fungsi untuk mendapatkan owner_uid dari sebuah room
func GetRoomOwner(db *sql.DB, roomName string) (int, error) {
    var ownerUID int
    err := db.QueryRow("SELECT owner_uid FROM rooms WHERE name = ?", roomName).Scan(&ownerUID)
    return ownerUID, err
}

// Fungsi untuk update owner (Succession)
func UpdateRoomOwner(db *sql.DB, roomName string, newOwnerUID int) error {
    _, err := db.Exec("UPDATE rooms SET owner_uid = ? WHERE name = ?", newOwnerUID, roomName)
    return err
}

func UpdateUserLanguage(db *sql.DB, uid int, lang string) error {
    _, err := db.Exec("UPDATE users SET language = ? WHERE uid = ?", lang, uid)
    return err
}
