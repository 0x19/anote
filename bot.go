package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anonutopia/gowaves"
	ui18n "github.com/unknwon/i18n"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const satInBtc = uint64(100000000)

const lang = "hr"

func executeBotCommand(tu TelegramUpdate) {
	if tu.Message.Text == "/price" || strings.HasPrefix(tu.Message.Text, "/price@"+conf.BotName) {
		priceCommand(tu)
	} else if tu.Message.Text == "/start" || strings.HasPrefix(tu.Message.Text, "/start@"+conf.BotName) {
		startCommand(tu)
	} else if tu.Message.Text == "/address" || strings.HasPrefix(tu.Message.Text, "/address@"+conf.BotName) {
		addressCommand(tu)
	} else if tu.Message.Text == "/register" || strings.HasPrefix(tu.Message.Text, "/register@"+conf.BotName) {
		dropCommand(tu)
	} else if tu.Message.Text == "/status" || strings.HasPrefix(tu.Message.Text, "/status@"+conf.BotName) {
		statusCommand(tu)
	} else if tu.Message.Text == "/mine" || strings.HasPrefix(tu.Message.Text, "/mine@"+conf.BotName) {
		mineCommand(tu)
	} else if tu.Message.Text == "/withdraw" || strings.HasPrefix(tu.Message.Text, "/withdraw@"+conf.BotName) {
		withdrawCommand(tu)
	} else if strings.HasPrefix(tu.Message.Text, "/") {
		unknownCommand(tu)
	} else if tu.UpdateID != 0 {
		if tu.Message.ReplyToMessage.MessageID == 0 {
			if tu.Message.NewChatMember.ID != 0 &&
				!tu.Message.NewChatMember.IsBot {
				registerNewUsers(tu)
			}
		} else {
			if tu.Message.ReplyToMessage.Text == ui18n.Tr(lang, "pleaseEnter") {
				avr, err := wnc.AddressValidate(tu.Message.Text)
				if err != nil {
					logTelegram(err.Error())
					messageTelegram(ui18n.Tr(lang, "error"), int64(tu.Message.Chat.ID))
				} else {
					if !avr.Valid {
						messageTelegram(ui18n.Tr(lang, "addressNotValid"), int64(tu.Message.Chat.ID))
					} else {
						tu.Message.Text = fmt.Sprintf("/drop %s", tu.Message.Text)
						dropCommand(tu)
					}
				}
			} else if tu.Message.ReplyToMessage.Text == "Please enter daily code:" {
				tu.Message.Text = fmt.Sprintf("/mine %s", tu.Message.Text)
				mineCommand(tu)
			}
		}
	}
}

func registerNewUsers(tu TelegramUpdate) {
	rUser := &User{TelegramID: tu.Message.From.ID}
	db.First(rUser, rUser)

	for _, user := range tu.Message.NewChatMembers {
		messageTelegram(fmt.Sprintf(ui18n.Tr(lang, "welcome"), tu.Message.NewChatMember.FirstName), int64(tu.Message.Chat.ID))
		now := time.Now()
		u := &User{TelegramID: user.ID,
			TelegramUsername: user.Username,
			ReferralID:       rUser.ID,
			MiningActivated:  &now,
			LastWithdraw:     &now}
		db.FirstOrCreate(u, u)
	}
}

func priceCommand(tu TelegramUpdate) {
	kv := &KeyValue{Key: "tokenPrice"}
	db.First(kv, kv)
	price := float64(kv.ValueInt) / float64(satInBtc)
	messageTelegram(fmt.Sprintf(ui18n.Tr(lang, "currentPrice"), price), int64(tu.Message.Chat.ID))
}

func startCommand(tu TelegramUpdate) {
	now := time.Now()
	u := &User{TelegramID: tu.Message.From.ID,
		TelegramUsername: tu.Message.From.Username,
		MiningActivated:  &now,
		LastWithdraw:     &now}
	db.FirstOrCreate(u, u)

	messageTelegram("Hello and welcome to Anonutopia!", int64(tu.Message.Chat.ID))
}

func addressCommand(tu TelegramUpdate) {
	messageTelegram(ui18n.Tr(lang, "myAddress"), int64(tu.Message.Chat.ID))
	messageTelegram(conf.NodeAddress, int64(tu.Message.Chat.ID))
	var pc tgbotapi.PhotoConfig
	if conf.Dev {
		pc = tgbotapi.NewPhotoUpload(int64(tu.Message.Chat.ID), "qrcodedev.png")
	} else {
		pc = tgbotapi.NewPhotoUpload(int64(tu.Message.Chat.ID), "qrcode.png")
	}
	pc.Caption = "QR Code"
	bot.Send(pc)
}

func dropCommand(tu TelegramUpdate) {
	msgArr := strings.Fields(tu.Message.Text)
	if len(msgArr) == 1 && strings.HasPrefix(tu.Message.Text, "/register") {
		msg := tgbotapi.NewMessage(int64(tu.Message.Chat.ID), ui18n.Tr(lang, "pleaseEnter"))
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: false}
		msg.ReplyToMessageID = tu.Message.MessageID
		bot.Send(msg)
	} else {
		avr, err := wnc.AddressValidate(msgArr[1])
		if err != nil {
			logTelegram(err.Error())
			messageTelegram(ui18n.Tr(lang, "error"), int64(tu.Message.Chat.ID))
		} else {
			if !avr.Valid {
				messageTelegram(ui18n.Tr(lang, "addressNotValid"), int64(tu.Message.Chat.ID))
			} else {
				user := &User{TelegramID: tu.Message.From.ID}
				db.First(user, user)

				if len(user.Address) > 0 {
					if user.Address == msgArr[1] {
						messageTelegram(ui18n.Tr(lang, "alreadyActivated"), int64(tu.Message.Chat.ID))
					} else {
						messageTelegram(ui18n.Tr(lang, "hacker"), int64(tu.Message.Chat.ID))
					}
				} else {
					if msgArr[1] == conf.NodeAddress {
						messageTelegram(ui18n.Tr(lang, "yourAddress"), int64(tu.Message.Chat.ID))
					} else {
						atr := &gowaves.AssetsTransferRequest{
							Amount:    100000000,
							AssetID:   conf.TokenID,
							Fee:       100000,
							Recipient: msgArr[1],
							Sender:    conf.NodeAddress,
						}

						_, err := wnc.AssetsTransfer(atr)
						if err != nil {
							messageTelegram(ui18n.Tr(lang, "error"), int64(tu.Message.Chat.ID))
							logTelegram(err.Error())
						} else {
							user.TelegramID = tu.Message.From.ID
							user.TelegramUsername = tu.Message.From.Username
							user.Address = msgArr[1]
							db.Save(user)

							if user.ReferralID != 0 {
								rUser := &User{}
								db.First(rUser, user.ReferralID)
								if len(rUser.Address) > 0 {
									atr := &gowaves.AssetsTransferRequest{
										Amount:    50000000,
										AssetID:   conf.TokenID,
										Fee:       100000,
										Recipient: rUser.Address,
										Sender:    conf.NodeAddress,
									}

									_, err := wnc.AssetsTransfer(atr)
									if err != nil {
										logTelegram(err.Error())
									} else {
										messageTelegram(fmt.Sprintf(ui18n.Tr(lang, "tokenSentR"), rUser.TelegramUsername), int64(tu.Message.Chat.ID))
									}
								}
							}

							messageTelegram(fmt.Sprintf(ui18n.Tr(lang, "tokenSent"), tu.Message.From.Username), int64(tu.Message.Chat.ID))
						}
					}
				}
			}
		}
	}
}

func statusCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)

	status := user.status()
	mining := user.isMiningStr()
	power := user.miningPowerStr()
	team := user.teamStr()
	teamInactive := user.teamInactiveStr()

	msg := fmt.Sprintf("⭕️  <strong><u>Your Anonutopia / Anote Status</u></strong>\n\n"+
		"<strong>Status:</strong> %s\n"+
		"<strong>Address:</strong> %s\n"+
		"<strong>Mining:</strong> %s\n"+
		"<strong>Mining Power:</strong> %s\n"+
		"<strong>Your Team:</strong> %s\n"+
		"<strong>Inactive:</strong> %s",
		status, user.Address, mining, power, team, teamInactive)

	messageTelegram(msg, int64(tu.Message.Chat.ID))
}

func mineCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)

	msgArr := strings.Fields(tu.Message.Text)
	if len(msgArr) == 1 && strings.HasPrefix(tu.Message.Text, "/mine") {
		msg := tgbotapi.NewMessage(int64(tu.Message.Chat.ID), "Please enter daily code:")
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: false}
		msg.ReplyToMessageID = tu.Message.MessageID
		bot.Send(msg)
	} else {
		now := time.Now()
		user.MiningActivated = &now
		user.Mining = true
		db.Save(user)
		messageTelegram("Great work, you started mining! 🚀", int64(tu.Message.Chat.ID))
	}
}

func withdrawCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)

	if time.Since(*user.LastWithdraw).Hours() < float64(24) {
		messageTelegram("You can do only one withdrawal in 24 hours.", int64(tu.Message.Chat.ID))
	} else if user.MinedAnotes == 0 {
		messageTelegram("You don't have any anotes to withdraw.", int64(tu.Message.Chat.ID))
	} else {
		atr := &gowaves.AssetsTransferRequest{
			Amount:    user.MinedAnotes,
			AssetID:   conf.TokenID,
			Fee:       100000,
			Recipient: user.Address,
			Sender:    conf.NodeAddress,
		}

		_, err := wnc.AssetsTransfer(atr)
		if err != nil {
			log.Printf("[withdraw] error assets transfer: %s", err)
			logTelegram(fmt.Sprintf("[withdraw] error assets transfer: %s", err))
		} else {
			now := time.Now()
			user.LastWithdraw = &now
			user.MinedAnotes = 0
			db.Save(user)
			messageTelegram("I've sent you your mined Anotes. 🚀", int64(tu.Message.Chat.ID))
		}

	}
}

func unknownCommand(tu TelegramUpdate) {
	messageTelegram(ui18n.Tr(lang, "commandNotAvailable"), int64(tu.Message.Chat.ID))
}
