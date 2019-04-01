package apimodel

import "fmt"

type UpdateTokenRequest struct {
	AccessToken string `json:"accessToken"`
	DeviceToken string `json:"deviceToken"`
}

func (resp UpdateTokenRequest) String() string {
	return fmt.Sprintf("%#v", resp)
}

const (
	MaxPeriodDefault     = int64(60000)
	OfflinePeriodDefault = int64(600000)
	MinForMenDefault     = int64(1)
	MinForWomenDefault   = int64(1)
	MinH                 = int64(9)
	MaxH                 = int64(23)

	//MaxPeriodDefault     = int64(86400000)
	//OfflinePeriodDefault = int64(14400000)
	//MinForMenDefault     = int64(10)
	//MinForWomenDefault   = int64(25)
	//MinH                 = int64(19)
	//MaxH                 = int64(23)
)
