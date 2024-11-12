package jobs

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goravel/framework/facades"
)

type TestJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestJob) Signature() string {
	return "test_job"
}

// Handle Execute the job.
func (receiver *TestJob) Handle(args ...any) error {
	var count int = 0
	for {
		time.Sleep(2 * time.Second)
		resp, err := http.Get("https://www.baidu.com/")
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body) // 读取响应 body, 返回为 []byte
		facades.Log().Infof("请求结果:%v\n", string(body))
		count += 1
		if count > 10 {
			break
		}
	}
	facades.Log().Infof("队列退出--end")
	return nil
}
