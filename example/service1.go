package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

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

func main() {

	s := service{}

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
