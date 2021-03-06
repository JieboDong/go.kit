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

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
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

func main() {
	logger := log.NewLogfmtLogger(os.Stderr) //定义日志输出类型 标准错误输出
	// mid := loggingMiddleware(log.With(logger, "method", "age"))
	var s Service
	s = service{}
	s = loggingMiddleware{logger, s}
	fmt.Printf("%v", s)
	getAgeHandler := httptransport.NewServer(
		makeGetAgepoint(s),
		decodeAgeRequest,
		encodeAgeResponse,
	)
	http.Handle("/age", getAgeHandler)
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
