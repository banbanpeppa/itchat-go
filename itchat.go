package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/banbanpeppa/itchat-go/model"
	"github.com/banbanpeppa/itchat-go/service"

	"github.com/banbanpeppa/itchat-go/util"
	"github.com/banbanpeppa/itchat-go/wx"
)

func loginAndStoreInfo() {
	handler := wx.NewWxHandler()
	handler.Login()
	for obj := range handler.LoginListener() {
		switch obj.(type) {
		case wx.ListenerCallback:
			callback := obj.(wx.ListenerCallback)
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
	var c wx.ListenerCallback
	util.Load(&c, "wx_cache/itchat.pkl")
	for {
		retcode, selector, err := service.SyncCheck(c.LoginMap)
		if err != nil {
			fmt.Println(retcode, selector)
			fmt.Println(err)
			if retcode == 1101 {
				fmt.Println("帐号已在其他地方登陆，程序将退出。")
				os.Exit(2)
			}
			continue
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func contract() {
	var c wx.ListenerCallback
	util.Load(&c, "wx_cache/itchat.pkl")
	fmt.Println(c.LoginMap)

	contraces, err := service.GetAllContact(c.LoginMap)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("成功获取 %d个 联系人信息,开始整理群组信息...\n", len(contraces))
	for key, value := range contraces {
		if value.RemarkName == "吴烘锐" {
			fmt.Println(key, ":", value, ": ", value.UserName)
		}
	}
}

func group() {
	var c wx.ListenerCallback
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

func send() {
	var c wx.ListenerCallback
	util.Load(&c, "wx_cache/itchat.pkl")
	wxSendMsg := model.WxSendMsg{}
	wxSendMsg.Type = 1
	wxSendMsg.Content = "我是banbanpeppa编写的微信机器人，我已帮你通知我的主人，请您稍等片刻，他会跟您联系"
	wxSendMsg.FromUserName = "@f7f003b1c084665bb6f397f42b0e4a8aff4ae35c96dfce92c7b0af6ac41e735f"
	// wxSendMsg.ToUserName = "@1ee21d885123c7ffc7738a5ba30599c9dd5a0c8b0363473f97d4a9fce333eb24"
	wxSendMsg.ToUserName = "@40fe03d1a0c7969a59f7589f0d4e20fe"
	wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
	wxSendMsg.ClientMsgId = wxSendMsg.LocalID

	/* 加点延时，避免消息次序混乱，同时避免微信侦察到机器人 */
	time.Sleep(time.Second)

	err := service.SendMsg(c.LoginMap, wxSendMsg)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// loginAndStoreInfo()
	// heartbeat()
	// contract()
	// group()
	send()
}
