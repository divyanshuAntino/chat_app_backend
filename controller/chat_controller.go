package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/divyanshu050303/chat-app-backend/models"
	"github.com/divyanshu050303/chat-app-backend/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"gorm.io/gorm"
)

// Easier to get running with CORS. Thanks for help @Vindexus and @erkie
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

type RoomController struct {
	Repo *repository.RoomRepository
}
type UserStatusController struct {
	Repo *repository.UserStatusRepository
}
type MessageController struct {
	Repo *repository.MessageRepository
}

func OnSocketConnect(ctx *fiber.Ctx, db *gorm.DB) {
	userStatusRepository := &repository.UserStatusRepository{DB: db}
	userStatusController := &UserStatusController{Repo: userStatusRepository}
	roomRepository := &repository.RoomRepository{DB: db}
	roomController := &RoomController{Repo: roomRepository}
	messageRepository := &repository.MessageRepository{DB: db}
	messageController := &MessageController{Repo: messageRepository}
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "chat message", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		message := data["message"].(string)
		senderID := data["sender_id"].(string)

		// Store message in database
		msg := &models.MessageModels{
			RoomID:   roomID,
			SenderID: senderID,
			Message:  message,
			IsRead:   false,
		}
		err := messageController.Repo.DB.Create(&msg).Error
		if err != nil {
			log.Printf("Error storing message: %v", err)
			return
		}

		// Broadcast the message to all clients in the room
		server.BroadcastToNamespace("/", "chat message", map[string]interface{}{
			"room_id":   roomID,
			"sender_id": senderID,
			"message":   message,
			"timestamp": msg.CreatedAt,
		})
	})
	server.OnEvent("/", "join room", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		sender := data["sender_id"].(string)
		receiver := data["receiver_id"].(string)

		// Step 1: Check if room exists, create if not
		var room models.RoomModels
		err := roomController.Repo.DB.FirstOrCreate(&room, models.RoomModels{ID: roomID}).Error
		if err != nil {
			log.Printf("Error creating or finding room: %v", err)
			return
		}

		// Step 2: Associate user with room (optional - if you have a join table)
		roomUser := &models.RoomModels{
			ID:      uuid.New().String(),
			RoomId:  roomID,
			UserId1: sender,
			UserId2: receiver,
		}
		err = roomController.Repo.DB.FirstOrCreate(&roomUser, roomUser).Error
		if err != nil {
			log.Printf("Error adding user to room: %v", err)
			return
		}

		// Step 3: Update user status to online
		status := &models.UserStatusModles{
			UserID:   sender,
			IsOnline: true,
			LastSeen: time.Now(),
		}
		err = userStatusController.Repo.DB.Save(status).Error
		if err != nil {
			log.Printf("Error updating user status: %v", err)
			return
		}

		// Step 4: Join the room via Socket.IO
		s.Join(roomID)
		log.Printf("User %s joined room %s", sender, roomID)

		// Step 5: Broadcast event to room
		server.BroadcastToNamespace("/", "user joined", map[string]interface{}{
			"room_id": roomID,
			"user_id": sender,
		})
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Client %s disconnected. Reason: %s", s.ID(), reason)
		// Update user status to offline
		status := &models.UserStatusModles{
			UserID:   s.ID(),
			IsOnline: true,
			LastSeen: time.Now(),
		}
		err := userStatusController.Repo.DB.Save(status).Error
		if err != nil {
			log.Printf("Error updating user status: %v", err)
			return
		}
	})

	// Start the Socket.IO server in a goroutine
	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("Socket.IO server error: %s\n", err)
		}
	}()

	// Set up HTTP handler for Socket.IO
	http.Handle("/socket.io/", server)

	// Start HTTP server for Socket.IO on a different port
	log.Println("Socket.IO server starting on :5001...")
	if err := http.ListenAndServe(":5001", nil); err != nil {
		log.Printf("Socket.IO HTTP server error: %s\n", err)
	}
}
