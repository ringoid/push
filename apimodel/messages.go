package apimodel

const (
	NewPeopleMessageText_ru = `Появились новые люди   ✨`
	NewPeopleMessageText_en = `Check out new users   ✨`

	NewLmmDataMessageText_ru = `Есть новый лайк, взаимная симпатия или сообщение   💖`
	NewLmmDataMessageText_en = `You have new like, match or message   💖`

	NewLikePushMessageText_ru = `Ты кому-то нравишься!   💓`
	NewLikePushMessageText_en = `Someone has liked you!   💓`

	NewMatchPushMessageText_ru = `Тебе ответили взаимностью!   😍😍😍`
	NewMatchPushMessageText_en = `Someone has liked you back (it is a match!)   😍😍😍`

	NewMessagePushMessageText_ru = `Тебе прислали сообщение...   💌`
	NewMessagePushMessageText_en = `Someone has sent you a message...   💌`
)

var NewPeopleMessageTexts map[string]string
var NewLmmDataMessageTexts map[string]string

var NewLikeMessageTexts map[string]string
var NewMatchMessageTexts map[string]string
var NewMessageMessageTexts map[string]string

func init() {
	NewPeopleMessageTexts = make(map[string]string)
	NewPeopleMessageTexts["ru"] = NewPeopleMessageText_ru
	NewPeopleMessageTexts["be"] = NewPeopleMessageText_ru
	NewPeopleMessageTexts["ua"] = NewPeopleMessageText_ru

	NewPeopleMessageTexts["en"] = NewPeopleMessageText_en
	NewPeopleMessageTexts["uk"] = NewPeopleMessageText_en

	NewLmmDataMessageTexts = make(map[string]string)
	NewLmmDataMessageTexts["ru"] = NewLmmDataMessageText_ru
	NewLmmDataMessageTexts["be"] = NewLmmDataMessageText_ru
	NewLmmDataMessageTexts["ua"] = NewLmmDataMessageText_ru

	NewLmmDataMessageTexts["en"] = NewLmmDataMessageText_en
	NewLmmDataMessageTexts["uk"] = NewLmmDataMessageText_en

	NewLikeMessageTexts = make(map[string]string)
	NewLikeMessageTexts["ru"] = NewLikePushMessageText_ru
	NewLikeMessageTexts["be"] = NewLikePushMessageText_ru
	NewLikeMessageTexts["ua"] = NewLikePushMessageText_ru

	NewLikeMessageTexts["en"] = NewLikePushMessageText_en
	NewLikeMessageTexts["uk"] = NewLikePushMessageText_en

	NewMatchMessageTexts = make(map[string]string)
	NewMatchMessageTexts["ru"] = NewMatchPushMessageText_ru
	NewMatchMessageTexts["be"] = NewMatchPushMessageText_ru
	NewMatchMessageTexts["ua"] = NewMatchPushMessageText_ru

	NewMatchMessageTexts["en"] = NewMatchPushMessageText_en
	NewMatchMessageTexts["uk"] = NewMatchPushMessageText_en

	NewMessageMessageTexts = make(map[string]string)
	NewMessageMessageTexts["ru"] = NewMessagePushMessageText_ru
	NewMessageMessageTexts["be"] = NewMessagePushMessageText_ru
	NewMessageMessageTexts["ua"] = NewMessagePushMessageText_ru

	NewMessageMessageTexts["en"] = NewMessagePushMessageText_en
	NewMessageMessageTexts["uk"] = NewMessagePushMessageText_en

}
