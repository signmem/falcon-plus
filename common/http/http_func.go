package http

import (
        "bytes"
        "context"
        "errors"
        "io"
        "net/http"
        "time"
        "encoding/json"
)

var (
        FalconAuthName = "falcon_api"
        FalconAuthSig = "b5bd81ef572011e79f8d48fd8e3b7eb0"
)

func FalconToken() (string, error) {

        // crate falcon api header token access

        token, err := json.Marshal(map[string]string{"name": FalconAuthName,
        "sig": FalconAuthSig})

        if err != nil {
                return "", err
        }

        return  string(token), nil
}

func HttpApiPut(fullApiUrl string, jsonData []byte, tokenType string) (status bool, err error) {

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        req, err := http.NewRequest(http.MethodPut, fullApiUrl, bytes.NewBuffer(jsonData))

        if err != nil {
                return false, err
        }

        req = req.WithContext(ctx)

        req.Header.Add("Content-Type", "application/json; charset=utf-8")

        if tokenType == "falcon" {
                token, err := FalconToken()
                if err == nil {
                        req.Header.Add("Apitoken", token)
                }
        }

        client := &http.Client{}
        resp, err := client.Do(req)

        if err != nil {
                return false, err
        }

        if  ( resp.StatusCode  == 200 ) {
                return true, nil
        } else {
                return false, errors.New("ttpApiPut() response not 200")
        }
}

func HttpApiGet(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {

        var httpUrl string

        if params != ""  {
                httpUrl = fullApiUrl + params
        } else {
                httpUrl = fullApiUrl
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        req, err := http.NewRequest("GET", httpUrl, nil)

        if err != nil {
                return nil, errors.New("HttpApiGet() http get error with NewRequest")
        }
        req = req.WithContext(ctx)

        req.Header.Add("Content-Type", "application/json; charset=utf-8")
        if tokenType == "falcon" {
                token, err := FalconToken()
                if err == nil {
                        req.Header.Add("Apitoken", token)
                }
        }
        client := &http.Client{}
        resp, err := client.Do(req)

        if err != nil {
                return nil, errors.New("HttpApiGet() http get error")
        }

        if resp.Body == nil {
                return nil, errors.New("HttpApiGet() resp body is nil.")
        }

        if ( resp.StatusCode  == 200 ) {
                return resp.Body, nil
        } else {
                return nil, errors.New("HttpApiGet() resp status code not 200.")
        }

}

func HttpApiPost(fullApiUrl string, params []byte, tokenType string) (io.ReadCloser, error) {
        // use to access http post
        // params = post params  [must be []byte format]
        // return http response


        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        req, err := http.NewRequest("POST", fullApiUrl, bytes.NewBuffer(params))
        if err != nil {
                return nil, errors.New("HttpApiPost() http post error with NewRequest")
        }

        req.Header.Set("Content-Type", "application/json")

        if tokenType == "falcon" {
                token, err := FalconToken()
                if err == nil {
                        req.Header.Add("Apitoken", token)
                }
        }

        req = req.WithContext(ctx)
        client := &http.Client{}

        resp, err := client.Do(req)

        if err != nil {
                return nil, errors.New("HttpApiPost()  client access error.")
        }

        if resp.Body == nil {
                return nil, err
        }

        defer resp.Body.Close()

        if ( resp.StatusCode  == 200 ) {
                return resp.Body, nil
        } else {
                return nil, errors.New("HttpApiPost() resp status code not 200.")
        }
}

func HttpApiDelete(fullApiUrl string, params string, tokenType string) (io.ReadCloser, error) {
        // use to do http Delete request
        // METHOD: DELETE

        var httpUrl string

        if params != ""  {
                httpUrl = fullApiUrl + params
        } else {
                httpUrl = fullApiUrl
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        req, err := http.NewRequest("DELETE", httpUrl, nil)
        if err != nil {
                return nil, errors.New("HttpApiDelete() http delete error with NewRequest")
        }

        req = req.WithContext(ctx)

        client := &http.Client{}

        req.Header.Add("Content-Type", "application/json; charset=utf-8")
        if tokenType == "falcon" {
                token, err := FalconToken()
                if err == nil {
                        req.Header.Add("Apitoken", token)
                }
        }

        resp, err := client.Do(req)
        defer resp.Body.Close()

        if err != nil {
                return nil, errors.New("HttpApiDelete() http delete error")
        }

        if ( resp.StatusCode  == 200 ) {
                return resp.Body, nil
        } else {
                return nil, errors.New("HttpApiDelete() resp status code not 200.")
        }
}
