package chat

import "sync"

type RoomManager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) GetRoom(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Create the room if it doesn't exist
	if _, exists := rm.rooms[name]; !exists {
		rm.rooms[name] = NewRoom()
	}
	return rm.rooms[name]
}

// internal/chat/room_manager.go
func (rm *RoomManager) ListRooms() []*Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rooms := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
