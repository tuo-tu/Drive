package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"net/http"
)

type MapServiceBiz struct {
	log *log.Helper
}

func NewMapServiceBiz(logger log.Logger) *MapServiceBiz {
	return &MapServiceBiz{log: log.NewHelper(logger)}
}

// 获取驾驶信息
func (msbiz *MapServiceBiz) GetDriverInfo(origin, destination string) (string, string, error) {
	// 一，请求获取
	key := "2b08113dc921fac3afd0992a2b45862e"
	api := "https://restapi.amap.com/v3/direction/driving"
	parameters := fmt.Sprintf("origin=%s&destination=%s&extensions=base&output=json&key=%s", origin, destination, key)
	url := api + "?" + parameters
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body) // io.Reader
	if err != nil {
		return "", "", err
	}
	//fmt.Println(string(body))
	// 二，解析出来,json
	ddResp := &DirectionDrivingResp{}
	if err := json.Unmarshal(body, ddResp); err != nil {
		return "", "", err
	}

	// 三，判定LSB请求结果
	if ddResp.Status == "0" {
		return "", "", errors.New(ddResp.Info)
	}

	// 四，正确返回，默认使用第一条路线
	path := ddResp.Route.Paths[0]
	return path.Distance, path.Duration, nil
}

type DirectionDrivingResp struct {
	Status   string `json:"status,omitempty"`
	Info     string `json:"info,omitempty"`
	Infocode string `json:"infocode,omitempty"`
	Count    string `json:"count,omitempty"`
	Route    struct {
		Origin      string `json:"origin,omitempty"`
		Destination string `json:"destination,omitempty"`
		Paths       []Path `json:"paths,omitempty"`
	} `json:"route"`
}
type Path struct {
	Distance string `json:"distance,omitempty"`
	Duration string `json:"duration,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}
