package http

import (
        "bytes"
        "context"
        "crypto/tls"
        "crypto/x509"
        "encoding/json"
        "errors"
        "io"
        "io/ioutil"
        "net/http"
        "time"
	"fmt"
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

        ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
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

        defer resp.Body.Close()

        if  ( resp.StatusCode  == 200 ) {
                return true, nil
        } else {
                return false, errors.New("ttpApiPut() response not 200")
        }
}

func HttpApiGet(fullApiUrl string, params string, tokenType string) ([]byte, error) {

        var httpUrl string

        if params != ""  {
                httpUrl = fullApiUrl + params
        } else {
                httpUrl = fullApiUrl
        }

        ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
        defer cancel()

        req, err := http.NewRequestWithContext(ctx, "GET", httpUrl, nil)

        if err != nil {
		msg := fmt.Sprintf("HttpApiGet() http get error with NewRequest url: %s", httpUrl)
                return nil, errors.New(msg)
        }

        req.Header.Set("Content-Type", "application/json; charset=utf-8")
        if tokenType == "falcon" {
                token, err := FalconToken()
                if err == nil {
                        req.Header.Add("Apitoken", token)
                }
        }
        client := &http.Client{
                Transport: &http.Transport {
                        DisableKeepAlives: true,
                },
        }

        resp, err := client.Do(req)

        if err != nil {
		msg := fmt.Sprintf("HttpApiGet() http do get %s error", httpUrl)
                return nil, errors.New(msg)
        }

        if resp.Body == nil {
                return nil, errors.New("HttpApiGet() resp body is nil.")
        }

        defer resp.Body.Close()

        if ( resp.StatusCode  != 200 ) {
                return nil, errors.New("HttpApiGet() resp status code not 200.")
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                if errors.Is(err, context.Canceled) {
                        return nil, errors.New("download timed out after 60 seconds")
                }
        }
        return body, nil
}


func HttpApiPost(fullApiUrl string, params []byte, tokenType string) ([]byte, error) {
        // use to access http post
        // params = post params  [must be []byte format]
        // return http response


        ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
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

        defer req.Body.Close()

        if resp.Body == nil {
                return nil, err
        }

        defer resp.Body.Close()

        if ( resp.StatusCode  != 200 ) {
                return nil, errors.New("HttpApiPost() resp status code not 200.")
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                if errors.Is(err, context.Canceled) {
                        return nil, errors.New("download timed out after 60 seconds")
                }
        }

        return body, nil

}

func HttpApiDelete(fullApiUrl string, params string, tokenType string) ([]byte, error) {
        // use to do http Delete request
        // METHOD: DELETE

        var httpUrl string

        if params != ""  {
                httpUrl = fullApiUrl + params
        } else {
                httpUrl = fullApiUrl
        }

        ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
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
        if err != nil {
                return nil, errors.New("HttpApiDelete() http delete error")
        }
        defer resp.Body.Close()
        if ( resp.StatusCode  != 200 ) {
                return nil, errors.New("HttpApiDelete() resp status code not 200.")
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                if errors.Is(err, context.Canceled) {
                        return nil, errors.New("download timed out after 60 seconds")
                }
        }
        return body, nil

}

type SSLCert struct {
        CaFile          string          `json:"cafile"`
        CertFile        string          `json:"certfile"`
        KeyFile         string          `json:"keyfile"`
}

func HttpsApiGet(fullApiUrl string, params string, sslfile SSLCert) ([]byte, error) {

	// 加载CA证书
	caCert, err := ioutil.ReadFile(sslfile.CaFile)
	if err != nil {
		return nil, errors.New("failed to read CA file: " + err.Error())
	}

        caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to parse CA certificate")
	}

	// 加载客户端证书
	cert, err := tls.LoadX509KeyPair(sslfile.CertFile, sslfile.KeyFile)
	if err != nil {
		return nil, errors.New("failed to load key pair: " + err.Error())
	}


        client := &http.Client {
		Timeout: 60 * time.Second,
                Transport: &http.Transport {
                        TLSClientConfig: &tls.Config {
                                RootCAs: caCertPool,
                                Certificates: []tls.Certificate{cert},
                        },
                },
        }

        var httpUrl string

        if params != ""  {
                httpUrl = fullApiUrl + params
        } else {
                httpUrl = fullApiUrl
        }

        resp, err := client.Get(httpUrl)

        if err != nil {
                return nil, errors.New("HttpsApiGet() https " + httpUrl +
                        " get error with NewRequest: " + err.Error())
        }

        defer resp.Body.Close()

        if  resp.StatusCode != http.StatusOK { 
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body: " + err.Error())
	}

	return data, nil
}
