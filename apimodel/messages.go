package apimodel

const (
	GeneralPushMessageText_ru = `Для тебя что-то есть...`
	GeneralPushMessageText_en = `You have something new...`
)

var MessageTexts map[string]string

func init() {
	MessageTexts = make(map[string]string)
	MessageTexts["ru"] = GeneralPushMessageText_ru
	MessageTexts["be"] = GeneralPushMessageText_ru
	MessageTexts["ua"] = GeneralPushMessageText_ru

	MessageTexts["en"] = GeneralPushMessageText_en
	MessageTexts["uk"] = GeneralPushMessageText_en
}
