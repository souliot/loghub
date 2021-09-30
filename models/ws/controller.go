package ws

import (
	"net/http"

	sutil "github.com/souliot/siot-util"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsController struct {
	BaseController
}

func (c *WsController) Logs(ctx *gin.Context) {
	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	id, key, ok := c.GetClientID(ctx)
	if !ok {
		ws.WriteJSON(sutil.ErrUserInput)
		ws.Close()
		return
	}
	Ws.AddClient(WsLog, ws, id, key)
}
