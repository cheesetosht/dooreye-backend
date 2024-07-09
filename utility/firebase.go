package utility

import (
	"context"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"

	"google.golang.org/api/option"
)

var (
	firebaseApp     *firebase.App
	firebaseOnce    sync.Once
	messagingClient *messaging.Client
	messagingOnce   sync.Once
)

func initFirebase() {
	opt := option.WithCredentialsFile(".dev.firebase.json")
	config := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("!! error initializing firebase: %v\n", err)
	}
	firebaseApp = app
}

func GetFirebaseApp() *firebase.App {
	firebaseOnce.Do(initFirebase)
	return firebaseApp
}

func GetFirebaseMessagingClient() *messaging.Client {
	messagingOnce.Do(func() {
		var err error
		messagingClient, err = GetFirebaseApp().Messaging(context.Background())
		if err != nil {
			log.Fatalf("!! error getting firebase messaging client: %v\n", err)
		}
	})
	return messagingClient
}
