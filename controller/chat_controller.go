package controller

import (
	"fmt"
	"log"
	"net/http"

	// "strings"
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
type UserInfoController struct {
	Repo *repository.UserRepository
}

func OnSocketConnect(ctx *fiber.Ctx, db *gorm.DB) {
	userStatusRepository := &repository.UserStatusRepository{DB: db}
	userStatusController := &UserStatusController{Repo: userStatusRepository}
	roomRepository := &repository.RoomRepository{DB: db}
	roomController := &RoomController{Repo: roomRepository}
	messageRepository := &repository.MessageRepository{DB: db}
	messageController := &MessageController{Repo: messageRepository}
	userRepository := &repository.UserRepository{DB: db}
	userController := &UserInfoController{Repo: userRepository}
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

		return nil
	})
	server.OnEvent("/", "user-online", func(s socketio.Conn, data map[string]interface{}) {
		userId := data["userId"].(string)
		status := &models.UserStatusModles{
			UserID:   userId,
			IsOnline: true,
			LastSeen: time.Now(),
		}
		err := userStatusController.Repo.DB.Save(status).Error
		if err != nil {
			log.Printf("Error updating user status: %v", err)
			return
		}
		userStatus := map[string]interface{}{
			"userId":   userId,
			"isOnline": status.IsOnline,
			"lastSeen": status.LastSeen,
		}
		s.Emit("get room", userId)
		s.Emit("user-online", userStatus)

	})
	server.OnEvent("/", "user-offline", func(s socketio.Conn, data map[string]interface{}) {
		userId := data["userId"].(string)
		status := &models.UserStatusModles{
			UserID:   userId,
			IsOnline: false,
			LastSeen: time.Now(),
		}
		err := userStatusController.Repo.DB.Save(status).Error
		if err != nil {
			log.Printf("Error updating user status: %v", err)
			return
		}
		userStatus := map[string]interface{}{
			"userId":   userId,
			"isOnline": status.IsOnline,
			"lastSeen": status.LastSeen,
		}
		s.Emit("get room", userId)
		s.Emit("user-online", userStatus)

	})
	server.OnEvent("/", "chat message", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		message := data["message"].(string)
		senderID := data["sender_id"].(string)

		// Store message in database
		msg := &models.MessageModels{
			ID:       uuid.New().String(),
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

		// Broadcast to all in the room including sender
		server.BroadcastToRoom("/", roomID, "new_message", map[string]interface{}{
			"room_id":    roomID,
			"sender_id":  senderID,
			"message":    message,
			"created_at": msg.CreatedAt.Format(time.RFC3339),
		})
	})
	server.OnEvent("/", "join room", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		sender := data["sender_id"].(string)
		receiver := data["receiver_id"].(string)

		// Check/create room
		var room models.RoomModels
		err := roomController.Repo.DB.Where("id = ?", roomID).First(&room).Error
		if err != nil {
			// Create new room if not exists
			room = models.RoomModels{
				ID:      roomID,
				UserId1: sender,
				UserId2: receiver,
			}
			if err := roomController.Repo.DB.Create(&room).Error; err != nil {
				log.Printf("Error creating room: %v", err)
				return
			}
		}

		s.Join(roomID)
		log.Printf("User %s joined room %s", sender, roomID)

		var messages []models.MessageModels
		if err := messageController.Repo.DB.
			Where("room_id = ?", roomID).
			Order("created_at asc").
			Find(&messages).Error; err != nil {
			log.Printf("Error fetching messages: %v", err)
		} else {
			s.Emit("messages", messages)
		}
	})
	server.OnEvent("/", "get room", func(s socketio.Conn, userId string) {
		var rooms []models.RoomModels
		var roomInfo []map[string]interface{}
		error := roomController.Repo.DB.Where("user_id1 = ? OR user_id2 = ?", userId, userId).Find(&rooms).Error
		fmt.Print(rooms)
		if error != nil {
			log.Printf("while fetching room info %s", error)
		}
		for _, room := range rooms {
			var user models.UserModels
			var message models.MessageModels
			var userStatus models.UserStatusModles
			otherUserId := room.UserId1
			if room.UserId1 == userId {
				otherUserId = room.UserId2
			}

			err := userController.Repo.DB.Where("user_id=?", otherUserId).Find(&user).Error
			if err != nil {
				log.Printf("while fetching user info %s", err)
			}
			err = userStatusController.Repo.DB.Where("user_id=?", otherUserId).Find(&userStatus).Error
			if err != nil {
				log.Printf("while fetching user info %s", err)
			}
			userInfo := map[string]interface{}{
				"userId":    user.UserId,
				"userName":  user.Name,
				"userImage": user.UserImage,
			}
			err = messageController.Repo.DB.Where("room_id = ?", room.ID).Order("created_at desc").First(&message).Error
			if err != nil {
				log.Printf("while fetching user info %s", err)
			}
			if message.ID == "" {
				roomUser := map[string]interface{}{
					"userInfo":    userInfo,
					"isOnline":    userStatus.IsOnline,
					"lastSeen":    userStatus.LastSeen,
					"lastMessage": nil,
					"roomID":      room.ID,
				}
				roomInfo = append(roomInfo, roomUser)
			}
			if message.ID != "" {
				roomUser := map[string]interface{}{
					"userInfo":    userInfo,
					"isOnline":    userStatus.IsOnline,
					"lastMessage": message,
					"roomID":      room.ID,
					"lastSeen":    userStatus.LastSeen,
				}
				roomInfo = append(roomInfo, roomUser)
			}

		}
		if error != nil {
			log.Printf("room is not fetched %s", error)
		}

		s.Emit("room_list", roomInfo)
	})
	server.OnEvent("/", "new message", func(s socketio.Conn, roomId string) {
		var message models.MessageModels
		err := messageController.Repo.DB.Where("room_id = ?", roomId).Order("created_at desc").First(&message).Error
		if err != nil {
			log.Printf("message is not fetched due to %s", err)
		}

		s.Emit("new_message", message)

	})

	server.OnEvent("/", "typing", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		senderID := data["sender_id"].(string)
		server.BroadcastToRoom("/", roomID, "user_typing", map[string]interface{}{
			"sender_id": senderID,
			"is_typing": true,
		})
	})
	server.OnEvent("/", "stop_typing", func(s socketio.Conn, data map[string]interface{}) {
		roomID := data["room_id"].(string)
		senderID := data["sender_id"].(string)
		s.Emit("user_stop_typing", map[string]interface{}{
			"sender_id": senderID,
			"is_typing": false,
		})
		server.BroadcastToRoom("/", roomID, "user_stop_typing", map[string]interface{}{
			"sender_id": senderID,
			"is_typing": false,
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
