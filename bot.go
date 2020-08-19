package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/anonutopia/gowaves"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const satInBtc = uint64(100000000)

const langHr = "hr"
const lang = "en-US"

func executeBotCommand(tu TelegramUpdate) {
	if tu.Message.Text == "/price" || strings.HasPrefix(tu.Message.Text, "/price@"+conf.BotName) {
		priceCommand(tu)
	} else if tu.Message.Text == "/start" || strings.HasPrefix(tu.Message.Text, "/start@"+conf.BotName) {
		startCommand(tu)
	} else if tu.Message.Text == "/address" || strings.HasPrefix(tu.Message.Text, "/address@"+conf.BotName) {
		addressCommand(tu)
	} else if tu.Message.Text == "/register" || strings.HasPrefix(tu.Message.Text, "/register@"+conf.BotName) {
		if tu.Message.Chat.Type != "private" {
			messageTelegram(tr(tu.Message.Chat.ID, "usePrivate"), int64(tu.Message.Chat.ID))
			return
		}
		dropCommand(tu)
	} else if tu.Message.Text == "/status" || strings.HasPrefix(tu.Message.Text, "/status@"+conf.BotName) {
		statusCommand(tu)
	} else if tu.Message.Text == "/mine" || strings.HasPrefix(tu.Message.Text, "/mine@"+conf.BotName) {
		if tu.Message.Chat.Type != "private" {
			messageTelegram(tr(tu.Message.Chat.ID, "usePrivate"), int64(tu.Message.Chat.ID))
			return
		}
		mineCommand(tu)
	} else if tu.Message.Text == "/withdraw" || strings.HasPrefix(tu.Message.Text, "/withdraw@"+conf.BotName) {
		withdrawCommand(tu)
	} else if tu.Message.Text == "/shoutinfo" || strings.HasPrefix(tu.Message.Text, "/shoutinfo@"+conf.BotName) {
		shoutinfoCommand(tu)
	} else if strings.HasPrefix(tu.Message.Text, "/") {
		unknownCommand(tu)
	} else if tu.UpdateID != 0 {
		if tu.Message.ReplyToMessage.MessageID == 0 {
			if tu.Message.NewChatMember.ID != 0 &&
				!tu.Message.NewChatMember.IsBot {
				registerNewUsers(tu)
			}
		} else {
			if tu.Message.ReplyToMessage.Text == tr(tu.Message.Chat.ID, "pleaseEnter") {
				avr, err := wnc.AddressValidate(tu.Message.Text)
				if err != nil {
					logTelegram(err.Error())
					messageTelegram(tr(tu.Message.Chat.ID, "error"), int64(tu.Message.Chat.ID))
				} else {
					if !avr.Valid {
						messageTelegram(tr(tu.Message.Chat.ID, "addressNotValid"), int64(tu.Message.Chat.ID))
					} else {
						tu.Message.Text = fmt.Sprintf("/drop %s", tu.Message.Text)
						dropCommand(tu)
					}
				}
			} else if tu.Message.ReplyToMessage.Text == tr(tu.Message.Chat.ID, "dailyCode") {
				tu.Message.Text = fmt.Sprintf("/mine %s", tu.Message.Text)
				mineCommand(tu)
			} else if tu.Message.ReplyToMessage.Text == tr(tu.Message.Chat.ID, "shoutMessage") {
				ss.setMessage(tu)
			} else if tu.Message.ReplyToMessage.Text == tr(tu.Message.Chat.ID, "shoutLink") {
				ss.setLink(tu)
			}
		}
	}
}

func shoutinfoCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)
	var shout Shout
	db.Where("finished = true and published = false").Order("price desc").First(&shout)
	price := float64(shout.Price) / float64(satInBtc)
	messageTelegram(fmt.Sprintf(tr(user.TelegramID, "shoutInfo"), price), int64(tu.Message.Chat.ID))
}

func registerNewUsers(tu TelegramUpdate) {
	rUser := &User{TelegramID: tu.Message.From.ID}
	db.First(rUser, rUser)

	for _, user := range tu.Message.NewChatMembers {
		messageTelegram(fmt.Sprintf(tr(tu.Message.Chat.ID, "welcome"), tu.Message.NewChatMember.FirstName), int64(tu.Message.Chat.ID))
		var name string
		var lng string
		if tu.Message.Chat.ID == tAnonBalkan {
			lng = langHr
		} else {
			lng = lang
		}
		if len(user.Username) > 0 {
			name = user.Username
		} else {
			name = user.FirstName
		}
		now := time.Now()
		if rUser.ID == 0 {
			rUser.ID = 1
		}
		u := &User{TelegramID: user.ID,
			TelegramUsername: name,
			ReferralID:       rUser.ID,
			LastStatus:       &now,
			LastWithdraw:     &now,
			Language:         lng}
		if err := db.FirstOrCreate(u, u).Error; err != nil {
			logTelegram(err.Error())
		}
	}
}

func priceCommand(tu TelegramUpdate) {
	u := &User{TelegramID: tu.Message.From.ID}
	db.First(u, u)
	kv := &KeyValue{Key: "tokenPrice"}
	db.First(kv, kv)
	price := float64(kv.ValueInt) / float64(satInBtc)
	messageTelegram(fmt.Sprintf(tr(u.TelegramID, "currentPrice"), price), int64(tu.Message.Chat.ID))
}

func startCommand(tu TelegramUpdate) {
	now := time.Now()
	u := &User{TelegramID: tu.Message.From.ID,
		TelegramUsername: tu.Message.From.Username,
		LastStatus:       &now,
		LastWithdraw:     &now,
		Language:         "en-US"}

	if err := db.FirstOrCreate(u, u).Error; err != nil {
		logTelegram(err.Error())
	}

	if u.ReferralID == 0 {
		r := &User{}
		db.First(r, 1)
		log.Println(r)
		u.Referral = r
		if err := db.Save(u).Error; err != nil {
			logTelegram(err.Error())
		}
	}

	messageTelegram(strings.Replace(tr(u.TelegramID, "hello"), "\\n", "\n", -1), int64(tu.Message.Chat.ID))
}

func addressCommand(tu TelegramUpdate) {
	u := &User{TelegramID: tu.Message.From.ID}
	db.First(u, u)
	messageTelegram(tr(u.TelegramID, "myAddress"), int64(tu.Message.Chat.ID))
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
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)
	msgArr := strings.Fields(tu.Message.Text)
	if len(msgArr) == 1 && strings.HasPrefix(tu.Message.Text, "/register") {
		msg := tgbotapi.NewMessage(int64(tu.Message.Chat.ID), tr(user.TelegramID, "pleaseEnter"))
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: false}
		msg.ReplyToMessageID = tu.Message.MessageID
		bot.Send(msg)
	} else {
		avr, err := wnc.AddressValidate(msgArr[1])
		if err != nil {
			logTelegram(err.Error())
			messageTelegram(tr(user.TelegramID, "error"), int64(tu.Message.Chat.ID))
		} else {
			if !avr.Valid {
				messageTelegram(tr(user.TelegramID, "addressNotValid"), int64(tu.Message.Chat.ID))
			} else {
				if len(user.Address) > 0 {
					if user.Address == msgArr[1] {
						messageTelegram(tr(user.TelegramID, "alreadyActivated"), int64(tu.Message.Chat.ID))
					} else {
						messageTelegram(tr(user.TelegramID, "hacker"), int64(tu.Message.Chat.ID))
					}
				} else if user.ReferralID == 0 {
					link := fmt.Sprintf("https://%s/%s/%d", conf.Hostname, msgArr[1], tu.Message.From.ID)
					messageTelegram(fmt.Sprintf(tr(user.TelegramID, "clickLink"), link), int64(tu.Message.Chat.ID))
				} else {
					if msgArr[1] == conf.NodeAddress {
						messageTelegram(tr(user.TelegramID, "yourAddress"), int64(tu.Message.Chat.ID))
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
							messageTelegram(tr(user.TelegramID, "error"), int64(tu.Message.Chat.ID))
							logTelegram(err.Error())
						} else {
							user.TelegramID = tu.Message.From.ID
							user.TelegramUsername = tu.Message.From.Username
							user.Address = msgArr[1]
							if err := db.Save(user).Error; err != nil {
								logTelegram(err.Error())
							}

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
										messageTelegram(fmt.Sprintf(tr(user.TelegramID, "tokenSentR"), rUser.TelegramUsername), int64(rUser.TelegramID))
									}
								}
							}

							messageTelegram(fmt.Sprintf(tr(user.TelegramID, "tokenSent"), tu.Message.From.Username), int64(tu.Message.Chat.ID))
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
	var link string
	var cycle string

	if user.MiningActivated != nil && user.Mining {
		var timeSince float64
		mined := user.MinedAnotes
		if user.LastStatus != nil {
			timeSince = time.Since(*user.LastStatus).Hours()
		} else {
			timeSince = float64(0)
		}
		if timeSince > float64(24) {
			timeSince = float64(24)
		}
		mined += int((timeSince * user.miningPower()) * float64(satInBtc))
		user.MinedAnotes = mined
		now := time.Now()
		user.LastStatus = &now
		if err := db.Save(user).Error; err != nil {
			logTelegram(err.Error())
		}
	}

	status := user.status()
	mining := user.isMiningStr()
	power := user.miningPowerStr()
	team := user.team()
	teamInactive := user.teamInactive()
	mined := float64(user.MinedAnotes) / float64(satInBtc)
	if user.MiningActivated != nil {
		sinceMine := time.Since(*user.MiningActivated)
		sinceHour := 23 - int(sinceMine.Hours())
		sinceMin := 0
		sinceSec := 0
		if sinceHour < 0 {
			sinceHour = 0
		} else {
			sinceMin = 59 - (int(sinceMine.Minutes()) - (int(sinceMine.Hours()) * 60))
			sinceSec = 59 - (int(sinceMine.Seconds()) - (int(sinceMine.Minutes()) * 60))
		}
		cycle = fmt.Sprintf("%.2d:%.2d:%.2d", sinceHour, sinceMin, sinceSec)
	} else {
		cycle = "00:00:00"
	}

	if len(user.Address) == 0 {
		link = tr(user.TelegramID, "regRequired")
	} else {
		link = ""
	}

	msg := fmt.Sprintf("⭕️  <strong><u>"+tr(user.TelegramID, "statusTitle")+"</u></strong>\n\n"+
		"<strong>Status:</strong> %s\n"+
		"<strong>"+tr(user.TelegramID, "statusAddress")+":</strong> %s\n"+
		"<strong>Mining:</strong> %s\n"+
		"<strong>"+tr(user.TelegramID, "statusPower")+":</strong> %s\n"+
		"<strong>"+tr(user.TelegramID, "statusTeam")+":</strong> %d\n"+
		"<strong>"+tr(user.TelegramID, "statusInactive")+":</strong> %d\n"+
		"<strong>"+tr(user.TelegramID, "mined")+":</strong> %.8f\n"+
		"<strong>"+tr(user.TelegramID, "miningCycle")+":</strong> %s\n"+
		"<strong>Referral Link: %s</strong>",
		status, user.Address, mining, power, team, teamInactive, mined, cycle, link)

	messageTelegram(msg, int64(tu.Message.Chat.ID))

	if len(user.Address) > 0 {
		msg := fmt.Sprintf("https://www.anonutopia.com/anote?r=%s", user.Address)
		messageTelegram(msg, int64(tu.Message.Chat.ID))
	}
}

func mineCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)

	ksmc := &KeyValue{Key: "miningCode"}
	db.FirstOrCreate(ksmc, ksmc)

	msgArr := strings.Fields(tu.Message.Text)
	if len(msgArr) == 1 && strings.HasPrefix(tu.Message.Text, "/mine") {
		msg := tgbotapi.NewMessage(int64(tu.Message.Chat.ID), tr(user.TelegramID, "dailyCode"))
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: false}
		msg.ReplyToMessageID = tu.Message.MessageID
		bot.Send(msg)
	} else if msgArr[1] == strconv.Itoa(int(ksmc.ValueInt)) {
		var timeSince float64
		mined := user.MinedAnotes
		if user.LastStatus != nil {
			timeSince = time.Since(*user.LastStatus).Hours()
		} else {
			timeSince = float64(0)
		}
		if timeSince > float64(24) {
			timeSince = float64(24)
		}
		mined += int((timeSince * user.miningPower()) * float64(satInBtc))
		user.MinedAnotes = mined
		now := time.Now()
		user.MiningActivated = &now
		user.LastStatus = &now
		user.Mining = true
		user.SentWarning = false
		if err := db.Save(user).Error; err != nil {
			logTelegram(err.Error())
		}
		messageTelegram(tr(user.TelegramID, "startedMining"), int64(tu.Message.Chat.ID))
	} else {
		messageTelegram(tr(user.TelegramID, "codeNotValid"), int64(tu.Message.Chat.ID))
	}
}

func withdrawCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)

	if user.LastWithdraw != nil && time.Since(*user.LastWithdraw).Hours() < float64(24) {
		messageTelegram(tr(user.TelegramID, "withdrawTimeLimit"), int64(tu.Message.Chat.ID))
	} else if user.MinedAnotes == 0 {
		messageTelegram(tr(user.TelegramID, "withdrawNoAnotes"), int64(tu.Message.Chat.ID))
	} else if len(user.Address) == 0 {
		messageTelegram(tr(user.TelegramID, "notRegistered"), int64(tu.Message.Chat.ID))
	} else {
		var timeSince float64
		mined := user.MinedAnotes
		if user.LastStatus != nil {
			timeSince = time.Since(*user.LastStatus).Hours()
		} else {
			timeSince = float64(0)
		}
		if timeSince > float64(24) {
			timeSince = float64(24)
		}
		mined += int((timeSince * user.miningPower()) * float64(satInBtc))
		user.MinedAnotes = mined
		if err := db.Save(user).Error; err != nil {
			logTelegram(err.Error())
		}

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
			if err := db.Save(user).Error; err != nil {
				logTelegram(err.Error())
			}
			messageTelegram(tr(user.TelegramID, "sentAnotes"), int64(tu.Message.Chat.ID))
		}

	}
}

func unknownCommand(tu TelegramUpdate) {
	user := &User{TelegramID: tu.Message.From.ID}
	db.First(user, user)
	messageTelegram(tr(user.TelegramID, "commandNotAvailable"), int64(tu.Message.Chat.ID))
}
