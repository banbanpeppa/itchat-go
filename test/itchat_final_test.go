package test

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/banbanpeppa/itchat-go/model"
	"github.com/banbanpeppa/itchat-go/service"

	"github.com/banbanpeppa/itchat-go/util"
	"github.com/banbanpeppa/itchat-go/wx"
)

func TestLoginAndStoreInfo(t *testing.T) {
	handler := wx.NewLoginHandler()
	handler.Login()
	for obj := range handler.LoginListener() {
		switch obj.(type) {
		case wx.LoginCallback:
			callback := obj.(wx.LoginCallback)
			if callback.Error == nil && callback.LoginMap != nil {
				fmt.Println("准备持久化:", callback)
				file, err := os.Create("wx_cache/itchat.pkl")
				if err != nil {
					fmt.Println(err)
				}
				enc := gob.NewEncoder(file)
				err = enc.Encode(callback)
				if err != nil {
					fmt.Println(err)
				}
				handler.LoginDone()
				return
			} else {
				fmt.Println(callback.Message)
				continue
			}
		default:
			fmt.Println(obj)
		}
	}
}

func heartbeat() {
	var c wx.LoginCallback
	util.Load(&c, "wx_cache/itchat.pkl")
	for {
		retcode, selector, err := service.SyncCheck(c.LoginMap)
		if err != nil {
			fmt.Println(retcode, selector)
			fmt.Println(err)
			if retcode == 1101 {
				fmt.Println("手机端退出或者帐号已在其他地方登陆，程序将退出。")
				return
			}
			continue
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func TestGetAllGroups(t *testing.T) {
	var c wx.LoginCallback
	util.Load(&c, "wx_cache/itchat.pkl")

	contraces, err := service.GetAllContact(c.LoginMap)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("成功获取 %d个 联系人信息,开始整理群组信息...\n", len(contraces))
	groupMap := service.MapGroupInfo(contraces)
	groupSize := 0
	for _, v := range groupMap {
		groupSize += len(v)
	}
	fmt.Printf("整理完毕，共有 %d个 群组是焦点群组，它们是：\n", groupSize)
	for key, v := range groupMap {
		fmt.Println(key)
		for _, user := range v {
			fmt.Println("========>" + user.NickName)
		}
	}
}

func TestSendMessage(t *testing.T) {
	var c wx.LoginCallback
	util.Load(&c, "wx_cache/itchat.pkl")
	wxSendMsg := model.WxSendMsg{}
	wxSendMsg.Type = 1
	wxSendMsg.Content = "我是banbanpeppa编写的微信机器人，我已帮你通知我的主人，请您稍等片刻，他会跟您联系"
	wxSendMsg.FromUserName = c.LoginMap.SelfUserName
	// wxSendMsg.ToUserName = "@1ee21d885123c7ffc7738a5ba30599c9dd5a0c8b0363473f97d4a9fce333eb24"
	wxSendMsg.ToUserName = "@069c9a42daa16199fc05a607afb2bf4de969aca515472329696bf2223362ff32"
	wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
	wxSendMsg.ClientMsgId = wxSendMsg.LocalID

	/* 加点延时，避免消息次序混乱，同时避免微信侦察到机器人 */
	time.Sleep(time.Second)

	err := service.SendMsg(c.LoginMap, wxSendMsg)
	if err != nil {
		fmt.Println(err)
	}
}

func TestGetLoginQrcode(t *testing.T) {
	handler := wx.NewLoginHandler()
	url, uuid, err := handler.GenerateUrl()
	if err != nil {
		panic(err)
	}
	fmt.Println(url, uuid)

	handler.LoginWithUUID(uuid)
	callback := wx.LoginCallback{}
OuterLoop:
	for obj := range handler.LoginListener() {
		switch obj.(type) {
		case wx.LoginCallback:
			callback = obj.(wx.LoginCallback)
			if callback.Error == nil && callback.LoginMap != nil {
				fmt.Println("准备持久化:", callback)
				//检查并创建临时目录
				if !util.IsDirExist("wx_cache") {
					os.Mkdir("wx_cache", 0755)
					fmt.Println("dir", "wx_cache", "created")
				}
				file, err := os.Create("wx_cache/itchat.pkl")
				if err != nil {
					fmt.Println(err)
				}
				enc := gob.NewEncoder(file)
				err = enc.Encode(callback)
				if err != nil {
					fmt.Println(err)
				}
				handler.LoginDone()
				break OuterLoop
			} else {
				fmt.Println(callback.Message)
				continue
			}
		default:
			fmt.Println(obj)
		}
	}

	fmt.Println("Test Done!")
	itchat := wx.NewItchatHandler(callback.LoginMap)
	itchat.Heartbeat()
}
