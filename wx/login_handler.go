package wx

import (
	"fmt"
	"log"
	"time"

	"github.com/banbanpeppa/itchat-go/model"

	"github.com/banbanpeppa/itchat-go/enum"
	"github.com/banbanpeppa/itchat-go/service"
)

type LoginHandler struct {
	loginListenerClosed bool
	loginListener       chan interface{}
}

type LoginCallback struct {
	Message  string
	LoginMap *model.LoginMap
	Error    error
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{loginListenerClosed: false, loginListener: make(chan interface{})}
}

func (handler *LoginHandler) LoginListener() <-chan interface{} {
	return handler.loginListener
}

func (handler *LoginHandler) LoginDone() {
	handler.loginListenerClosed = true
}

func downloadQrcode() (uuid string, err error) {
	/* 从微信服务器获取UUID */
	uuid, err = service.GetUUIDFromWX()
	if err != nil {
		return "", err
	}

	/* 根据UUID获取二维码 */
	err = service.DownloadImagIntoDir(enum.QRCODE_URL+uuid, "./wx_cache")
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func (handler *LoginHandler) GenerateUrl() (url, uuid string, err error) {
	/* 从微信服务器获取UUID */
	uuid, err = service.GetUUIDFromWX()
	if err != nil {
		return "", "", err
	}

	log.Println("登陆二维码生成成功，二维码地址为：", enum.QRCODE_URL+uuid)
	return enum.QRCODE_URL + uuid, uuid, nil
}

func (handler *LoginHandler) LoginWithUUID(uuid string) {
	go func() {
		handler.handleLogin(uuid)
	}()
}

/*
在执行了下载二维码的操作之后，进入到登录操作，如果说登录操作一直是失败的话，会进入循环，等待最后的信息
成功或者失败与否都会将返回内容写入到管道中，通过Listener获取回调信息
*/
func (handler *LoginHandler) Login() {
	go func() {
		uuid, err := downloadQrcode()
		callback := LoginCallback{}
		if err != nil {
			callback.Error = err
			callback.LoginMap = nil
			callback.Message = fmt.Sprintln("下载二维码失败: ", err)
			handler.loginListener <- callback
			return
		}
		handler.handleLogin(uuid)
	}()
}

func (handler *LoginHandler) handleLogin(uuid string) {
	callback := LoginCallback{}
	/* 轮询服务器判断二维码是否扫过暨是否登陆了 */
	for {
		if handler.loginListenerClosed {
			break
		}
		fmt.Println("正在验证登陆... ...")
		status, msg := service.CheckLogin(uuid)

		if status == 200 {
			fmt.Println("登陆成功,处理登陆信息...")
			loginMap, err := service.ProcessLoginInfo(msg)
			if err != nil {
				log.Println("处理登录信息出错: ", err)
				callback.Error = err
				callback.LoginMap = nil
				callback.Message = fmt.Sprintln("处理登录信息出错: ", err)
				handler.loginListener <- callback
				continue // 不再继续往下执行
			}

			fmt.Println("登陆信息处理完毕,正在初始化微信...")
			err = service.WebInit(&loginMap)
			if err != nil {
				log.Println("初始化微信失败: ", err)
				callback.Error = err
				callback.LoginMap = nil
				callback.Message = fmt.Sprintln("初始化微信失败: ", err)
				handler.loginListener <- callback
				continue // 不再继续往下执行
			}

			fmt.Println("初始化完毕,通知微信服务器登陆状态变更...")
			service.ShowMobileLogin(&loginMap)

			fmt.Println("通知完毕,本次登陆信息：")
			fmt.Println(enum.SKey + "\t\t" + loginMap.BaseRequest.SKey)
			fmt.Println(enum.PassTicket + "\t\t" + loginMap.PassTicket)
			callback.Error = err
			callback.LoginMap = &loginMap
			callback.Message = fmt.Sprintln("通知完毕,本次登陆信息：\n",
				enum.SKey+"\t\t"+loginMap.BaseRequest.SKey, "\n",
				enum.PassTicket+"\t\t"+loginMap.PassTicket)
			handler.loginListener <- callback
		} else if status == 201 {
			fmt.Println("请在手机上确认")
			callback.Error = nil
			callback.LoginMap = nil
			callback.Message = fmt.Sprintln("状态码为201, 用户没有在手机📱上面确认")
			handler.loginListener <- callback
		} else if status == 408 {
			fmt.Println("请扫描二维码")
			callback.Error = nil
			callback.LoginMap = nil
			callback.Message = fmt.Sprintln("状态码为408, 用户没有用手机📱扫描二维码")
			handler.loginListener <- callback
		} else {
			fmt.Println(msg)
			callback.Error = nil
			callback.LoginMap = nil
			callback.Message = msg
			handler.loginListener <- callback
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Println("退出登录微信操作.")
}
