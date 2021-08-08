package middleware

import (
	"github.com/gogf/gf/net/ghttp"
	"github.com/iWinston/qk-library/frame/qservice"
	"gorm.io/gorm"
)

// 事务
func TX(r *ghttp.Request) {
	db := qservice.ReqContext.GetDB(r.Context())
	db.Transaction(func(tx *gorm.DB) error {
		ctx := qservice.ReqContext.Get(r.Context())
		ctx.TX = tx
		ctx.OrgTX = tx
		r.Middleware.Next()
		return qservice.ReqContext.GetError(r.Context())
	})
}
