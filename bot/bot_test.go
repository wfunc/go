package bot

import (
    "testing"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestSendMessage_NotInitialized(t *testing.T) {
    // ensure uninitialized state
    botAPI = nil
    chatID = 0
    if _, err := SendMessage("hello"); err == nil {
        t.Fatalf("expected error when bot not initialized")
    }
}

func TestSendMessage_ChatIDZero(t *testing.T) {
    // simulate initialized bot client but chatID not set
    botAPI = &tgbotapi.BotAPI{}
    chatID = 0
    if _, err := SendMessage("hello"); err == nil {
        t.Fatalf("expected error when chatID is zero")
    }
}

func TestSendHTMLMessage_NotInitialized(t *testing.T) {
    botAPI = nil
    chatID = 0
    if _, err := SendHTMLMessage("<b>hello</b>"); err == nil {
        t.Fatalf("expected error when bot not initialized (HTML)")
    }
}

func TestSendHTMLMessage_ChatIDZero(t *testing.T) {
    botAPI = &tgbotapi.BotAPI{}
    chatID = 0
    if _, err := SendHTMLMessage("<b>hello</b>"); err == nil {
        t.Fatalf("expected error when chatID is zero (HTML)")
    }
}
