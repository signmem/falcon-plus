package tools

import (
	"bytes"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"encoding/json"
	"time"
	"errors"
	"context"
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

	client := &http.Client{}

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
		g.Logger.Errorf("HttpApiPut() Do() error:%s", err)
		return false, err
	}

	if  ( resp.StatusCode  == 200 ) {
		return true, nil
	} else {
		return false, errors.New("[ERROR] HttpApiPut() response not 200")
	}

}

func HttpApiGet(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {

	client := &http.Client{}
	httpUrl := fullApiUrl + params

	req, err := http.NewRequest("GET", httpUrl, nil)

	if err != nil {
		g.Logger.Errorf("HttpApiGet()  NewRequest() error:%s", err)
		return nil, errors.New("HttpApiGet() http get error with NewRequest")
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(req)

	if err != nil {
		g.Logger.Errorf("HttpApiGet() Do %s error:%s ", fullApiUrl, err)
		return nil, errors.New("HttpApiGet() http get error")
	}

	if ( resp.StatusCode  == 200 ) {
		return resp.Body, nil
	} else {
		g.Logger.Errorf("HttpApiGet() resp status error, code:%d ", resp.StatusCode)
		return nil, errors.New("HttpApiGet() resp status code not 200.")
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

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", fullApiUrl, bytes.NewBuffer(params))

	if err != nil {
		g.Logger.Errorf("HttpApiPost() NewRequest error:%s", err)
		return nil, errors.New("HttpApiPost() http post error with NewRequest")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	request := req.WithContext(ctx)


	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if tokenType == "falcon" {
		token := FalconToken()
		req.Header.Add("Apitoken", token)
	}

	resp, err := client.Do(request)

	if err != nil {
		return nil, errors.New("HttpApiPost()  client access error.")
	}
	defer cancelFunc()
	if resp.Body == nil {
		defer resp.Body.Close()
		return nil, err
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
		return nil, errors.New("HttpApiPost() resp status code not 200.")
	}
}

func HttpApiDelete(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {
	// use to do http Delete request
	// METHOD: DELETE

	client := &http.Client{}
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
		g.Logger.Errorf("HttpApiDelete() resp.StatusCode code is:%d", resp.StatusCode)
		return nil, errors.New("HttpApiDelete() resp status code not 200.")
	}
}