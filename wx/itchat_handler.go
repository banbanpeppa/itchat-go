package wx

import (
	"fmt"
	"log"
	"time"

	"github.com/banbanpeppa/itchat-go/model"
	"github.com/banbanpeppa/itchat-go/service"
)

type ItchatHandler struct {
	loginMap *model.LoginMap
}

func NewItchatHandler(loginMap_ *model.LoginMap) *ItchatHandler {
	return &ItchatHandler{loginMap: loginMap_}
}

func (handler *ItchatHandler) Heartbeat() {
	for {
		retcode, selector, err := service.SyncCheck(handler.loginMap)
		if err != nil {
			log.Println("sync check with code{", retcode, ":", selector, "}", err)
			if retcode == 1101 {
				log.Println("手机端退出或者帐号已在其他地方登陆，程序将退出。")
				return
			}
			continue
		}
		time.Sleep(time.Duration(5) * time.Second) // 5秒进行一次心跳💗检查
	}
}

func (handler *ItchatHandler) Send(toUserName, content string) error {
	wxSendMsg := model.WxSendMsg{}
	wxSendMsg.Type = 1
	wxSendMsg.Content = content
	wxSendMsg.FromUserName = handler.loginMap.SelfUserName
	wxSendMsg.ToUserName = toUserName
	wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
	wxSendMsg.ClientMsgId = wxSendMsg.LocalID

	/* 加点延时，避免消息次序混乱，同时避免微信侦察到机器人 */
	time.Sleep(time.Second)

	err := service.SendMsg(handler.loginMap, wxSendMsg)
	if err != nil {
		log.Println("发送消息失败", err)
		return err
	}
	return nil
}
