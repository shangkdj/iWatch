package services

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

var (
	apnsClient *apns2.Client
	apnsTopic  string
	apnsOnce   sync.Once
)

// APNSConfig APNs 配置
type APNSConfig struct {
	KeyPath    string // .p8 文件路径
	KeyID      string // Key ID
	TeamID     string // Team ID
	BundleID   string // App Bundle ID
	Production bool   // 是否使用生产环境
}

// InitAPNS 初始化 APNs 客户端
func InitAPNS(cfg APNSConfig) error {
	var initErr error

	apnsOnce.Do(func() {
		// 1. 读取 .p8 文件
		p8Content, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			initErr = fmt.Errorf("读取 .p8 文件失败: %v", err)
			return
		}

		// 2. 创建 Token
		authKey, err := token.AuthKeyFromBytes(p8Content)
		if err != nil {
			initErr = fmt.Errorf("解析 .p8 文件失败: %v", err)
			return
		}

		// 3. 创建 APNs 客户端
		tokenClient := &token.Token{
			AuthKey: authKey,
			KeyID:   cfg.KeyID,
			TeamID:  cfg.TeamID,
		}

		if cfg.Production {
			apnsClient = apns2.NewTokenClient(tokenClient).Production()
			log.Println("🚀 APNs 客户端已初始化（生产环境）")
		} else {
			apnsClient = apns2.NewTokenClient(tokenClient).Development()
			log.Println("🚀 APNs 客户端已初始化（开发环境）")
		}

		apnsTopic = cfg.BundleID
	})

	return initErr
}

// PushResult 推送结果
type PushResult struct {
	Success bool
	Reason  string
}

// SendSilentPush 发送静默推送（content-available: 1）
// 用途：唤醒 App 后台刷新数据（如天气、睡眠、心情）
func SendSilentPush(deviceToken string) PushResult {
	if apnsClient == nil {
		return PushResult{Success: false, Reason: "APNs 客户端未初始化"}
	}

	// 构建静默推送 payload
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       apnsTopic,
		Payload:     payload.NewPayload().ContentAvailable(),
		Priority:    apns2.PriorityLow,
	}

	res, err := apnsClient.Push(notification)
	if err != nil {
		log.Printf("❌ APNs 推送失败: %v", err)
		return PushResult{Success: false, Reason: err.Error()}
	}

	if res.StatusCode != 200 {
		log.Printf("❌ APNs 返回错误: %d, %s", res.StatusCode, res.Reason)
		return PushResult{Success: false, Reason: res.Reason}
	}

	log.Printf("✅ APNs 静默推送成功: %s", deviceToken[:20]+"...")
	return PushResult{Success: true}
}

// SendAlertPush 发送通知推送（带标题和内容）
// 用途：直接显示通知（不唤醒 App）
func SendAlertPush(deviceToken, title, body string) PushResult {
	if apnsClient == nil {
		return PushResult{Success: false, Reason: "APNs 客户端未初始化"}
	}

	// 构建通知 payload
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       apnsTopic,
		Payload: payload.NewPayload().
			AlertTitle(title).
			AlertBody(body).
			Sound("default"),
		Priority: apns2.PriorityHigh,
	}

	res, err := apnsClient.Push(notification)
	if err != nil {
		log.Printf("❌ APNs 推送失败: %v", err)
		return PushResult{Success: false, Reason: err.Error()}
	}

	if res.StatusCode != 200 {
		log.Printf("❌ APNs 返回错误: %d, %s", res.StatusCode, res.Reason)
		return PushResult{Success: false, Reason: res.Reason}
	}

	log.Printf("✅ APNs 通知推送成功: %s", deviceToken[:20]+"...")
	return PushResult{Success: true}
}
