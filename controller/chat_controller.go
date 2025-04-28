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
		log.Println("connected:", s.ID())
		return nil
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
		err := roomController.Repo.DB.Where("id=?", roomID).Find(&room).Error
		if err != nil {
			log.Printf("room is not created")

		}

		if room.ID != "" {
			var messages []models.MessageModels
			err := messageController.Repo.DB.Where("room_id=?", room.ID).Find(&messages).Error
			if err != nil {
				log.Printf("messages is not fetched form the room")
			}
			s.Emit("messages", messages)
			return
		}

		// Step 2: Associate user with room (optional - if you have a join table)
		roomUser := &models.RoomModels{
			ID: roomID,

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
	server.OnEvent("/", "get room", func(s socketio.Conn, userId string) {
		var rooms []models.RoomModels
		var roomInfo []map[string]interface{}
		error := roomController.Repo.DB.Where("user_id1=?", userId).Find(&rooms).Error
		for _, room := range rooms {
			var user models.UserModels
			var message models.MessageModels
			var userStatus models.UserStatusModles
			err := userController.Repo.DB.Where("user_id=?", room.UserId2).Find(&user).Error
			if err != nil {
				log.Printf("while fetching user info %s", err)
			}
			err = userStatusController.Repo.DB.Where("user_id=?", room.UserId2).Find(&userStatus).Error
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
					"lastMessage": nil,
				}
				roomInfo = append(roomInfo, roomUser)
			}
			if message.ID != "" {
				roomUser := map[string]interface{}{
					"userInfo":    userInfo,
					"isOnline":    userStatus.IsOnline,
					"lastMessage": message,
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
