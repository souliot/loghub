package ws

import (
	"loghub/utils"
	"net/url"
	"public/libs_go/gateway/master"
	"strconv"

	"github.com/gin-gonic/gin"
	sutil "github.com/souliot/siot-util"
)

type BaseController struct {
}

func (c *BaseController) DefaultBool(ctx *gin.Context, key string, defaultValue bool) (b bool) {
	v := ctx.Query(key)
	if v == "" {
		return defaultValue
	}
	b, _ = strconv.ParseBool(v)
	return
}

func (c *BaseController) DefaultInt(ctx *gin.Context, key string, defaultValue int) (i int) {
	v := ctx.Query(key)
	if v == "" {
		return defaultValue
	}
	i, _ = strconv.Atoi(v)
	return
}

func (c *BaseController) DefaultInt32(ctx *gin.Context, key string, defaultValue int32) (i int32) {
	v := ctx.Query(key)
	if v == "" {
		return defaultValue
	}
	d, _ := strconv.Atoi(v)
	return int32(d)
}

func (c *BaseController) DefaultInt64(ctx *gin.Context, key string, defaultValue int64) (i int64) {
	v := ctx.Query(key)
	if v == "" {
		return defaultValue
	}
	i, _ = strconv.ParseInt(v, 10, 64)
	return
}

func (c *BaseController) HandlerNoRouter(ctx *gin.Context) {
	errC := sutil.Err404
	errC.MoreInfo = "访问页面不存！"
	ctx.JSON(200, errC)
	return
}

func (c *BaseController) GetClientID(ctx *gin.Context) (id string, key string, ok bool) {
	id = ctx.Request.Header.Get("X-Real-IP")
	if id == "" {
		u, err := url.Parse("//" + ctx.Request.RemoteAddr)
		if err == nil {
			id = u.Hostname()
		}
	}
	if id == "" {
		id = master.GetID()
	}
	typ := ctx.Query("typ")
	instance := ctx.Query("instance")
	if id == "" || typ == "" || instance == "" {
		ok = false
		return
	}
	id = utils.StringJion(id, "_", typ, instance)
	key = utils.StringJion(typ, "_", instance)
	ok = true
	return
}
