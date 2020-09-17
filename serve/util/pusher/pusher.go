package pusher

import (
	"fmt"

	"github.com/pusher/pusher-http-go"
)

type Doer interface {
	Do(channel string, event string, data interface{}) error
}

// Poke messages the specified user's devices via pusher.
func Poke(userID int, doer Doer) {
	err := doer.Do(fmt.Sprintf("u-%d", userID), "poke", "hello")
	if err != nil {
		fmt.Printf("DEBUG: Error sending request: %v\n", err)
	}
}

type RealDoer struct {
}

func (_ RealDoer) Do(channel string, event string, data interface{}) error {
	pusherClient := pusher.Client{
		AppID:   "1074810",
		Key:     "8084fa6056631d43897d",
		Secret:  "22e0a7ee283f5bd7b353",
		Cluster: "us3",
		Secure:  true,
	}
	return pusherClient.Trigger(channel, event, data)
}
