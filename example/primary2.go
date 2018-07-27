package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/go-kit/kit/endpoint"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//声明服务模型
type Service interface {
	GetAge(context.Context, string) (int, error)
	// Save(context.Context, UserRequest) error
}

//创建使用服务的主体
type service struct{}

func (service) GetAge(_ context.Context, name string) (int, error) {

	if name == " " {

		return 0, errors.New("对不起,不能传入空姓名\n")

	}

	return 23, nil

}

// func (s service) Save(user UserRequest) error {

// 	if user.Age < 18 || user.Name == " " {

// 		return errors.New("对不起,您年龄还不满18岁")

// 	}

// 	return nil

// }

//user字段
type User struct {
	Name string
	Age  int
}

//根据用户名获取年龄响应
type ageRequest struct {
	Name string `json:"name"`
}

//响应年龄
type ageResponse struct {
	Age int `json:"age"`
}

//端点创建

func makeGetAgepoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ageRequest)
		v, err := s.GetAge(ctx, req.Name)
		if err == nil {
			return ageResponse{v}, nil
		}
		return nil, err
	}
}

//创建一个日志记录中间件
// func loggingMiddleware(logger log.Logger) endpoint.Middleware {

// 	return func(next endpoint.Endpoint) endpoint.Endpoint {
// 		return func(ctx context.Context, request interface{}) (interface{}, error) {
// 			logger.Log("msg", "calling endpoint")
// 			defer logger.Log("msg", "callend endpoing")
// 			return next(ctx, request)
// 		}
// 	}
// }

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) GetAge(ctx context.Context, s string) (output int, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "age",
			"input", s,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	fmt.Println(s)
	output, err = mw.next.GetAge(ctx, s) //触发服务主体函数，获取输出值
	return
}

//应用监控
type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	// countResult    metrics.Histogram
	next Service
}

func (mw instrumentingMiddleware) GetAge(ctx context.Context, s string) (output int, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetAge", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)                                 //统计请求次数
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds()) //统计响应时间
	}(time.Now())
	output, err = mw.next.GetAge(ctx, s)
	return
}

func main() {

	logger := log.NewLogfmtLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
	//请求次数统计
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "sms_send",
		Name:      "send",
		Help:      "请求次数统计",
	}, fieldKeys)
	//请求延迟统计
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "sms",
		Subsystem: "sms_time",
		Name:      "time",
		Help:      "延迟统计",
	}, fieldKeys)

	var s Service
	s = service{}
	s = loggingMiddleware{logger, s}
	s = instrumentingMiddleware{requestCount, requestLatency, s}

	getAgeHandler := httptransport.NewServer(
		makeGetAgepoint(s),
		decodeAgeRequest,
		encodeAgeResponse,
	)
	http.Handle("/age", getAgeHandler)
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":9091", nil)
}

//针对请求的json数据解码
func decodeAgeRequest(_ context.Context, r *http.Request) (interface{}, error) {

	var req ageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return nil, err
	}
	return req, nil

}
func encodeAgeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {

	return json.NewEncoder(w).Encode(response)

}
