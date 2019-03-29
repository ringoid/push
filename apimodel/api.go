package apimodel

import "fmt"

type UpdateTokenRequest struct {
	AccessToken string `json:"accessToken"`
	DeviceToken string `json:"deviceToken"`
}

func (resp UpdateTokenRequest) String() string {
	return fmt.Sprintf("%#v", resp)
}
