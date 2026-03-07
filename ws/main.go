package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/atornesi/shared-docs/db"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	h := NewHub()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		claims, err := ParseToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()

		ctx := r.Context()
		ctx = context.WithValue(ctx, "hub", h)
		ctx = context.WithValue(ctx, "claims", claims)

		for {
			err := handleConnection(r.Context(), conn)
			if err != nil {
				log.Println("handleConnection:", err)
				return
			}
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnection(ctx context.Context, conn *websocket.Conn) error {
	var connectionPayload struct {
		DocumentId int64 `json:"room"`
	}

	err := conn.ReadJSON(&connectionPayload)
	if err != nil {
		log.Println("read:", err)
		return err
	}

	dbConn, err := sql.Open("postgres", "postgres://postgres:postgres@db:5432/shared-docs?sslmode=disable")
	if err != nil {
		log.Println("open:", err)
		return err
	}

	claims, ok := ctx.Value("claims").(*Claims)
	if !ok {
		log.Println("no claims")
		return fmt.Errorf("no claims")
	}

	db.New(dbConn).ValidateAccess(ctx, db.ValidateAccessParams{
		ID:      connectionPayload.DocumentId,
		OwnerID: claims.UserID,
	})

	return nil
}
