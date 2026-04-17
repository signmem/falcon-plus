package http


import (
	"io"
	"net/http"
	"errors"
	"github.com/signmem/falcon-plus/modules/status-collectd/g"
	"fmt"
)

func HttpApiGet(fullApiUrl string, params string) (io.ReadCloser, error) {

	client := &http.Client{}
	httpUrl := fullApiUrl + params

	req, err := http.NewRequest("GET", httpUrl, nil)

	if err != nil {
		msg := fmt.Sprintf("HttpApiGet()  NewRequest() url: %s error: %s", httpUrl, err)
		g.Logger.Errorf(msg)
		return nil, errors.New(msg)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)

	if err != nil {
		msg := fmt.Sprintf("HttpApiGet() Do %s error:%s ", fullApiUrl, err)
		g.Logger.Errorf(msg)
		return nil, errors.New(msg)
	}

	if ( resp.StatusCode  == 200 ) {
		return resp.Body, nil
	} else {
		g.Logger.Errorf("HttpApiGet() resp status error, code:%d ", resp.StatusCode)
		return nil, errors.New("HttpApiGet() resp status code not 200.")
	}

}
