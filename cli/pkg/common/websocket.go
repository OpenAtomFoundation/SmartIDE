package common

import (
	"encoding/json"
	"fmt"
	"runtime"

	"sync"
	"time"

	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

// 定义函数Map类型，便于后续快捷使用
type ControllerMapsType map[string]reflect.Value

type ClientMethod struct {
}

const (
	SVR_KEY_BUSINESS = "business" //业务服务
	SVR_KEY_LOGIC    = "logic"    //逻辑服务
)

type conn struct {
	c   *websocket.Conn
	mtx *sync.Mutex
}

var (
	addr     string                               //websocket地址
	wsclient = conn{mtx: new(sync.Mutex), c: nil} // websocket client

	httpReqMsgReceiver = receiver{
		cmdMtx:    new(sync.Mutex),
		curMaxCmd: 0,

		receiverMtx:       new(sync.RWMutex),
		maxReceiver:       100,
		receiver:          make(map[int]chan []byte),
		receiveMsgTimeout: 15 * time.Second,
	}

	clientArr = make(map[string]*websocket.Conn)
)

type receiver struct {
	curMaxCmd int
	cmdMtx    *sync.Mutex

	receiver    map[int]chan []byte
	receiverMtx *sync.RWMutex
	maxReceiver int

	receiveMsgTimeout time.Duration
}

/**
 * Description: 下划线写法转为驼峰写法
 * author: 	kenanlu@leansoftx.com
 * param: 	name
 * create on:	2021-04-17 08:56:25
 * return: 	string
 */
func case2Camel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

/**
 * Description: websocket客户端接收数据指令调用对应函数
 * author: 	kenanlu@leansoftx.com
 * create on:	2021-04-16 18:05:21
 */
func (w *receiver) ClientCodeToFunc(data baseMsg) {
	funcName := case2Camel(data.Actioncode)
	vft := w.serverReturnFunc()

	params := make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(data)
	if vft[funcName].IsValid() {
		vft[funcName].Call(params)
	}
}

/**
 * Description: 查询结构体中的方法
 * author: 	kenanlu@leansoftx.com
 * create on:	2021-04-17 11:47:12
 * return: 	ControllerMapsType
 */
func (w *receiver) serverReturnFunc() ControllerMapsType {
	var m ClientMethod
	vf := reflect.ValueOf(&m)
	vft := vf.Type()
	//读取方法数量
	mNum := vf.NumMethod()
	crMap := make(ControllerMapsType, 0)

	//遍历所有的方法，并将其存入映射变量中
	for i := 0; i < mNum; i++ {
		mName := vft.Method(i).Name
		crMap[mName] = vf.Method(i)
	}
	return crMap
}
func (w *receiver) pushMap(k int, c chan []byte) int {
	w.receiverMtx.Lock()
	if _, ok := w.receiver[k]; ok {
		return 51003
	}
	w.receiver[k] = c
	defer w.receiverMtx.Unlock()

	return 200
}

func (w *receiver) deleteMap(k int) {
	w.receiverMtx.Lock()
	if _, ok := w.receiver[k]; ok {
		if len(w.receiver[k]) > 0 {
			<-w.receiver[k]
		}
		close(w.receiver[k])
		delete(w.receiver, k)
	}
	defer w.receiverMtx.Unlock()
}

func (w *receiver) getCmd() (v int) {
	w.cmdMtx.Lock()
	w.curMaxCmd += 1
	v = w.curMaxCmd
	defer w.cmdMtx.Unlock()
	return
}

func (w *receiver) recoverCmd(v int) {
	if len(w.receiver) == 0 {
		w.cmdMtx.Lock()
		w.curMaxCmd = 0
		defer w.cmdMtx.Unlock()
		return
	}

	if v == w.curMaxCmd {
		w.cmdMtx.Lock()
		w.curMaxCmd -= 1
		defer w.cmdMtx.Unlock()
	}
}

var clientManager = map[string]string{
	"business": "",
}

func WebsocketStart(wsURL string) {
	clientManager["business"] = wsURL
	for k, v := range clientManager {
		go connServer(k, v)
	}

}

type baseMsg struct {
	Company       string      `json:"company"`
	Actioncode    string      `json:"actioncode"`
	Data          interface{} `json:"data"`
	ModID         string      `json:"modID"`
	Token         string      `json:"token"`
	Result        int         `json:"result"`
	ResultMessage string      `json:"result_message"`
	CmdSequence   int         `json:"CmdSequence"`
}

func connServer(key, Addr string) {
	var err error
	defer func() {
		wsclient.mtx.Lock()
		if clientArr[key] != nil {
			_ = clientArr[key].Close()
		}

		clientArr[key] = nil
		defer wsclient.mtx.Unlock()
		WSLog(Addr + " 自动重连")
		//自动重连机制
		time.Sleep(3 * time.Second)
		connServer(key, Addr)
	}()
	wsclient.c, _, err = websocket.DefaultDialer.Dial(Addr, nil)

	if err != nil {
		return
	}
	clientArr[key] = wsclient.c

	for {
		var res []byte
		_, res, err := clientArr[key].ReadMessage()
		if err != nil {
			WSLog("数据读取失败")
			return
		}
		var msg baseMsg
		err = json.Unmarshal(res, &msg)
		if err != nil {
			WSLog("数据解析失败")
			return
		}
		SmartIDELog.ConsoleDebug(fmt.Sprintf("服务器推送数据：%+v\n", msg))
		httpReqMsgReceiver.ClientCodeToFunc(msg)
		// httpReqMsgReceiver.receiveMsg(msg.CmdSequence, res)
	}
}

// 发送给固定客户端
func (w *receiver) receiveMsg(cmd int, msg []byte) {
	w.receiverMtx.RLock()
	defer w.receiverMtx.RUnlock()
	if _, ok := w.receiver[cmd]; ok {
		go func() {
			select {
			case w.receiver[cmd] <- msg:
			case <-time.After(1 * time.Second):
			}
		}()
	}
}

/*
* description: 发送数据到各端
* author: kenan
* created on: 2021/4/8 11.03
* param key: business--表示发送给业务服务器，logic--表示逻辑服务器
* param Actioncode: 发送指令
* param ModID: 站点ID（3位前补零）+服务器模块类型（2位前补零），示例为1号站点服务器的数据服务
* param Token: 登录token
* param Data: 发送的具体数据
* return return_1:
 */
func SendAndReceive(key, Actioncode, ModID, Token string, Data interface{}) (data baseMsg, code int) {
	if httpReqMsgReceiver.maxReceiver < len(httpReqMsgReceiver.receiver) {
		return data, 51002
	}

	cmd := httpReqMsgReceiver.getCmd()
	ch := make(chan []byte)
	pErr := httpReqMsgReceiver.pushMap(cmd, ch)

	defer func() {
		httpReqMsgReceiver.deleteMap(cmd)
		httpReqMsgReceiver.recoverCmd(cmd)
	}()

	if pErr != 200 {
		return data, 51004
	}

	msg := baseMsg{
		Company:    "SmartIDE",
		Actioncode: Actioncode,
		Data:       Data,
		ModID:      ModID,
		Token:      Token,
	}

	msgByte, _ := json.Marshal(msg)
	wsclient.mtx.Lock()
	if clientArr[key] != nil {
		sErr := clientArr[key].WriteMessage(websocket.BinaryMessage, msgByte)
		defer wsclient.mtx.Unlock()
		if sErr != nil {
			return data, 51004
		}

		select {
		case res := <-ch:
			var msg baseMsg
			err := json.Unmarshal([]byte(res), &msg)
			if err != nil {
				WSLog("数据解析失败")
				return data, 51004
			}
			SmartIDELog.ConsoleDebug(fmt.Sprintf("服务器推送数据：%+v\n", msg))
			return msg, msg.Result
		case <-time.After(httpReqMsgReceiver.receiveMsgTimeout):
			SmartIDELog.ConsoleDebug(fmt.Sprintf("接收发送数据：连接超时\n"))
			WSLog("连接超时")
			return data, 51005
		}
	}

	defer wsclient.mtx.Unlock()
	return data, 51006
}

/*
* description: ws接口日志记录器
* author: kenan
* created on: 2021-04-08 上午11:20
* param param_1:
* param param_2:
* return return_1:
 */
func WSLog(msg string, params ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	SmartIDELog.ConsoleDebug(fmt.Sprintf("[WS] | %s:%d | %s\n", functionName, line, fmt.Sprintf(msg, params...)))
}
