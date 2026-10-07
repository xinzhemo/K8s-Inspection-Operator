package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func SendFeishuAlert(webhook, text string) error {
	msg := FeishuMessage{
		MsgType: "text",
	}
	msg.Content.Text = text
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书返回错误状态码: %d", resp.StatusCode)
	}
	return nil
}
