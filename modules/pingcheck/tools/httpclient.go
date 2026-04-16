package tools

import (
	"bytes"
	"fmt"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"encoding/json"
	"strings"
	"time"
	"errors"
	"context"
)

type CMDBResp struct {
	Code    int    `json:"code"`
	Msg 	string  `json:"message"`
	SUCCESS bool 	`json:"success"`
}

const (
	DefaultHttpTimeout = 180 * time.Second
	TestHttpTimeout    = 300 * time.Second
)


func FalconToken() (string) {

        // crate falcon api header token access

        token, err := json.Marshal(map[string]string{"name": g.Config().Falcon.FalconAuthName,
         "sig": g.Config().Falcon.FalconAuthSig})

        if err != nil {
                log.Println(err)
        }

        return  string(token)
}


func HttpApiPut(fullApiUrl string, jsonData []byte, tokenType string) (status bool, err error) {

	client := &http.Client{
		Timeout: DefaultHttpTimeout,
	}

	req, err := http.NewRequest(http.MethodPut, fullApiUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		g.Logger.Errorf("HttpApiPut()  NewRequest() error:%s", err)
		return false, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(req)

	if err != nil {
		g.Logger.Errorf("HttpApiPut() Do() error: %s", err)
		return false, err
	}

	defer resp.Body.Close()

	if  ( resp.StatusCode  == 200 ) {
		return true, nil
	} else {
		resp.Body.Close()
		return false, errors.New("[ERROR] HttpApiPut() response not 200")
	}

}

func HttpApiGet(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {

	cmdbUrl :=  g.Config().Cmdb.Url
	var timeout time.Duration

	if strings.Contains(cmdbUrl, "test") {
		timeout = TestHttpTimeout
	} else {
		timeout = DefaultHttpTimeout
	}

	client := &http.Client{
		Timeout: timeout,
	}

	httpUrl := fullApiUrl + params

	req, err := http.NewRequest("GET", httpUrl, nil)

	if err != nil {
		msg := fmt.Sprintf("HttpApiGet()  NewRequest() error: %s", err)
		g.Logger.Errorf(msg)
		return nil, errors.New(msg)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(req)

	if err != nil {
		msg := fmt.Sprintf("HttpApiGet() Do %s error: %s ", fullApiUrl, err)
		g.Logger.Errorf(msg)
		return nil, errors.New(msg)
	}

	if  resp.StatusCode == 200  {
		return resp.Body, nil
	} else {
		if g.Config().Debug == true {

			b, err := httputil.DumpResponse(resp, true)
			if err != nil {
				g.Logger.Errorf("HttpApiGet() dump error with %s", err)
			}

			g.Logger.Errorf("HttpApiGet() code not 200 %s", string(b))
		}

		resp.Body.Close()
		g.Logger.Errorf("HttpApiGet() resp status error, code:%d ", resp.StatusCode)
		return nil, errors.New("HttpApiGet() resp: code not 200")
	}

}

func HttpApiPost(fullApiUrl string, params []byte, tokenType string) (io.ReadCloser, error) {
	// use to access http post
	// params = post params  [must be []byte format]
	// return http response


	tr := &http.Transport{
		MaxIdleConns: 10,
		IdleConnTimeout: 10 * time.Second,
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   DefaultHttpTimeout,
	}

	req, err := http.NewRequest("POST", fullApiUrl, bytes.NewBuffer(params))

	if err != nil {
		g.Logger.Errorf("HttpApiPost() NewRequest error:%s", err)
		return nil, errors.New("HttpApiPost() http post error with NewRequest")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultHttpTimeout)
	defer cancel()
	request := req.WithContext(ctx)

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(request)

	if err != nil {
		resp.Body.Close()
		g.Logger.Errorf("HttpApiPost() client access error: %s", err)
		return nil, errors.New("HttpApiPost() client access error.")
	}


	if ( resp.StatusCode  == 200 ) {
		return resp.Body, nil
	} else {
		if g.Config().Debug == true {

			b, err := httputil.DumpResponse(resp, true)
			if err != nil {
				g.Logger.Errorf("HttpApiPost() dump with %s", err)
			}

			g.Logger.Errorf("HttpApiPost() ", string(b))
		}
		resp.Body.Close()
		return nil, errors.New("HttpApiPost() resp status code not 200.")
	}
}

func HttpApiDelete(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {
	// use to do http Delete request
	// METHOD: DELETE

	client := &http.Client{
		Timeout: DefaultHttpTimeout,
	}
	httpUrl := fullApiUrl + params
	req, err := http.NewRequest("DELETE", httpUrl, nil)

	if err != nil {
		g.Logger.Errorf("HttpApiDelete() NewRequest() error:%s", err)
		return nil, errors.New("HttpApiDelete() http delete error with NewRequest")
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(req)

	if err != nil {
		g.Logger.Errorf("HttpApiDelete() Do() error:%s", err)
		return nil, errors.New("HttpApiDelete() http delete error")
	}

	if ( resp.StatusCode  == 200 ) {
		return resp.Body, nil
	} else {
		resp.Body.Close()
		g.Logger.Errorf("HttpApiDelete() resp.StatusCode code is:%d", resp.StatusCode)
		return nil, errors.New("HttpApiDelete() resp status code not 200.")
	}
}