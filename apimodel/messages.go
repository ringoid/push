package apimodel

const (
	NewPeopleMessageText_ru = `Появились новые люди...`
	NewPeopleMessageText_en = `Check out new users...`

	NewLmmDataMessageText_ru = `Для тебя что-то есть...`
	NewLmmDataMessageText_en = `You have something new...`
)

var NewPeopleMessageTexts map[string]string
var NewLmmDataMessageTexts map[string]string

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
}
