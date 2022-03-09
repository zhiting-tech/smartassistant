package cloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
)

const (
	NoSoftwareRecordStatus = 105010
	NoFirmwareRecordStatus = 105008
)

var (
	RequestSoftwareLastVersionErr  = fmt.Errorf("request software last version error")
	NoSoftwareLastVersionRecordErr = fmt.Errorf("no software record")

	RequestFirmwareLastVersionErr  = fmt.Errorf("request firmware last version error")
	NoFirmwareLastVersionRecordErr = fmt.Errorf("no firmware record")
)

type SoftwareLastVersionHttpResult struct {
	Status int                       `json:"status"`
	Reason string                    `json:"reason"`
	Data   SoftwareLastVersionResult `json:"data"`
}

type SoftwareLastVersionResult struct {
	Name     string                                `json:"name"`
	Version  string                                `json:"version"`
	Remark   string                                `json:"remark"`
	UpdateAt uint64                                `json:"update_at"`
	Services []SoftwareLastVersionSubServiceResult `json:"services"`
}

type SoftwareLastVersionSubServiceResult struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Image   string `json:"image"`
}

func GetLastSoftwareVersion() (result *SoftwareLastVersionHttpResult, err error) {
	var (
		req  *http.Request
		resp *http.Response
		data []byte
	)
	url := fmt.Sprintf("%s/common/service/software/lastest", config.GetConf().SmartCloud.URL())
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}

	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = RequestSoftwareLastVersionErr
		return
	}

	defer func() {
		resp.Body.Close()
	}()

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	result = &SoftwareLastVersionHttpResult{}
	if err = json.Unmarshal(data, result); err != nil {
		return
	}

	if result.Status != 0 {
		if result.Status == NoSoftwareRecordStatus {
			err = NoSoftwareLastVersionRecordErr
		} else {
			err = fmt.Errorf(result.Reason)
		}

		return
	}

	return
}

type FirmwareLastVersionHttpResult struct {
	Status int                       `json:"status"`
	Reason string                    `json:"reason"`
	Data   FirmwareLastVersionResult `json:"data"`
}

type FirmwareLastVersionResult struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Remark    string `json:"remark"`
	UpdateAt  uint64 `json:"update_at"`
	FileName  string `json:"file_name"`
	FileUrl   string `json:"file_url"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
}

func GetLastFirmwareVersion() (result *FirmwareLastVersionHttpResult, err error) {
	var (
		req  *http.Request
		resp *http.Response
		data []byte
	)
	url := fmt.Sprintf("%s/common/service/firmware/lastest", config.GetConf().SmartCloud.URL())
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}

	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = RequestSoftwareLastVersionErr
		return
	}

	defer func() {
		resp.Body.Close()
	}()

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	result = &FirmwareLastVersionHttpResult{}
	if err = json.Unmarshal(data, result); err != nil {
		return
	}

	if result.Status != 0 {
		if result.Status == NoFirmwareRecordStatus {
			err = NoFirmwareLastVersionRecordErr
		} else {
			err = fmt.Errorf(result.Reason)
		}

		return
	}

	return
}
