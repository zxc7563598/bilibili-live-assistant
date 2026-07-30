// 测试 B站 API，打印请求结果到控制台。
//
// 运行方式:
//
//	go test -v -run Test ./test/
//
// 或在项目根目录:
//
//	cd /path/to/project && go test -v -run Test ./test/
package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

func newClient() *bilibili.Client {
	opts := []bilibili.Option{bilibili.WithStateFile("bilibili_state.json")}
	opts = append(opts, bilibili.WithDebug(os.Stderr))
	client := bilibili.NewClient(opts...)
	if client.Session() == nil {
		fmt.Println("⚠️  未检测到登录态（Cookie 未加载）")
	} else {
		fmt.Printf("已加载登录态: %s (UID: %d)\n", client.Session().Username, client.Session().UID)
	}
	return client
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// TestGetUserInfo 获取当前登录用户的 nav 信息
// go test -v -run TestGetUserInfo ./test/
func TestGetUserInfo(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	userInfo, err := client.Auth.GetUserInfo(ctx)
	if err != nil {
		t.Fatalf("GetUserInfo 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("UID: %d\n", userInfo.UID)
	fmt.Printf("UName: %s\n", userInfo.UName)
	fmt.Printf("Face: %s\n", userInfo.Face)
	fmt.Printf("IsLogin: %v\n", userInfo.IsLogin)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetWbiKeys 获取 WBI 签名密钥
// go test -v -run TestGetWbiKeys ./test/
func TestGetWbiKeys(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	wbiKeys, err := client.Auth.GetWbiKeys(ctx)
	if err != nil {
		t.Fatalf("GetWbiKeys 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("ImgKey: %s\n", wbiKeys.ImgKey)
	fmt.Printf("SubKey: %s\n", wbiKeys.SubKey)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetBuvid 获取设备指纹 Buvid3/Buvid4
// go test -v -run TestGetBuvid ./test/
func TestGetBuvid(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	buvidInfo, err := client.Auth.GetBuvid(ctx)
	if err != nil {
		t.Fatalf("GetBuvid 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("Buvid3: %s\n", buvidInfo.Buvid3)
	fmt.Printf("Buvid4: %s\n", buvidInfo.Buvid4)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetRealRoomID 获取直播间真实房间号
// go test -v -run TestGetRealRoomID ./test/
func TestGetRealRoomID(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	roomID, err := client.Room.GetRealRoomID(ctx, 1774310)
	if err != nil {
		t.Fatalf("GetRealRoomID 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("roomID: %d\n", roomID)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetRealRoomInfo 获取直播间详细信息
// go test -v -run TestGetRealRoomInfo ./test/
func TestGetRealRoomInfo(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	realRoomInfo, err := client.Room.GetRealRoomInfo(ctx, 1774310)
	if err != nil {
		t.Fatalf("GetRealRoomInfo 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("UID: %d\n", realRoomInfo.UID)
	fmt.Printf("RoomID: %d\n", realRoomInfo.RoomID)
	fmt.Printf("Title: %s", realRoomInfo.Title)
	fmt.Printf("LiveStatus: %d\n", realRoomInfo.LiveStatus)
	fmt.Printf("Online: %d\n", realRoomInfo.Online)
	fmt.Printf("Attention: %d\n", realRoomInfo.Attention)
	fmt.Printf("LiveTime: %s\n", realRoomInfo.LiveTime)
	fmt.Printf("Keyframe: %s\n", realRoomInfo.Keyframe)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetDanmuInfo 获取弹幕 WebSocket 连接信息
// go test -v -run TestGetDanmuInfo ./test/
func TestGetDanmuInfo(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	wbiKeys, err := client.Auth.GetWbiKeys(ctx)
	danmuInfo, err := client.Room.GetDanmuInfo(ctx, 1774310, wbiKeys.ImgKey, wbiKeys.SubKey)
	if err != nil {
		t.Fatalf("GetDanmuInfo 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("Token: %s\n", danmuInfo.Token)
	fmt.Printf("Host: %s\n", danmuInfo.Host)
	fmt.Printf("Port: %d\n", danmuInfo.Port)
	fmt.Printf("WSPort: %d\n", danmuInfo.WSPort)
	fmt.Printf("WSSPort: %d\n", danmuInfo.WSSPort)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetBarragePermission 获取用户在目标直播间的弹幕发送权限
// go test -v -run TestGetBarragePermission ./test/
func TestGetBarragePermission(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	barragePermission, err := client.Room.GetBarragePermission(ctx, 1774310)
	if err != nil {
		t.Fatalf("GetBarragePermission 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("Mode: %d\n", barragePermission.Mode)
	fmt.Printf("Color: %d\n", barragePermission.Color)
	fmt.Printf("Length: %d\n", barragePermission.Length)
	fmt.Printf("Bubble: %d\n", barragePermission.Bubble)
	fmt.Println("================================")
	fmt.Println()
}

// TestSendDanmu 发送弹幕
// go test -v -run TestSendDanmu ./test/
func TestSendDanmu(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	csrf, err := client.CSRF()
	if err != nil {
		t.Fatalf("CSRF 失败: %v", err)
	}
	errDanmu := client.Room.SendDanmu(ctx, 1774310, "测试弹幕发送", csrf)
	if errDanmu != nil {
		t.Fatalf("SendDanmu 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("================================")
	fmt.Println()
}

// TestGetOnlineGoldRank 获取直播间在线金瓜子榜
// go test -v -run TestGetOnlineGoldRank ./test/
func TestGetOnlineGoldRank(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	onlineGoldRank, err := client.Room.GetOnlineGoldRank(ctx, 617459493, 22384516)
	if err != nil {
		t.Fatalf("GetOnlineGoldRank 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("OnlineNum: %d\n", onlineGoldRank.OnlineNum)
	fmt.Printf("Items: %v\n", onlineGoldRank.Items)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetVipNumbers 获取直播间大航海总人数
// go test -v -run TestGetVipNumbers ./test/
func TestGetVipNumbers(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	num, err := client.Room.GetVipNumbers(ctx, 617459493, 22384516)
	if err != nil {
		t.Fatalf("GetVipNumbers 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("num: %d\n", num)
	fmt.Println("================================")
	fmt.Println()
}

// TestAddSilentUser 禁言用户
// go test -v -run TestAddSilentUser ./test/
func TestAddSilentUser(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	csrf, err := client.CSRF()
	if err != nil {
		t.Fatalf("CSRF 失败: %v", err)
	}
	silenErr := client.Room.AddSilentUser(ctx, 1774310, 3494375988398393, "", csrf)
	if silenErr != nil {
		t.Fatalf("AddSilentUser 失败: %v", silenErr)
	}
	fmt.Println()
	fmt.Println("================================")
	fmt.Println()
}

// TestGetSilentUserList 获取直播间禁言用户列表
// go test -v -run TestGetSilentUserList ./test/
func TestGetSilentUserList(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	csrf, err := client.CSRF()
	if err != nil {
		t.Fatalf("CSRF 失败: %v", err)
	}
	// 逐页请求，汇总全部禁言用户
	var allItems []room.SilentUserItem
	page := int64(1)
	for {
		silentList, silentListErr := client.Room.GetSilentUserList(ctx, 1774310, page, csrf)
		if silentListErr != nil {
			t.Fatalf("GetSilentUserList 第%d页失败: %v", page, silentListErr)
		}
		allItems = append(allItems, silentList.Items...)
		fmt.Printf("第%d页: 获取%d条 (总计%d条, 共%d页)\n",
			page, len(silentList.Items), silentList.Total, silentList.TotalPage)

		if page >= int64(silentList.TotalPage) {
			break
		}
		page++
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("共 %d 条禁言记录:\n", len(allItems))
	for i, item := range allItems {
		fmt.Printf("  [%d] ID=%d, UID=%d, Name=%s\n", i+1, item.ID, item.UID, item.Name)
	}
	fmt.Println("================================")
	fmt.Println()
}

// TestDelSilentUser 解除禁言
// go test -v -run TestDelSilentUser ./test/
func TestDelSilentUser(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	csrf, err := client.CSRF()
	if err != nil {
		t.Fatalf("CSRF 失败: %v", err)
	}
	silenErr := client.Room.DelSilentUser(ctx, 1774310, 19337984, csrf)
	if silenErr != nil {
		t.Fatalf("DelSilentUser 失败: %v", silenErr)
	}
	fmt.Println()
	fmt.Println("================================")
	fmt.Println()
}

// TestGetMasterInfo 获取指定 UID 的主播基本信息
// go test -v -run TestGetMasterInfo ./test/
func TestGetMasterInfo(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	masterInfo, err := client.User.GetMasterInfo(ctx, 617459493)
	if err != nil {
		t.Fatalf("GetMasterInfo 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("UID: %d\n", masterInfo.UID)
	fmt.Printf("Name: %s\n", masterInfo.Name)
	fmt.Printf("Face: %s\n", masterInfo.Face)
	fmt.Println("================================")
	fmt.Println()
}

// TestGetStreamerInfo 获取指定 UID 的用户空间信息
// go test -v -run TestGetStreamerInfo ./test/
func TestGetStreamerInfo(t *testing.T) {
	client := newClient()
	ctx, cancel := ctx()
	defer cancel()
	// 请求接口
	wbiKeys, err := client.Auth.GetWbiKeys(ctx)
	streamerInfo, err := client.User.GetStreamerInfo(ctx, 617459493, wbiKeys.ImgKey, wbiKeys.SubKey)
	if err != nil {
		t.Fatalf("GetStreamerInfo 失败: %v", err)
	}
	fmt.Println()
	fmt.Println("========== 返回数据 ==========")
	fmt.Printf("MID: %d\n", streamerInfo.MID)
	fmt.Printf("Name: %s\n", streamerInfo.Name)
	fmt.Printf("Sex: %s\n", streamerInfo.Sex)
	fmt.Printf("Face: %s\n", streamerInfo.Face)
	fmt.Printf("Sign: %s\n", streamerInfo.Sign)
	fmt.Println("================================")
	fmt.Println()
}
