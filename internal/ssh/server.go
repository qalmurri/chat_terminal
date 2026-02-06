package ssh

import (
    "database/sql" // Tambahkan ini
    "fmt"
    "mager/internal/database" 
    "golang.org/x/crypto/ssh"
)

// Gunakan sql.DB, bukan database.sql.DB
func CreateServerConfig(db *sql.DB) *ssh.ServerConfig {
    return &ssh.ServerConfig{
	PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	    fingerprint := ssh.FingerprintSHA256(key)
	    
	    // Log ini akan muncul di terminal 'go run .' Anda
	    fmt.Printf("Mencoba login: User=%s, Fingerprint=%s\n", conn.User(), fingerprint)
	
	    user, err := database.GetOrCreateUser(db, fingerprint)
	    if err != nil {
	        fmt.Printf("Database error: %v\n", err) // Cek apakah SQLite error
	        return nil, err
	    }
	
	    return &ssh.Permissions{
		    Extensions: map[string]string{
			    "nick": user.Nick,
			    "fp": fingerprint,
		    },
	    }, nil
	},
    }
}
