package bot

import (
    "errors"
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

// Bootstrap initializes the default bot client and chat id.
// Returns an error if token is invalid or chatID cannot be parsed.
func Bootstrap(token string, chatIDStr string) error {
    bt, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return err
    }
    id, err := strconv.ParseInt(chatIDStr, 10, 64)
    if err != nil {
        return err
    }
    botAPI = bt
    chatID = id
    return nil
}

func SendMessageWithBot(inputToken string, inputChatID int64, msg string) (tgbotapi.Message, error) {
    bt, err := tgbotapi.NewBotAPI(inputToken)
    if err != nil {
        return tgbotapi.Message{}, err
    }
    message := tgbotapi.NewMessage(inputChatID, escapeMarkdownV2(msg))
    message.ParseMode = "MarkdownV2"
    return bt.Send(message)
}

// SendHTMLMessageWithBot sends an HTML-formatted message with a specific bot token and chat id
func SendHTMLMessageWithBot(inputToken string, inputChatID int64, msg string) (tgbotapi.Message, error) {
    bt, err := tgbotapi.NewBotAPI(inputToken)
    if err != nil {
        return tgbotapi.Message{}, err
    }
    message := tgbotapi.NewMessage(inputChatID, msg)
    message.ParseMode = "HTML"
    return bt.Send(message)
}

func SendMessage(msg string) (tgbotapi.Message, error) {
    if botAPI == nil {
        return tgbotapi.Message{}, errors.New("bot not initialized")
    }
    if chatID == 0 {
        return tgbotapi.Message{}, errors.New("chatID is not set")
    }
    message := tgbotapi.NewMessage(chatID, escapeMarkdownV2(msg))
    message.ParseMode = "MarkdownV2"
    return botAPI.Send(message)
}

func SendHTMLMessage(msg string) (tgbotapi.Message, error) {
    if botAPI == nil {
        return tgbotapi.Message{}, errors.New("bot not initialized")
    }
    if chatID == 0 {
        return tgbotapi.Message{}, errors.New("chatID is not set")
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
