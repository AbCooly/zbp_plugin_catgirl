package catgirl

import (
	"bytes"
	"encoding/json"
	"io"
)

type DataRes struct {
	Code    int     `json:"code"`
	Data    AllData `json:"data"`
	Success bool    `json:"success"`
}

type AllData struct {
	Id           int     `json:"id"`
	PaintingSign string  `json:"paintingSign"`
	State        string  `json:"state"`
	ImageUrl     string  `json:"imageUrl"`
	Progress     float64 `json:"progress"`
}

type DrawReq struct {
	Prompt                 string      `json:"prompt"`
	NegativePrompt         string      `json:"negativePrompt"`
	AddModelNegativePrompt bool        `json:"addModelNegativePrompt"`
	Mid                    int         `json:"mid"`
	Tags                   string      `json:"tags"`
	Shape                  int         `json:"shape"`
	Img                    string      `json:"img"`
	ImgRatio               int         `json:"imgRatio"`
	Seed                   int         `json:"seed"`
	StableSeed             bool        `json:"stableSeed"`
	DetailsLevel           int         `json:"detailsLevel"`
	FanFictionId           interface{} `json:"fanFictionId"`
}

func setDrawData(prompt string) (res io.Reader) {
	dq := DrawReq{
		Prompt:                 prompt,
		NegativePrompt:         "",
		AddModelNegativePrompt: true,
		Mid:                    1,
		Tags:                   "",
		Shape:                  1,
		FanFictionId:           nil,
		Img:                    "",
		ImgRatio:               2,
		Seed:                   3805737033,
		StableSeed:             false,
		DetailsLevel:           9,
	}
	mas, err := json.Marshal(&dq)
	if err != nil {
		return
	}
	res = bytes.NewReader(mas)
	return
}

type ProcessReq struct {
	TaskId string `json:"taskId"`
}

func setProcessData(taskId string) (res io.Reader) {
	pr := ProcessReq{
		TaskId: taskId,
	}
	mas, err := json.Marshal(&pr)
	if err != nil {
		return
	}
	res = bytes.NewReader(mas)
	return
}
