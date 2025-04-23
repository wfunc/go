package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
)

var (
	botAPI *tgbotapi.BotAPI
	chatID int64
)

func Bootstrap(token string, chatIDStr string) {
	botAPI, _ = tgbotapi.NewBotAPI(token)
	chatID, _ = strconv.ParseInt(chatIDStr, 10, 64)
}

func SendMessage(msg string) (tgbotapi.Message, error) {
	if botAPI == nil {
		return tgbotapi.Message{}, fmt.Errorf("botAPI is nil")
	}
	message := tgbotapi.NewMessage(chatID, escapeMarkdownV2(msg))
	message.ParseMode = "MarkdownV2"
	return botAPI.Send(message)
}

func SendHTMLMessage(msg string) (tgbotapi.Message, error) {
	if botAPI == nil {
		return tgbotapi.Message{}, fmt.Errorf("botAPI is nil")
	}
	message := tgbotapi.NewMessage(chatID, msg)
	message.ParseMode = "HTML"
	return botAPI.Send(message)
}

func SendDepositMessage(userID int64, quantity decimal.Decimal) {
	if botAPI == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, BuildDepositMessage(strconv.FormatInt(userID, 10), quantity.InexactFloat64()))
	msg.ParseMode = "MarkdownV2"
	botAPI.Send(msg)
}

func SendWithdrawMessage(userID int64, quantity decimal.Decimal) {
	if botAPI == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, BuildWithdrawMessage(strconv.FormatInt(userID, 10), quantity.InexactFloat64()))
	msg.ParseMode = "MarkdownV2"
	botAPI.Send(msg)
}

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.", // <- ✅ 这个必须有
		"!", "\\!",
	)
	return replacer.Replace(text)
}
func BuildDepositMessage(username string, amount float64) string {
	usernameEsc := escapeMarkdownV2(username)
	amountEsc := escapeMarkdownV2(fmt.Sprintf("%.2f", amount))
	timeStr := escapeMarkdownV2(time.Now().Format("2006-01-02 15:04:05"))
	return fmt.Sprintf(
		"💰 *充值发起通知*\n👤 用户：`%s`\n💵 金额：*%s* \n🕒 时间：`%s`",
		usernameEsc, amountEsc, timeStr,
	)
}

func BuildWithdrawMessage(username string, amount float64) string {
	usernameEsc := escapeMarkdownV2(username)
	amountEsc := escapeMarkdownV2(fmt.Sprintf("%.2f", amount))
	timeStr := escapeMarkdownV2(time.Now().Format("2006-01-02 15:04:05"))
	return fmt.Sprintf(
		"💸 *提现申请通知*\n👤 用户：`%s`\n💵 金额：*%s* \n🕒 时间：`%s`",
		usernameEsc, amountEsc, timeStr,
	)
}
