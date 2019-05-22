package balenaapi

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/levigross/grequests"
)

var BALENA_API_BASE_URL = "https://api.balena-cloud.com"

type BalenaDeviceTag struct {
	Id     int    `json:"id"`
	TagKey string `json:"tag_key"`
	Value  string `json:"value"`
}

type BalenaDeviceTagResponse struct {
	Data []BalenaDeviceTag `json:"d"`
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func GetTagValue(apiKey string, uuid string, tagKey string) (string, error) {
	options := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", apiKey),
		},
		RequestTimeout: time.Duration(5) * time.Second,
	}
	filters, err := UrlEncoded(fmt.Sprintf("$filter=device/uuid eq '%s'&$filter=tag_key eq '%s'", uuid, tagKey))
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v4/device_tag?%s", BALENA_API_BASE_URL, filters)
	resp, err := grequests.Get(url, options)
	if err != nil {
		return "", err
	}
	if !resp.Ok {
		message := fmt.Sprintf("Request failed with response status code: %d", resp.StatusCode)
		return "", errors.New(message)
	}

	tags := new(BalenaDeviceTagResponse)
	err = resp.JSON(tags)
	if err != nil {
		return "", err
	}

	if len(tags.Data) == 0 {
		return "", nil
	} else {
		window := tags.Data[0].Value
		return window, nil
	}
}
