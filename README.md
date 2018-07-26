# go.kit

 					**kit微服务框架中文详解**
1.端点
	

    ·定义业务逻辑的服务模型接口
    type Service interface{
    	GetAge(string) int error
    	Save(interface{}) int error
    }
	·创建一个实现服务模型的结构体,并调用服务模型声明的接口
	type service struct{}
	func (s service) GetAge(s string) (int error){
		if s==" "{
			return nil,errors.New("不能传入空姓名")
		} 
		//业务逻辑处理
		a:=12
		return a,nil
	}
	·定义请求和响应数据的结构体
	  //定义获取用户年龄请求字段
	type User struct{

	    Name string `json:"name"`

	}

	·

2.传输层

3.请求跟踪

4.服务监控

5.服务发现