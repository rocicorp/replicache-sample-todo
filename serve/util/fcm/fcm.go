package fcm

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Poke messages the specified user's devices via FCM to ask them to sync.
func Poke(userID int, d Doer) {
	fcmKey := os.Getenv("FCM_SERVER_KEY")
	if fcmKey == "" {
		fmt.Println("DEBUG: Cannot poke user, no fcmKey")
		return
	}
	req, err := http.NewRequest(
		"POST",
		"https://fcm.googleapis.com/fcm/send",
		strings.NewReader(fmt.Sprintf(
			`{"data": {},  "to": "/topics/u-%d"}`, userID)))
	if err != nil {
		fmt.Printf("DEBUG: Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", fcmKey))

	_, err = d.Do(req)
	if err != nil {
		fmt.Printf("DEBUG: Error sending request: %v\n", err)
	}
}
