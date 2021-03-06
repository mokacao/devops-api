package controllers

import (
	"fmt"
	"os"

	"github.com/astaxie/beego"
	"github.com/chanyipiaomiao/hltool"
	"github.com/satori/go.uuid"

	"devops-api/common"
)

var (
	uploadPath string
)

// BaseController 基础控制器
type BaseController struct {
	beego.Controller
}

// Prepare 覆盖Controller的方法
func (b *BaseController) Prepare() {

	// 获取客户端IP
	remoteIP := b.Ctx.Input.IP()
	b.Data["RemoteIP"] = remoteIP

	// 从配置文件中获取 RequestID或者TraceID,如果配置文件中没有配置默认就是 RequestId
	uniqueIDName := beego.AppConfig.String("uniqueIDName")
	if uniqueIDName == "" {
		uniqueIDName = "RequestID"
	}
	uniqueID := b.Ctx.Input.Header(uniqueIDName)
	if uniqueID == "" {
		uid, err := uuid.NewV4()
		if err != nil {
			common.GetLogger().Error(map[string]interface{}{
				"entryType": "Get UUID",
			}, fmt.Sprintf("%s", err))
			uniqueID = ""
		} else {
			uniqueID = fmt.Sprintf("%s", uid)
		}
	}
	b.Data["RequestID"] = uniqueID

	// 配置文件文件中启用了token功能,才验证token
	if common.EnableToken {

		// 获取 DEVOPS-API-TOKEN 头信息
		token := b.Ctx.Input.Header("DEVOPS-API-TOKEN")
		if token == "" {
			b.Data["json"] = map[string]string{"result": "need DEVOPS-API-TOKEN header", "statuscode": "1"}
			b.ServeJSON()
			b.StopRun()
		}

		// 验证 DEVOPS-API-TOKEN 是否有效
		jwtoken, err := common.NewToken()
		logFields := map[string]interface{}{
			"entryType": "JWToken Auth",
			"requestID": b.Data["RequestID"],
		}
		if err != nil {
			common.GetLogger().Error(logFields, fmt.Sprintf("%s", err))
			b.Data["json"] = map[string]string{"result": "DEVOPS-API-TOKEN auth fail", "statuscode": "1"}
			b.ServeJSON()
			b.StopRun()
		}

		// 验证是否是root token 不能使用root token
		isroot, err := jwtoken.IsRootToken(token)
		if err != nil {
			common.GetLogger().Error(logFields, fmt.Sprintf("%s", err))
			b.Data["json"] = map[string]string{"result": "DEVOPS-API-TOKEN auth fail", "statuscode": "1"}
			b.ServeJSON()
			b.StopRun()
		}
		if isroot {
			warn := "can't use root token"
			common.GetLogger().Error(logFields, warn)
			b.Data["json"] = map[string]string{"result": warn, "statuscode": "1"}
			b.ServeJSON()
			b.StopRun()
		}

		_, err = jwtoken.IsTokenValid(token)
		if err != nil {
			common.GetLogger().Error(logFields, fmt.Sprintf("%s", err))
			b.Data["json"] = map[string]string{"result": "DEVOPS-API-TOKEN auth fail", "statuscode": "1"}
			b.ServeJSON()
			b.StopRun()
		}
	}

}

// PasswordController 密码管理控制器
type PasswordController struct {
	BaseController
}

// MD5Controller MD5管理控制器
type MD5Controller struct {
	BaseController
}

// EmailController  发送邮件控制器
type EmailController struct {
	BaseController
}

// VersionController 密码管理控制器
type VersionController struct {
	BaseController
}

func init() {

	// 上传目录是否存在
	uploadPath = beego.AppConfig.String("uploadDir")
	if !hltool.IsExist(uploadPath) {
		os.MkdirAll(uploadPath, os.ModePerm)
		os.Create(uploadPath)
	}
}
