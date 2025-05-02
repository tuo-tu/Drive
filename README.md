# “滴滴代驾”

## 项目描述

基于 Golang 的 代驾项目，实现了高效的后端服务、实时的司机与乘客匹配以及支付系统的集成。

## Kratos框架简介

框架网站：https://go-kratos.dev/docs/

使用kratos的优点：

1. 可以通过模板快速创建项目
2. 高效处理 Protobuf 文件
3. 丰富的命令行工具
4. 操作简单，开发效率高

**Kratos目录结构详解：**

```go
├── api // 目录维护了微服务使用的proto文件以及根据它们所生成的go文件
├── cmd  // 整个项目启动的入口文件
│   ├── main.go
│   ├── wire.go  // 我们使用wire来维护依赖注入（依赖注入：即对象不再自己创建依赖的对象，而是通过外部的方式将所需要的依赖对象“注入”给它。）
│   └── wire_gen.go
├── configs  //维护一些本地调试用的样例配置文件
│   └── config.yaml
├── internal  // 这个目录包括该服务所有不对外暴露的代码，通常的业务逻辑都在这下面，使用internal避免错误引用
│   ├── biz   // 业务逻辑的组装层，类似 DDD 的 domain 层，data 类似 DDD 的 repo，而 repo 接口在这里定义，使用依赖倒置的原则。biz层用于处理核心业务逻辑。例如在本项目的customer服务中，biz层加了一个customer.go文件，文件中①定义了customer模型；②定义了CustomerBiz{}空结构体，并绑定了获取估价服务GetEstimatePrice； 
│   │   ├── biz.go
│   │   └── greeter.go
│   ├── conf  // 内部使用的config的结构定义，使用proto格式生成。
│   │   ├── conf.pb.go
│   │   └── conf.proto
│   ├── data  // 业务数据访问，包含 cache、db 等封装（在这层直接与数据库交互），实现了 biz 的 repo 接口。我们可能会把 data 与 dao 混淆在一起，data 偏重业务的含义，它所要做的是将领域对象重新拿出来，我们去掉了 DDD 的 infra层。
│   │   ├── data.go
│   │   └── greeter.go
│   ├── server  // http和grpc实例的创建和配置，中间件也放在这里。
│   │   ├── grpc.go
│   │   ├── http.go
│   │   └── server.go
│   └── service  // 实现了 api 定义的服务层，类似 DDD 的 application 层，处理 DTO 到 biz 领域实体的转换(DTO -> DO)，同时协同各类 biz 交互，但是不应处理复杂逻辑
│       ├── greeter.go
│       └── service.go
└── third_party  // api 依赖的第三方proto。（后期用到再补充）
    ├── google
    │   └── api
    │       ├── annotations.proto
    │       ├── http.proto
    │       └── httpbody.proto
    └── validate
        └── validate.proto
```

## 验证码服务

基于**验证码长度**和**验证码类型**生成随机验证码。验证码服务供内部其他服务使用，因此仅提供gRPC接口；

### 创建项目模板

```cmd
kratos new verify-code
```

### 修改配置文件

在**config.yaml**文件中修改:

- **端口配置**：http 和 grpc 服务的端口配置
- **数据库的配置**：主要是database、redis；

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s

```

### 创建proto文件

```shell
kratos proto add api/ verifyCode/ verifyCode.proto
```

### 编辑proto文件

服务接口定义：

```protobuf
syntax = "proto3";

package api.verifyCode;
option go_package = "verifyCode/api/verifyCode;verifyCode";

// 类型常量
enum TYPE {
	DEFAULT = 0;
	DIGIT = 1;
	LETTER = 2;
	MIXED = 3;
};
// 定义 GetVerifyCodeRequest 消息
message GetVerifyCodeRequest {
	//	验证码长度
	uint32 length = 1;
	// 验证码类型
	TYPE type = 2;

}
// 定义 GetVerifyCodeReply 消息
message GetVerifyCodeReply {
	//	生成的验证码
	string code = 1;
}

service VerifyCode {
	rpc GetVerifyCode (GetVerifyCodeRequest) returns (GetVerifyCodeReply);
}
```

- 验证码有三种类型：数字、字母、混合类型；

- GetVerifyCodeRequest：包括验证码**长度、类型**，

- GetVerifyCodeReply：返回一个**验证码字符串**；

### 生成client端相关代码

客户端代码即pb.go和grpc.pb.go，文件说明略。

```cmd
kratos proto client api/verifyCode/verifyCode.proto
```

### 生成server端相关代码

```cmd
kratos proto server api/verifyCode/verifyCode.proto -t internal/service
```

命令生成`verifycode.go`，文件实现proto里面定义的rpc服务接口，即获取验证码GetVerifyCode；

```go
package service

import (
	"context"
	"math/rand"
	"strings"

	pb "verifyCode/api/verifyCode"
)

type VerifyCodeService struct {
	pb.UnimplementedVerifyCodeServer
}

func NewVerifyCodeService() *VerifyCodeService {
	return &VerifyCodeService{}
}

func (s *VerifyCodeService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeRequest) (*pb.GetVerifyCodeReply, error) {
	//log.Info("current verifyCode service Run")
	return &pb.GetVerifyCodeReply{
		Code: RandCode(int(req.Length), req.Type),
	}, nil
}

// RandCode 开放的被调用的方法，用于区分类型
func RandCode(l int, t pb.TYPE) string {
	switch t {
	case pb.TYPE_DEFAULT:
		fallthrough
	case pb.TYPE_DIGIT:
		// idxBits表示使用4位二进制数就可以表示完chars的索引了
		return randCode("0123456789", l, 4)
	case pb.TYPE_LETTER:
		return randCode("abcdefghijklmnopqrstuvwxyz", l, 5)
	case pb.TYPE_MIXED:
		return randCode("0123456789abcdefghijklmnopqrstuvwxyz", l, 6)
	default:
	}
	return ""
}

// 随机数的核心方法（优化实现）
// 一次随机多个随机位，分部分多次使用，
// idxBits表示使用4位二进制数就可以表示完chars的索引了
func randCode(chars string, l, idxBits int) string {
	// 计算有效的二进制数位，基于 chars 的长度
	// 推荐写死，因为chars固定，对应的位数长度也固定
	// 形成掩码，mask
	// 例如，使用低idxBits位：00000000000111111
	idxMask := 1<<idxBits - 1 // 00001000000 - 1 = 00000111111
	// 63 位可以用多少次（每一次的排列表示一个随机字符，所以也表示总共可以生成几个随机数）；
	// 为什么是63而不是64？因为最高位是符号位；
	idxMax := 63 / idxBits
	// 利用string builder构建结果缓冲
	sb := strings.Builder{}
	sb.Grow(l) //提前分配足够的内存
	//result := make([]byte, l)
	// 生成随机字符cache:随机位缓存 ;remain:当前还可以用几次
	for i, cache, remain := 0, rand.Int63(), idxMax; i < l; {
		// 如果使用的剩余次数为0，则重新获取随机
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		// 利用掩码获取cache的低位作为randIndex（索引）
		if randIndex := int(cache & int64(idxMask)); randIndex < len(chars) {
			//result[i] = chars[randIndex]
			sb.WriteByte(chars[randIndex])
			i++
		}
		// 使用下一组随机位。右移会丢掉先前的低位，高位补0
		cache >>= idxBits
		remain--
	}
	// return string(result)
	return sb.String()
}

// 随机的核心方法(简单的实现)
//func randCode(chars string, l int) string {
//	charsLen := len(chars)
//	// 结果
//	result := make([]byte, l)
//	// 根据目标长度，进行循环
//	for i := 0; i < l; i++ {
//		// 核心函数 rand.Intn() 生成[0, n)的整型随机数
//		randIndex := rand.Intn(charsLen)
//		// 字符串的单个字符是uint8类型，即byte类型,因此可以赋值
//		result[i] = chars[randIndex]
//	}
//	return string(result)
//}

```

实现过程如下：

#### 初始化验证码服务结构体

即VerifyCodeService结构体，自动初始化，在这个服务中无需修改；

```go
type VerifyCodeService struct {
	pb.UnimplementedVerifyCodeServer
}
func NewVerifyCodeService() *VerifyCodeService {
	return &VerifyCodeService{}
}
```

#### 实现GetVerifyCode函数

参数是req.Length和req.Type，核心方法是RandCode；

```go
func (s *VerifyCodeService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeRequest) (*pb.GetVerifyCodeReply, error) {
	//log.Info("current verifyCode service Run")
	return &pb.GetVerifyCodeReply{
		Code: RandCode(int(req.Length), req.Type),
	}, nil
}
```

#### RandCode函数实现如下

RandCode通过调用randCode函数，生成不同的字符串

```go
// RandCode 开放的被调用的方法，用于区分类型
func RandCode(l int, t pb.TYPE) string {
	switch t {
	case pb.TYPE_DEFAULT:
		fallthrough
	case pb.TYPE_DIGIT:
		// idxBits表示使用4位二进制数就可以表示完chars的索引了
		return randCode("0123456789", l, 4)
	case pb.TYPE_LETTER:
		return randCode("abcdefghijklmnopqrstuvwxyz", l, 5)
	case pb.TYPE_MIXED:
		return randCode("0123456789abcdefghijklmnopqrstuvwxyz", l, 6)
	default:
	}
	return ""
}
```

#### randCode函数实现过程如下（核心）

随机生成一个64位随机数cache，分部分多次使用cache的随机位，生成一次可以使用多次。即通过位操作来高效地从字符池中随机选择字符。

```go
// idxBits表示使用idxBits位二进制数就可以表示完chars的索引了
func randCode(chars string, l, idxBits int) string {
	// 计算有效的二进制数位，基于 chars 的长度，idxBits推荐写死，因为chars固定，对应的位数长度也固定
	// 1.形成掩码，mask（掩码的基本作用是在对随机数进行位操作时，保留我们需要的部分，丢弃不需要的部分）。例如，使用低idxBits位：00000000000111111；
idxMask := 1<<idxBits - 1 // 00001000000 - 1 = 00000111111
	// 2.计算63 位可以用多少次（每一次的排列表示一个随机字符的索引，所以也表示总共可以生成几个随机数）；
	// 为什么是63而不是64？因为最高位是符号位；
	idxMax := 63 / idxBits
	// 利用string builder构建结果缓冲，高效拼接字符串;
	sb := strings.Builder{}
	sb.Grow(l) // 提前分配好足够的内存，
	// 生成随机字符cache：随机位缓存；remain表示当前还可以用几次
	for i, cache, remain := 0, rand.Int63(), idxMax; i < l; {
		// 3.如果使用的剩余次数为0，则重新生成随机数；
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		// 4. 利用掩码获取cache的低几位（最高位肯定为1）作为chars的索引（randIndex）
		if randIndex := int(cache & int64(idxMask)); randIndex < len(chars) {
			sb.WriteByte(chars[randIndex])
			i++
		}
		// 5. 使用cache下一组随机位。右移会砍掉先前用过的低位，高位补0；
		cache >>= idxBits
		remain--
	}
	return sb.String()
}
```

### 在grpc中注册服务

这意味着验证码服务提供grpc访问方式，可供其他服务调用，在`internal/server/grpc.go` 文件中加入：

```go
package server

import (
	v1 "verifyCode/api/helloworld/v1"
	"verifyCode/api/verifyCode"
	"verifyCode/internal/conf"
	"verifyCode/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, VerifyCodeService *service.VerifyCodeService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGreeterServer(srv, greeter)
	verifyCode.RegisterVerifyCodeServer(srv, VerifyCodeService)
	return srv
}
```

### 服务注册到consul中

在`verifyCode/cmd/verifyCode/main.go`里面；将验证码服务注册到consul服务注册中心，consul是分布式部署，基于 Raft 协议保证一致性，使用docker-compose部署consul；

#### 准备consul配置文件（首次使用需要）

docker-compose.yml文件

```yy
services:
  consul:
    container_name: laomaDJConsul
    image: consul
    ports:
      - "8500:8500"
	command: agent -dev -client=0.0.0.0
```

#### 启动consul服务（首次使用需要）

```cmd
docker-compose up -d
```

#### 在main.go中实现

```go
package main

import (
	"flag"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"os"
	"time"

	"verifyCode/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "VerifyCode"
	// Version is the version of the compiled software.
	Version string = "1.0.0"
	// flagconf is the config flag.
	flagconf string

	//id, _ = os.Hostname()
	// 使用唯一的uuid，作为id
	id = Name + "-" + uuid.NewString()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	// 一，获取consul客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "192.168.43.144:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	// 二，获取consul注册管理器
	reg := consul.New(consulClient)
	// 设置meta属性，设置weight
	mate := map[string]string{
		"weight": "999",
	}
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(mate),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		// 三，创建服务时，指定服务器注册
		kratos.Registrar(reg),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	rand.NewSource(time.Now().UnixNano())
	// start and wait for stop signal
	if err := app.Run(); err != nil {
		//fmt.Println(err)
		panic(err)
	}
}

```

##### 配置verifyCode服务基础信息

在`backend/verifyCode/cmd/verifyCode/main.go`里，包括验证码服务的**name，Version，id**等信息；这些信息会在main()函数里面得到使用；

```go
var (
	Name string = "VerifyCode"
	Version string = "1.0.0"
	flagconf string
	// 使用唯一的uuid，作为服务的id
	id = Name + "-" + uuid.NewString()
)
```

##### 新建consul客户端

在`newApp()`函数里面，使用默认配置

```go
consulConfig := api.DefaultConfig()
consulConfig.Address = "192.168.43.144:8500"
consulClient, err := api.NewClient(consulConfig)
```

##### 新建consul服务注册中心

```go
reg := consul.New(consulClient)
```

##### 设置meta属性

设置服务权重weight

```go
mate := map[string]string{
    "weight": "999",
}
```

##### 创建服务

将服务注册到**consul**中，关键步骤:`kratos.Registrar(reg)`

```go
return kratos.New(
	kratos.ID(id),
	kratos.Name(Name),
	kratos.Version(Version),
	kratos.Metadata(mate),
	kratos.Logger(logger),
	kratos.Server(
		gs,
		hs,
	),
	// 创建服务时，指定服务器注册中心reg；
	kratos.Registrar(reg),
)
```

### 依赖注入

在`internal/service/service.go` 文件中加入：

```go
package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGreeterService, NewVerifyCodeService)
```

然后执行依赖注入命令：

```cmd
go generate ./...
```

注意：第一次使用需要引入依赖注入包

```cmd
go get github.com/google/wire/cmd/wire
```

### 启动服务

```cmd
kratos run
```

## 顾客服务

两大关键步骤，**生成验证码和登录认证**

### 创建顾客服务模板

```cmd
kratos new customer
```

### 修改配置文件

主要是修改http 和 grpc服务端口；

```yaml
server:
  http:
    addr: 0.0.0.0:8100
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9100
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:123456@tcp(192.168.43.144:3306)/laomadj_customer?parseTime=True&loc=Local
  redis:
    addr: 192.168.43.144:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
```

### 创建proto 文件

创建顾客服务相关proto 文件

```cmd
kratos proto add api/customer/customer.proto -- 举例一个，一共需要3个文件哦
```

### 编辑proto文件

- **verifyCode.proto**：包含**获取验证码**（GetVerifyCode）接口。不用生成service代码。

  ```protobuf
  syntax = "proto3";
  
  package api.verifyCode;
  option go_package = "verifyCode/api/verifyCode;verifyCode";
  
  // 类型常量
  enum TYPE {
  	DEFAULT = 0;
  	DIGIT = 1;
  	LETTER = 2;
  	MIXED = 3;
  };
  // 定义 GetVerifyCodeRequest 消息
  message GetVerifyCodeRequest {
  	//	验证码长度
  	uint32 length = 1;
  	// 验证码类型
  	TYPE type = 2;
  
  }
  // 定义 GetVerifyCodeReply 消息
  message GetVerifyCodeReply {
  	//	生成的验证码
  	string code = 1;
  }
  
  service VerifyCode {
  	rpc GetVerifyCode (GetVerifyCodeRequest) returns (GetVerifyCodeReply);
  }
  ```

- **valuation.proto**：包含**费用预估**（EstimatePrice）接口。不用生成service代码

  ```protobuf
  syntax = "proto3";
  
  package api.valuation;
  
  option go_package = "customer/api/valuation;valuation";
  
  service Valuation {
  	rpc GetEstimatePrice (GetEstimatePriceReq) returns (GetEstimatePriceReply);
  }
  
  message GetEstimatePriceReq {
  	string origin = 1;
  	string destination = 2;
  }
  message GetEstimatePriceReply {
  	string origin = 1;
  	string destination = 2;
  	int64 price = 3;
  }
  ```

- **customer.proto**：包括获取验证码、**登录**（Login）、退出登录（Logout）、费用预估；在API的定义中增加option选项，并增加配置http访问方式，这意味着这几个接口服务可以通过HTTP方式进行访问，需要生成service代码，详细如下：

  ```protobuf
  syntax = "proto3";
  
  package api.customer;
  
  // 导入包
  import "google/api/annotations.proto";
  
  option go_package = "customer/api/customer;customer";
  
  service Customer {
  	// 获取验证码
  	rpc GetVerifyCode (GetVerifyCodeReq) returns (GetVerifyCodeResp) {
  		// 这意味着这个RPC方法可以通过HTTP协议进行访问
  		option (google.api.http) = {
  			get: "/customer/get-verify-code/{telephone}"
  		};
  	}
  
  	// 登录
  	rpc Login (LoginReq) returns (LoginResp) {
  		option (google.api.http) = {
  			post: "/customer/login",
  			body: "*",
  		};
  	}
  
  	// 退出登陆
  	rpc Logout (LogoutReq) returns (LogoutResp) {
  		option (google.api.http) = {
  			get: "/customer/logout",
  		};
  	}
  
  	// 价格预估
  	rpc EstimatePrice (EstimatePriceReq) returns (EstimatePriceResp) {
  		option (google.api.http) = {
  			get: "/customer/estimate-price/{origin}/{destination}",
  		};
  	}
  }
  
  // 获取验证码的消息
  message GetVerifyCodeReq {
  	string telephone = 1;
  }
  message GetVerifyCodeResp {
  	int64 code = 1;
  	string message = 2;
  	// 验证码
  	string verify_code = 3;
  	// 生成时间 unix timestamp
  	int64 verify_code_time = 4;
  	// 有效期，单位 second
  	int32 verify_code_life = 5;
  };
  
  // 登录的消息
  message LoginReq {
  	string telephone = 1;
  	string verify_code = 2;
  };
  message LoginResp {
  	int64 code = 1;
  	string message = 2;
  	// token,登录表示，特殊的字符串，JWT 编码格式
  	string token = 3;
  	// 生成时间 unix timestamp
  	int64 token_create_at = 4;
  	// 有效期，单位 second
  	int32 token_life = 5;
  }
  
  message LogoutReq {};
  message LogoutResp {
  	int64 code = 1;
  	string message = 2;
  };
  
  message EstimatePriceReq {
  	string origin = 1;
  	string destination = 2;
  };
  
  message EstimatePriceResp {
  	int64 code = 1;
  	string message = 2;
  	string origin = 3;
  	string destination = 4;
  	int64 price = 5;
  };
  ```

### 生成client端相关代码

即pb.go、grpc.pb.go、http.pb.go

```cmd
kratos proto client api/customer/customer.proto
```

### 生成server端相关代码

客户端代码在customer.go里面

```cmd
kratos proto server api/customer/customer.proto -t internal/service
```

然后编辑**customer.go文件**，实现顾客服务的4个功能：获取验证码GetVerifyCode、登录Login、退出登录Logout、预估价格EstimatePrice；

```go
package service

import (
	"context"
	"customer/api/verifyCode"
	"customer/internal/biz"
	"customer/internal/data"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/consul/api"
	"log"
	"regexp"
	"time"

	pb "customer/api/customer"
)

type CustomerService struct {
	pb.UnimplementedCustomerServer
	CD *data.CustomerData
	cb *biz.CustomerBiz
}

func NewCustomerService(cd *data.CustomerData, cb *biz.CustomerBiz) *CustomerService {
	return &CustomerService{
		CD: cd,
		cb: cb,
	}
}

func (s *CustomerService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	// 一，校验手机号
	pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
	// 生成一个正则表达式对象
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(req.Telephone) {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "电话号码格式错误",
		}, nil
	}
	// 二，通验证码服务生成验证码（服务间通信，grpc）
	// 使用服务发现
	// 1.获取consul客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "192.168.43.144:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	// 2.获取服务发现管理器
	// 创建一个新的 Consul 注册表实例
	dis := consul.New(consulClient)
	//selector.SetGlobalSelector(random.NewBuilder())
	selector.SetGlobalSelector(wrr.NewBuilder())
	//selector.SetGlobalSelector(p2c.NewBuilder())
	//log.Println(selector.GlobalSelector())
	// 2.1,连接目标grpc服务器
	endpoint := "discovery:///verifyCode"
	conn, err := grpc.DialInsecure(
		context.Background(),
		//grpc.WithEndpoint("localhost:9000"), // verifyCode grpc service 地址
		grpc.WithEndpoint(endpoint), // 目标服务的名字
		// 使用服务发现
		grpc.WithDiscovery(dis),
	)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码服务不可用",
		}, nil
	}
	//关闭,跟直接用 defer conn.Close() 关闭有什么区别？没区别
	defer func() {
		_ = conn.Close()
	}()
	// 2.2,发送获取验证码请求
	client := verifyCode.NewVerifyCodeClient(conn)
	reply, err := client.GetVerifyCode(context.Background(), &verifyCode.GetVerifyCodeRequest{
		Length: 6,
		Type:   2,
	})
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码获取错误",
		}, nil
	}

	// 三，redis的临时存储
	const life = 60
	if err := s.CD.SetVerifyCode(req.Telephone, reply.Code, life); err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码存储错误（Redis的操作服务）",
		}, nil
	}
	// 没有错误就返回正确结果
	return &pb.GetVerifyCodeResp{
		Code:           0,
		VerifyCode:     reply.Code, //关键的一步，这样就连起来了
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLife: life,
	}, nil
}

func (s *CustomerService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 一、校验电话和验证码，从redis获取
	code := s.CD.GetVerifyCode(req.Telephone)
	// 将redis中的code与req中的code比较，req中的VerifyCode哪里来的？body？
	if code == "" || code != req.VerifyCode {
		return &pb.LoginResp{
			Code:    1,
			Message: "验证码不匹配",
		}, nil
	}

	// 二、判定电话号码是否注册，来获取顾客信息
	customer, err := s.CD.GetCustomerByTelephone(req.Telephone)
	if err != nil {
		return &pb.LoginResp{
			Code:    1,
			Message: "顾客信息获取错误",
		}, nil
	}

	// 三、设置token，jwt-token
	token, err := s.CD.GenerateTokenAndSave(customer, biz.CustomerDuration*time.Second, biz.CustomerSecret)
	log.Println(err)
	if err != nil {
		return &pb.LoginResp{
			Code:    1,
			Message: "Token生成失败",
		}, nil
	}

	// 四，响应token
	return &pb.LoginResp{
		Code:          0,
		Message:       "login success",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.CustomerDuration,
	}, nil
}

func (s *CustomerService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	// 一，获取用户id
	claims, _ := jwt.FromContext(ctx)
	// 获取，断言使用
	claimsMap := claims.(jwtv5.MapClaims)

	// 二，删除用户的token
	if err := s.CD.DelToken(claimsMap["jti"]); err != nil {
		return &pb.LogoutResp{
			Code:    1,
			Message: "Token删除失败",
		}, nil
	}
	// 三，成功，响应
	return &pb.LogoutResp{
		Code:    0,
		Message: "logout success",
	}, nil
}

func (s *CustomerService) EstimatePrice(ctx context.Context, req *pb.EstimatePriceReq) (*pb.EstimatePriceResp, error) {
	price, err := s.cb.GetEstimatePrice(ctx, req.Origin, req.Destination)
	if err != nil {
		return &pb.EstimatePriceResp{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &pb.EstimatePriceResp{
		Code:        0,
		Message:     "SUCCESS",
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}
```

#### 初始化顾客服务相关结构体

##### service层

初始化CustomerService结构体，CustomerService结构体新增了两个成员

- 业务逻辑的组装（biz.CustomerBiz）

- 业务数据的访问（data.CustomerData）；

CustomerService绑定了4个功能：GetVerifyCode、Login、Logout、EstimatePrice；

```go
type CustomerService struct {
	pb.UnimplementedCustomerServer
	CD *data.CustomerData // 1.与mysql和redis交互
	cb *biz.CustomerBiz // 2.绑定了了费用预估方法（GetEstimatePrice）
}
func NewCustomerService(cd *data.CustomerData, cb *biz.CustomerBiz) *CustomerService {
	return &CustomerService{
		CD: cd,
		cb: cb,
	}
}
```

##### biz层

在biz/customer.go文件中，CustomerBiz{}结构体；

```go
type CustomerBiz struct{}

func NewCustomerBiz() *CustomerBiz {
	return &CustomerBiz{}
}
```

##### data层-CustomerData

在data/customer.go文件中，CustomerData嵌入data成员；

```go
type CustomerData struct {
	data *Data
}
func NewCustomerData(data *Data) *CustomerData {
	return &CustomerData{data: data}
}
```

##### data层-Data

嵌入成员为redis客户端和MySQL客户端；

```go
package data

import (
	"customer/internal/biz"
	"customer/internal/conf"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
// 增加 NewCustomerData 的 provider
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)

// Data .
type Data struct {
	// TODO wrapped database client
	// 操作Redis的客户端
	Rdb *redis.Client
	// 操作MySQL的客户端
	Mdb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	data := &Data{}
	// 一、初始化 Rdb
	// 连接redis，使用服务的配置，c就是解析后的配置信息(来自于config.yaml)
	redisURL := fmt.Sprintf("redis://%s/1?dial_timeout=%d", c.Redis.Addr, 1)
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		data.Rdb = nil
		log.Fatal(err)
	}
	// new client 不会立即连接，建立客户端，需要执行命令时才会连接
	data.Rdb = redis.NewClient(options)

	// 二、初始化Mdb
	// 连接mysql，使用配置
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		data.Mdb = nil
		log.Fatal(err)
	}
	data.Mdb = db

	// 三、开发阶段，自动迁移表。发布阶段，表结构稳定，不需要migrate
	migrateTable(db)
	cleanup := func() {
		// 清理了 Redis 连接
		_ = data.Rdb.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
	if err := db.AutoMigrate(&biz.Customer{}); err != nil {
		log.Info("customer table migrate error,err:", err)
	}
}

```

**NewData具体过程如下：**

1. 初始化redis客户端：

   ```go
   data := &Data{}
   // 1.初始化redisURL，提供redis地址和dial_timeout；
   redisURL := fmt.Sprintf("redis://%s/1?dial_timeout=%d", c.Redis.Addr, 1)
   // 2.解析redisURL为redis.Options；
   options, err := redis.ParseURL(redisURL)
   // 3.错误处理，将data.Rdb置空；
   if err != nil {
   	data.Rdb = nil
   	log.Fatal(err)
   }
   // 4.创建redis客户端；new client不会立即连接，需要执行命令时才会连接；
   data.Rdb = redis.NewClient(options)
   ```

2. 初始化MySQL客户端：

   ```go
   dsn := c.Database.Source
   db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
   if err != nil {
   	data.Mdb = nil
   	log.Fatal(err)
   }
   data.Mdb = db
   ```

3. 迁移数据表：migrateTable(db)，迁移Customer数据表，开发阶段，自动迁移表，发布阶段，表结构稳定，不需要migrate；

   ```go
   func migrateTable(db *gorm.DB) {
   	if err := db.AutoMigrate(&biz.Customer{}); err != nil {
   		log.Info("customer table migrate error,err:", err)
   	}
   }
   ```

4. cleanup中加入清理 Redis 连接操作：

   ```go
   cleanup := func() {
   	// 清理了 Redis 连接
   	_ = data.Rdb.Close()
   	log.NewHelper(logger).Info("closing the data resources")
   }
   ```

5. 返回结果：return data, cleanup, nil；

#### 生成验证码并存储

具体代码如下：

```go
func (s *CustomerService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	// 一，校验手机号
	pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
	// 生成一个正则表达式对象
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(req.Telephone) {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "电话号码格式错误",
		}, nil
	}
	// 二，通验证码服务生成验证码（服务间通信，grpc）
	// 使用服务发现
	// 1.获取consul客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "192.168.43.144:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	// 2.获取服务发现管理器
	// 创建一个新的 Consul 注册表实例
	dis := consul.New(consulClient)
	//selector.SetGlobalSelector(random.NewBuilder())
	selector.SetGlobalSelector(wrr.NewBuilder())
	//selector.SetGlobalSelector(p2c.NewBuilder())
	//log.Println(selector.GlobalSelector())
	// 2.1,连接目标grpc服务器
	endpoint := "discovery:///verifyCode"
	conn, err := grpc.DialInsecure(
		context.Background(),
		//grpc.WithEndpoint("localhost:9000"), // verifyCode grpc service 地址
		grpc.WithEndpoint(endpoint), // 目标服务的名字
		// 使用服务发现
		grpc.WithDiscovery(dis),
	)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码服务不可用",
		}, nil
	}
	//关闭,跟直接用 defer conn.Close() 关闭有什么区别？没区别
	defer func() {
		_ = conn.Close()
	}()
	// 2.2,发送获取验证码请求
	client := verifyCode.NewVerifyCodeClient(conn)
	reply, err := client.GetVerifyCode(context.Background(), &verifyCode.GetVerifyCodeRequest{
		Length: 6,
		Type:   2,
	})
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码获取错误",
		}, nil
	}

	// 三，redis的临时存储
	const life = 60
	if err := s.CD.SetVerifyCode(req.Telephone, reply.Code, life); err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: "验证码存储错误（Redis的操作服务）",
		}, nil
	}
	// 没有错误就返回正确结果
	return &pb.GetVerifyCodeResp{
		Code:           0,
		VerifyCode:     reply.Code, //关键的一步，这样就连起来了
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLife: life,
	}, nil
}
```

##### 校验手机号

验证电话号码格式的正确性，使用正则表达式匹配；

```go
pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
// 生成一个正则表达式对象
regexpPattern := regexp.MustCompile(pattern)
if !regexpPattern.MatchString(req.Telephone) {
	return &pb.GetVerifyCodeResp{
		Code:    1,
		Message: "电话号码格式错误",
	}, nil
}
```

##### 生成验证码

调用验证码服务（服务发现）：使用服务间通信grpc生成验证码；

1. 新建consul客户端：使用默认配置DefaultConfig；

   ```go
   consulConfig := api.DefaultConfig()
   consulConfig.Address = "192.168.43.144:8500"
   consulClient, err := api.NewClient(consulConfig)
   ```

2. 新建consul服务注册中心：

   ```go
   dis := consul.New(consulClient)；
   ```

3. 设置负载均衡策略，使用加权轮询wrr作为全局的负载均衡策略；

   ```go
   selector.SetGlobalSelector(wrr.NewBuilder())；
   ```

4. 获取验证码服务的客户端连接：使用grpc.DialInsecure()函数获取grpc.ClientConn；

   ```go
   endpoint := "discovery:///verifyCode"
   conn, err := grpc.DialInsecure(
   	context.Background(),
   	grpc.WithEndpoint(endpoint), //目标服务的名字
   	// 使用服务发现
   	grpc.WithDiscovery(dis),
   )
   // 并使用defer语句延迟关闭连接：
   defer func() {
   	_ = conn.Close()
   }()
   ```

5. 创建验证码服务客户端：

   ```go
   client := verifyCode.NewVerifyCodeClient(conn)
   ```

6. 获取验证码：发送获取验证码的请求，注意在本服务中需要提前添加verifyCode.proto，用于验证码服务；生成的验证码在reply里面;

   ```go
   reply, err := client.GetVerifyCode(context.Background(), &verifyCode.GetVerifyCodeRequest{
   	Length: 6,
   	Type:   2,
   })
   ```

##### 存储验证码

将验证码临时存储在redis里面（电话号码加前缀作为key，验证码作为value），并且设置过期时间；

```go
const life = 60
if err := s.CD.SetVerifyCode(req.Telephone, reply.Code, life); err != nil {
	return &pb.GetVerifyCodeResp{
		Code:    1,
		Message: "验证码存储错误（Redis的操作服务）",
	}, nil
}
```

SetVerifyCode具体实现过程在internal/data/customer.go里面，customer.go实现的功能有：

1. 初始化CustomerData结构体

2. **存储验证码**
3. 获取验证码

4. 根据电话获取顾客信息

5. 生成用户token并存储

6. 使用顾客ID，获取数据库中对应的token

7. 利用顾客ID，删除对应的token


SetVerifyCode具体代码如下：

```go
// 设置验证码的方法
func (cd CustomerData) SetVerifyCode(telephone, code string, ex int64) error {
	// 设置key, customer-verify-code
	status := cd.data.Rdb.Set(context.Background(), "CVC:"+telephone, code, time.Duration(ex)*time.Second)
	if _, err := status.Result(); err != nil {
		return err
	}
	return nil
}
```

在同目录中的internal/data/data.go文件里面，有redis和DB的初始化操作`NewData()`；关键操作，data.go进行依赖注入：

```go
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)；
```

##### 返回验证码

无错误，则返回验证码正确结果；

```go
// 没有错误就返回正确结果
return &pb.GetVerifyCodeResp{
	Code:           0,
	VerifyCode:     reply.Code, //关键的一步，这样就连起来了
	VerifyCodeTime: time.Now().Unix(),
	VerifyCodeLife: life,
}, nil
```

#### 用户登录

```go
func (s *CustomerService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 一、校验电话和验证码，从redis获取
	code := s.CD.GetVerifyCode(req.Telephone)
	// 将redis中的code与req中的code比较，req中的VerifyCode哪里来的？body？
	if code == "" || code != req.VerifyCode {
		return &pb.LoginResp{
			Code:    1,
			Message: "验证码不匹配",
		}, nil
	}

	// 二、判定电话号码是否注册，来获取顾客信息
	customer, err := s.CD.GetCustomerByTelephone(req.Telephone)
	if err != nil {
		return &pb.LoginResp{
			Code:    1,
			Message: "顾客信息获取错误",
		}, nil
	}

	// 三、设置token，jwt-token
	token, err := s.CD.GenerateTokenAndSave(customer, biz.CustomerDuration*time.Second, biz.CustomerSecret)
	log.Println(err)
	if err != nil {
		return &pb.LoginResp{
			Code:    1,
			Message: "Token生成失败",
		}, nil
	}

	// 四，响应token
	return &pb.LoginResp{
		Code:          0,
		Message:       "login success",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.CustomerDuration,
	}, nil
}
```

##### 获取验证码

根据手机号从redis获取验证码

```go
code := s.CD.GetVerifyCode(req.Telephone)
```

##### 校验验证码

> 注意：生成验证码的功能测试，使用postman的grpc功能，输入为json格式。
>
> ![image-20250428115648186](C:\Users\Mr chen\AppData\Roaming\Typora\typora-user-images\image-20250428115648186.png)

将redis中的code与req中的code比较，有错误则响应“验证码不匹配”；

```go
// 将redis中的code与req中的code比较，req中的VerifyCode哪里来的？body中自己根据短信输入
if code == "" || code != req.VerifyCode {
	return &pb.LoginResp{
		Code:    1,
		Message: "验证码不匹配",
	}, nil
}

```

##### 获取客户信息

根据用户电话号码，获取客户信息（也表示检验手机号是否被注册）

```go
// 二、判定电话号码是否注册，来获取顾客信息
customer, err := s.CD.GetCustomerByTelephone(req.Telephone)
if err != nil {
	return &pb.LoginResp{
		Code:    1,
		Message: "顾客信息获取错误",
	}, nil
}
```

在客户表中（MySQL数据库）查询此客户，有注册则返回客户信息，如果手机号未被注册，则创建一个新客户db.Create(customer)，并返回customer；

GetCustomerByTelephone具体实现如下：

```go
// 根据电话，获取顾客信息
func (cd CustomerData) GetCustomerByTelephone(telephone string) (*biz.Customer, error) {
	// 查询基于电话
	customer := &biz.Customer{}
	result := cd.data.Mdb.Where("telephone=?", telephone).First(customer)
	// query 执行成功，同时查询到记录
	if result.Error == nil && customer.ID > 0 {
		return customer, nil
	}

	// 有记录不存在的错误
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 创建customer并返回
		customer.Telephone = telephone
		customer.Name = sql.NullString{Valid: false}
		customer.Email = sql.NullString{Valid: false}
		customer.Wechat = sql.NullString{Valid: false}
		resultCreate := cd.data.Mdb.Create(customer)
		// 插入成功
		if resultCreate.Error != nil {
			return nil, resultCreate.Error
		} else {
			return customer, nil
		}
	}
	// 有错误，但是不是记录不存在的错误，不做业务逻辑处理
	return nil, result.Error
}
```

##### 生成token并存储

传入参数customer、有效期、密钥（secret口令）；实现过程如下：

```go
// 三、设置token，jwt-token
token, err := s.CD.GenerateTokenAndSave(customer, biz.CustomerDuration*time.Second, biz.CustomerSecret)
log.Println(err)
if err != nil {
	return &pb.LoginResp{
		Code:    1,
		Message: "Token生成失败",
	}, nil
}
```

GenerateTokenAndSave具体实现：

```go
// 生成token并存储
func (cd CustomerData) GenerateTokenAndSave(c *biz.Customer, duration time.Duration, secret string) (string, error) {
	// 一，生成token
	// 处理token中的数据
	// 标准的 JWT 的 payload（载荷，属于数据部分）
	claims := jwt.RegisteredClaims{
		// 签发方，签发机构
		Issuer: "LaoMaDJ",
		// 说明
		Subject: "customer-authentication",
		// 签发给谁使用
		Audience: []string{"customer", "other"},
		// 有效期至
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		// 何时启用
		NotBefore: nil,
		// 签发时间
		IssuedAt: jwt.NewNumericDate(time.Now()),
		// ID, 用户的ID
		ID: fmt.Sprintf("%d", c.ID),
	}
	// 生成token，使用HS256进行签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名token，注意传入的是字节数组
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	// 二，存储
	c.Token = signedToken
	c.TokenCreatedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	// save()：不存在则插入，存在则更新；
	if result := cd.data.Mdb.Save(c); result.Error != nil {
		return "", result.Error
	}

	// 操作成功，返回生成的签名
	return signedToken, nil
}
```

1. 使用jwt-token的标准结构jwt.RegisteredClaims{}，这是JWT 的 载荷（payload），属于数据部分； 

   ```go
   claims := jwt.RegisteredClaims{
   	// 签发方，签发机构
   	Issuer: "LaoMaDJ",
   	// 说明
   	Subject: "customer-authentication",
   	// 签发给谁使用
   	Audience: []string{"customer", "other"},
   	// 有效期至
   	ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
   	// 何时启用
   	NotBefore: nil,
   	// 签发时间
       IssuedAt: jwt.NewNumericDate(time.Now()),
   	// ID, 用户的ID
   	ID: fmt.Sprintf("%d", c.ID),
   }
   ```

2. 生成原始token结构体，使用HS256进行签名；

   ```go
   token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
   ```

3. 生成真正的token：使用参数的密钥（secret）对原始token进行签名

   ```go
   signedToken, err := token.SignedString([]byte(secret))
   ```

   得到一个签名后的令牌字符串（signedToken，这是真正的token）；

4. 存储token：将signedToken添加到customer结构体中，并使用DB.Save(customer)将整个结构体存储到MySQL数据库； 

   ```go
   c.Token = signedToken
   c.TokenCreatedAt = sql.NullTime{
   	Time:  time.Now(),
       Valid: true,
   }
   // save()：不存在则插入，存在则更新；
   if result := cd.data.Mdb.Save(c); result.Error != nil {
       return "", result.Error
   }
   ```

5. 返回用户token：操作成功，返回生成的signedToken！

##### 返回 LoginResp，响应登录成功！

```go
return &pb.LoginResp{
		Code:          0,
		Message:       "login success",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.CustomerDuration,
	}, nil
}
```

#### 退出登录

后端的退出登录需要删除用户对应的token（将token清空，设置为“”）；

```go
func (s *CustomerService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	// 一，获取用户id
	claims, _ := jwt.FromContext(ctx)
	// 获取，断言使用
	claimsMap := claims.(jwtv5.MapClaims)

	// 二，删除用户的token
	if err := s.CD.DelToken(claimsMap["jti"]); err != nil {
		return &pb.LogoutResp{
			Code:    1,
			Message: "Token删除失败",
		}, nil
	}
	// 三，成功，响应
	return &pb.LogoutResp{
		Code:    0,
		Message: "logout success",
	}, nil
}
```

1. 获取用户customer的id：从ctx的claimsMap结构中获取，id存放在key = "jti"里面。

   ```go
   claims, _ := jwt.FromContext(ctx)
   claimsMap := claims.(jwtv5.MapClaims)
   id := claimsMap["jti"]
   ```

2. 删除token：将用户的token和TokenCreatedAt置空

   ```go
   // 利用顾客ID，删除对应的token
   func (cd CustomerData) DelToken(id interface{}) error {
   	c := &biz.Customer{}
   	// 找到customer
   	if result := cd.data.Mdb.First(c, id); result.Error != nil {
   		return result.Error
   	}
   	// 删除customer的token
   	c.Token = ""
       // Valid：布尔值，true 表示 Time 包含有效值，false 表示数据库 NULL 值
   	c.TokenCreatedAt = sql.NullTime{Valid: false}
   	cd.data.Mdb.Save(c)
   	return nil
   }
   ```

3. 响应LogoutResp：退出登录成功！

#### 费用预估

调用grpc远程服务，具体实现如下：

```go
func (s *CustomerService) EstimatePrice(ctx context.Context, req *pb.EstimatePriceReq) (*pb.EstimatePriceResp, error) {
	price, err := s.cb.GetEstimatePrice(ctx, req.Origin, req.Destination)
	if err != nil {
		return &pb.EstimatePriceResp{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &pb.EstimatePriceResp{
		Code:        0,
		Message:     "SUCCESS",
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}
```

1. **调用费用预估服务**：从req中获取起点Origin、终点Destination，然后调用GetEstimatePrice服务获取price：

   ```go
   price, err := s.cb.GetEstimatePrice(ctx, req.Origin, req.Destination)
   if err != nil {
   	return &pb.EstimatePriceResp{
           Code:    1,
   		Message: err.Error(),
   	}, nil
   }
   ```

   GetEstimatePrice调用过程如下：

   1. 获取consul客户端；

   2. 新建consul服务注册中心；

   3. 获取费用预估服务连接；

   4. 创建Valuation客户端；

   5. 获取预估费用并返回；

   GetEstimatePrice具体实现过程在biz/customer.go中，具体代码如下：

   ```go
   func (cb *CustomerBiz) GetEstimatePrice(ctx context.Context, origin, destination string) (int64, error) {
   	// 一，grpc 获取
   	// 1.获取consul客户端
   	consulConfig := api.DefaultConfig()
   	consulConfig.Address = "192.168.43.144:8500"
   	consulClient, err := api.NewClient(consulConfig)
   	// 2.获取服务发现管理器
   	dis := consul.New(consulClient)
   	if err != nil {
   		return 0, err
   	}
   	// 2.1,连接目标grpc服务器,并使用Valuation服务
   	endpoint := "discovery:///Valuation"
   	conn, err := grpc.DialInsecure(
   		context.Background(),
   		grpc.WithEndpoint(endpoint), // 目标服务的名字
   		// 使用服务发现
   		grpc.WithDiscovery(dis),
   	)
   	if err != nil {
   		return 0, nil
   	}
   	//关闭
   	defer func() {
   		_ = conn.Close()
   	}()
   
   	// 2.2,发送获取费用请求
   	client := valuation.NewValuationClient(conn) // 创建客户端
   	reply, err := client.GetEstimatePrice(context.Background(), &valuation.GetEstimatePriceReq{
   		Origin:      origin,
   		Destination: destination,
   	})
   	if err != nil {
   		return 0, err
   	}
   	return reply.Price, nil
   }
   ```

2. **响应成功信息**：起点、终点、费用；

   ```go
   return &pb.EstimatePriceResp{
   	Code:        0,
   	Message:     "SUCCESS",
       Origin:      req.Origin,
   	Destination: req.Destination,
   	Price:       price,
   }, nil
   ```

### 在http中注册顾客服务并添加中间件

在internal/server/http.go文件中，注意仅在http.go里操作，因为顾客服务只提供http访问方式。

```go
package server

import (
	"context"
	"customer/api/customer"
	v1 "customer/api/helloworld/v1"
	"customer/internal/biz"
	"customer/internal/conf"
	"customer/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, customerService *service.CustomerService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			// 自己设置的中间件
			// CORS，全部的请求（响应）都使用该中间件
			selector.Server(MWCors()).Match(func(ctx context.Context, operation string) bool {
				return true
			}).Build(),
			selector.Server(
				jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
					return []byte(biz.CustomerSecret), nil
				}),
				customerJWT(customerService),
			).Match(func(ctx context.Context, operation string) bool {
				// 根据自己的需要完成是否启用该中间件的校验
				noJWT := map[string]struct{}{
					"/api.customer.Customer/Login":         {},
					"/api.customer.Customer/GetVerifyCode": {},
				}
				if _, exists := noJWT[operation]; exists {
					return false
				}
				return true
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	// 注册customer的http服务
	customer.RegisterCustomerHTTPServer(srv, customerService)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}
```

#### 注册顾客服务到http中

在函数参数中传入customerService，然后在函数体增加一行关键代码

```go
customer.RegisterCustomerHTTPServer(srv, customerService)
```

#### 添加中间件

3个自定义中间件：跨域资源共享、登录状态校验-请求中间件、JWT与顾客存储token的校验；

##### 跨域资源共享中间件

MWCors()，全部的请求/响应都使用，具体实现如下：

```go
selector.Server(MWCors()).Match(func(ctx context.Context, operation string) bool {
    return true
}).Build()
```

`return true` 表示 **匹配函数始终返回 `true`**，即任何操作（`operation`）都会被认为是匹配的。

> 备注：可以考虑引入gin-contrib/cors中间件库，简化操作；

##### 登录状态校验-请求中间件

只有部分请求/响应使用；即token合法校验，token合法才能访问后面的资源；主要检验token的格式规范、有效期，签名有效等，使用kratos内置的jwt中间件；

```go
jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
    return []byte(biz.CustomerSecret), nil
}),
```

参数token与登录服务中的原始token结构体是一致的；这里仍然需要return secret密钥，用于验证 JWT 令牌的合法性；

##### 请求头中token与顾客存储token的校验

`customerJWT(customerService)` ，只有部分请求/响应使用，中间件具体实现如下：

```go
package server

import (
	"context"
	"customer/internal/service"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"strings"
)

// customerJWT 生成中间件的方法，将ctx的请求头中的token与数据库中的token进行比较，不符合则不给通过；
func customerJWT(customerService *service.CustomerService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 一，获取存储在jwt中的用户（顾客）id
			claims, ok := jwt.FromContext(ctx)
			fmt.Println(claims, ok)
			if !ok {
				// 没有获取到 claims
				return nil, errors.Unauthorized("UNAUTHORIZED", "claims not found")
			}
			// 1.2 断言使用
			claimsMap := claims.(jwtv5.MapClaims)
			id := claimsMap["jti"] //获取JWT的ID

			// 二，获取id对应的顾客的token
			token, err := customerService.CD.GetToken(id)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "customer not found")
			}

			// 三，比对数据表中的token与请求的token是否一致
			// 获取请求头
			header, _ := transport.FromServerContext(ctx)
			// 从header获取token
			auths := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
			jwtToken := auths[1]
			// 比较请求中的token与数据表中获取的token是否一致
			if jwtToken != token {
				return nil, errors.Unauthorized("UNAUTHORIZED", "token was updated")
			}
			// 四，校验通过，发行，继续执行
			// 交由下个中间件（handler）处理
			return handler(ctx, req)
		}
	}
}
```

1. 获取存储在JWT ID（也对应customer{}模型的id）

   ```go
   // 一，获取存储在jwt中的用户（顾客）id
   claims, ok := jwt.FromContext(ctx)
   fmt.Println(claims, ok)
   if !ok {
       // 没有获取到 claims
       return nil, errors.Unauthorized("UNAUTHORIZED", "claims not found")
   }
   // 1.2 断言使用
   claimsMap := claims.(jwtv5.MapClaims)
   id := claimsMap["jti"] //获取JWT的ID
   ```

   "jti" 通常用来表示 JWT ID，它也可以唯一标识一个 JWT 令牌；

2. 获取id对应的customer的token

   ```go
   token, err := customerService.CD.GetToken(id)；
   ```

   GetToken方法在/data/customer.go中，具体实现如下：

   ```go
   // 利用顾客ID，获取数据库中对应的token
   func (cd CustomerData) GetToken(id interface{}) (string, error) {
   	c := &biz.Customer{}
   	if result := cd.data.Mdb.First(c, id); result.Error != nil {
   		return "", result.Error
   	}
   	return c.Token, nil
   }
   ```

3. 获取请求头header中对应的jwtToken：

   ```go
   // 获取请求头
   header, _ := transport.FromServerContext(ctx)
   // 从header获取token
   auths := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
   jwtToken := auths[1]
   ```

4. 将token与jwtToken进行比较，检验是否相等：

   ```go
   // 比较请求中的token与数据表中获取的token是否一致
   if jwtToken != token {
   	return nil, errors.Unauthorized("UNAUTHORIZED", "token was updated")
   }
   // 四，校验通过，发行，继续执行
   // 交由下个中间件（handler）处理
   return handler(ctx, req)
   ```

   验证通过后继续执行下一个中间件：return handler(ctx, req)；

##### 剔除不需要使用中间件的请求

某些请求（如登录、获取验证码）不需要使用后面两个中间件，使用map存储无需使用此中间件的路由；

```go
// 根据自己的需要完成是否启用该中间件的校验
noJWT := map[string]struct{}{
	"/api.customer.Customer/Login":         {},
	"/api.customer.Customer/GetVerifyCode": {},
}
if _, exists := noJWT[operation]; exists {
	return false
}
```

### 顾客服务不需要注册到consul；

### 依赖注入

共3个地方需要进行依赖注入

1. service层：在service.go 文件中。

   ```go
   var ProviderSet = wire.NewSet(NewGreeterService, NewCustomerService)
   ```

2. biz层：在biz.go文件中。

   ```go
   var ProviderSet = wire.NewSet(NewGreeterUsecase, NewCustomerBiz)
   ```

3. data层：在data.go文件中。

   ```go
   var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)
   ```

4. 执行依赖注入命令：

   ```shell
   go generate ./...
   ```

### 启动服务

```shell
kratos run
```

## 地图服务

地图服务调用了高德地图API

### 创建地图服务模板

```shell
kratos new map
```

### 修改配置文件

主要是修改服务端口，http是8200, grpc是9200

```yaml
server:
  http:
    addr: 0.0.0.0:8200
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9200
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
```

### 生成地图服务proto文件

```shell
kratos proto add api/mapService/mapService.proto
```

### 编辑proto文件

定义获取驾驶信息api，其中req包含起点、终点，reply包含起点、终点、行驶距离、行驶时长；

```protobuf
syntax = "proto3";

package api.mapService;

option go_package = "map/api/mapService;mapService";

service MapService {
  rpc GetDrivingInfo (GetDrivingInfoReq) returns (GetDrivingReply);
}

message GetDrivingInfoReq {
  string origin = 1; // 起点
  string destination = 2; // 终点
}
message GetDrivingReply {
  string origin = 1;
  string destination = 2;
  string distance = 3; // 行驶距离
  string duration = 4; // 行驶时长
}
```

### 生成client端代码

```shell
kratos proto client api/mapService/mapService.proto
```

### 生成server端代码

```shell
kratos proto server api/mapService/mapService.proto
```

生成mapservice.go文件，地图服务实现GetDrivingInfo（获取驾驶信息），具体实现过程如下：

#### 初始化地图服务相关结构体

##### service层

初始化MapServiceService{}结构体，MapServiceService实现了GetDriverInfo方法；

```go
type MapServiceService struct {
	pb.UnimplementedMapServiceServer
	msbiz *biz.MapServiceBiz // 绑定了获取驾驶信息功能GetDriverInfo
}
func NewMapServiceService(msbiz *biz.MapServiceBiz) *MapServiceService {
	return &MapServiceService{
		msbiz: msbiz,
	}
}
```

##### biz层

初始化MapServiceBiz{}结构体，嵌入日志记录工具log.Helper；MapServiceBiz实现了GetDriverInfo()方法；实现过程见后续；

```go
type MapServiceBiz struct {
	log *log.Helper
}
func NewMapServiceBiz(logger log.Logger) *MapServiceBiz {
	return &MapServiceBiz{log: log.NewHelper(logger)}
}
```

#### 获取驾驶信息

```go
func (s *MapServiceService) GetDrivingInfo(ctx context.Context, req *pb.GetDrivingInfoReq) (*pb.GetDrivingReply, error) {
	distance, duration, err := s.msbiz.GetDriverInfo(req.Origin, req.Destination)
	if err != nil {
		return nil, errors.New(200, "LBS_ERROR", "lbs api error")
	}
	return &pb.GetDrivingReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Distance:    distance,
		Duration:    duration,
	}, nil
}
```

GetDriverInfo函数具体实现在biz层。

##### 获取行驶距离和时间

从req里面获取起点Origin、终点Destination信息，然后调用biz层的GetDriverInfo方法。

```go
distance, duration, err := s.msbiz.GetDriverInfo(req.Origin, req.Destination)
if err != nil {
    return nil, errors.New(200, "LBS_ERROR", "lbs api error")
}
```

GetDriverInfo方法调用了高德地图API，实现过程如下：

```go
// 获取驾驶信息
func (msbiz *MapServiceBiz) GetDriverInfo(origin, destination string) (string, string, error) {
    // 一，请求获取
    key := "2b08113dc921fac3afd0992a2b45862e"
    api := "https://restapi.amap.com/v3/direction/driving"
    parameters := fmt.Sprintf("origin=%s&destination=%s&extensions=base&output=json&key=%s", origin, destination, key)
    url := api + "?" + parameters
    resp, err := http.Get(url)
    if err != nil {
       return "", "", err
    }
    defer func() {
       _ = resp.Body.Close()
    }()
    body, err := io.ReadAll(resp.Body) // io.Reader
    if err != nil {
       return "", "", err
    }
    //fmt.Println(string(body))
    // 二，解析出来,json
    ddResp := &DirectionDrivingResp{}
    if err := json.Unmarshal(body, ddResp); err != nil {
       return "", "", err
    }

    // 三，判定LSB请求结果
    if ddResp.Status == "0" {
       return "", "", errors.New(ddResp.Info)
    }

    // 四，正确返回，默认使用第一条路线
    path := ddResp.Route.Paths[0]
    return path.Distance, path.Duration, nil
}

type DirectionDrivingResp struct {
    Status   string `json:"status,omitempty"`
    Info     string `json:"info,omitempty"`
    Infocode string `json:"infocode,omitempty"`
    Count    string `json:"count,omitempty"`
    Route    struct {
       Origin      string `json:"origin,omitempty"`
       Destination string `json:"destination,omitempty"`
       Paths       []Path `json:"paths,omitempty"`
    } `json:"route"`
}
type Path struct {
    Distance string `json:"distance,omitempty"`
    Duration string `json:"duration,omitempty"`
    Strategy string `json:"strategy,omitempty"`
}
```

1. **拼接url请求**：主要包括起点、终点、高德地图API密钥key

   ```go
   key := "2b08113dc921fac3afd0992a2b45862e"
   api := "https://restapi.amap.com/v3/direction/driving"
   parameters := fmt.Sprintf("origin=%s&destination=%s&extensions=base&output=json&key=%s", origin, destination, key)
   url := api + "?" + parameters
   ```

2. 发起http请求，获取地图 API 返回的结果resp

   ```go
   resp, err := http.Get(url)
   if err != nil {
       return "", "", err
   }
   defer func() {
       _ = resp.Body.Close()
   }()
   ```

3. 读取响应的body内容

   ```go
   body, err := io.ReadAll(resp.Body) // io.Reader
   if err != nil {
       return "", "", err
   }
   ```

4. 解析body：解析body的内容到DirectionDrivingResp{}（这个结构用于表示驾驶数据）；

   ```go
   ddResp := &DirectionDrivingResp{}
   if err := json.Unmarshal(body, ddResp); err != nil {
       return "", "", err
   }
   ```

5. 判端请求是否成功：判定LSB请求结果；

   ```go
   if ddResp.Status == "0" {
       return "", "", errors.New(ddResp.Info)
   }
   ```

6. 选择最佳路线：默认使用第一条路线

   ```go
   // 四，正确返回，默认使用第一条路线
   path := ddResp.Route.Paths[0]
   ```

7. 响应结果：返回行驶距离和行驶时间。

   ```go
   return path.Distance, path.Duration, nil
   ```

##### 响应驾驶信息GetDrivingReply

返回包括起点、终点、行驶距离、行驶时间；

```go
return &pb.GetDrivingReply{
    Origin:      req.Origin,
    Destination: req.Destination,
    Distance:    distance,
    Duration:    duration,
}, nil
```

### 在grpc中注册地图服务

仅在map/internal/server/grpc.go里面增加地图服务，因为地图服务只提供grpc访问方式；

```go
package server

import (
    v1 "map/api/helloworld/v1"
    mapService2 "map/api/mapService"
    "map/internal/conf"
    "map/internal/service"

    "github.com/go-kratos/kratos/v2/log"
    "github.com/go-kratos/kratos/v2/middleware/recovery"
    "github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, mapService *service.MapServiceService, logger log.Logger) *grpc.Server {
    var opts = []grpc.ServerOption{
       grpc.Middleware(
          recovery.Recovery(),
       ),
    }
    if c.Grpc.Network != "" {
       opts = append(opts, grpc.Network(c.Grpc.Network))
    }
    if c.Grpc.Addr != "" {
       opts = append(opts, grpc.Address(c.Grpc.Addr))
    }
    if c.Grpc.Timeout != nil {
       opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
    }
    srv := grpc.NewServer(opts...)
    v1.RegisterGreeterServer(srv, greeter)
    // 注册
    mapService2.RegisterMapServiceServer(srv, mapService)
    return srv
}
```

### 将map服务注册到consul

在map/cmd/map/main.go里面，将地图服务注册到consul；

```go
// go build -ldflags "-X main.Version=x.y.z"
var (
    // Name is the name of the compiled software.
    Name string = "Map"
    // Version is the version of the compiled software.
    Version string = "1.0.0"
    // flagconf is the config flag.
    flagconf string
    //id, _ = os.Hostname()
    id = Name + "-" + uuid.NewString()
)

func init() {
    flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
    // 一，获取consul客户端
    consulConfig := api.DefaultConfig()
    consulConfig.Address = "192.168.43.144:8500"
    consulClient, err := api.NewClient(consulConfig)
    if err != nil {
       log.Fatal(err)
    }
    // 二，获取consul注册管理器
    reg := consul.New(consulClient)
    return kratos.New(
       kratos.ID(id),
       kratos.Name(Name),
       kratos.Version(Version),
       kratos.Metadata(map[string]string{}),
       kratos.Logger(logger),
       kratos.Server(
          gs,
          hs,
       ),
       // 三，创建服务时，指定注册管理器
       kratos.Registrar(reg),
    )
}

func main() {
    flag.Parse()
    logger := log.With(log.NewStdLogger(os.Stdout),
       "ts", log.DefaultTimestamp,
       "caller", log.DefaultCaller,
       "service.id", id,
       "service.name", Name,
       "service.version", Version,
       "trace.id", tracing.TraceID(),
       "span.id", tracing.SpanID(),
    )
    c := config.New(
       config.WithSource(
          file.NewSource(flagconf),
       ),
    )
    defer c.Close()

    if err := c.Load(); err != nil {
       panic(err)
    }

    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
       panic(err)
    }

    app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
    if err != nil {
       panic(err)
    }
    defer cleanup()

    // start and wait for stop signal
    if err := app.Run(); err != nil {
       panic(err)
    }
}
```

1. 配置地图服务基础信息：Name、Version、id；

2. 新建consul客户端

   ```go
   consulClient, err := api.NewClient(consulConfig)
   ```

3. 新建consul服务注册中心：`reg := consul.New(consulClient)；`

4. 创建服务：将地图服务注册到consul中，关键步骤：`kratos.Registrar(reg)；`

### 依赖注入

有2个地方需要进行依赖注入；

1. **service层**：在service.go文件里面

```go
var ProviderSet = wire.NewSet(NewGreeterService, NewMapServiceService)
```

2. **biz层**：在biz.go文件里面

```go
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewMapServiceBiz)
```

3. 然后需要执行依赖注入命令：`go generate ./...`

### 启动服务

```shell
kratos run
```

## 费用预估服务

即计价服务valuation

### 创建valuation服务模板

```shell
kratos new valuation
```

### 修改配置文件

http服务端口8300，grpc服务端口9300；

```yaml
server:
  http:
    addr: 0.0.0.0:8300
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9300
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:123456@tcp(192.168.43.144:3306)/laomadj_valuation?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
```

### 创建proto文件

```shell
kratos proto add
```

### 编辑proto文件

包括mapService.proto（同map服务）和valuation.proto，valuation服务的req包含起点、终点，reply包含起点、终点、价格；

```protobuf
syntax = "proto3";
package api.valuation;
option go_package = "valuation/api/valuation;valuation";

service Valuation {
  rpc GetEstimatePrice (GetEstimatePriceReq) returns (GetEstimatePriceReply);
}

message GetEstimatePriceReq {
  string origin = 1;
  string destination = 2;
}

message GetEstimatePriceReply {
  string origin = 1;
  string destination = 2;
  int64 price = 3;
}
```

### 生成client端代码

```shell
kratos proto client
```

### 生成server端代码

```shell
kratos proto server
```

生成valuation.go文件，其中ValuationService{}结构体实现了GetEstimatePrice（获取预估价格）功能；具体实现如下：

```go
package service

import (
    "context"
    "github.com/go-kratos/kratos/v2/errors"
    "valuation/internal/biz"

    pb "valuation/api/valuation"
)

type ValuationService struct {
    pb.UnimplementedValuationServer
    // 引用业务对象
    vb *biz.ValuationBiz
}

func NewValuationService(vb *biz.ValuationBiz) *ValuationService {
    return &ValuationService{
       vb: vb,
    }
}

func (s *ValuationService) GetEstimatePrice(ctx context.Context, req *pb.GetEstimatePriceReq) (*pb.GetEstimatePriceReply, error) {
    // 得到距离、时长
    distance, duration, err := s.vb.GetDrivingInfo(ctx, req.Origin, req.Destination)
    if err != nil {
       return nil, errors.New(200, "MAP ERROR", "get driving info error")
    }
    // 费用
    price, err := s.vb.GetPrice(ctx, distance, duration, 1, 9)
    if err != nil {
       return nil, errors.New(200, "PRICE ERROR", "cal price error")
    }
    return &pb.GetEstimatePriceReply{
       Origin:      req.Origin,
       Destination: req.Destination,
       Price:       price,
    }, nil
}
```

#### 初始化费用预估服务相关结构体

##### service层

初始化ValuationService结构体，即增加*biz.ValuationBiz成员：

```go
type ValuationService struct {
    pb.UnimplementedValuationServer
    // 引用业务对象
    vb *biz.ValuationBiz
}

func NewValuationService(vb *biz.ValuationBiz) *ValuationService {
    return &ValuationService{
       vb: vb,
    }
}
```

##### biz层

初始化ValuationBiz结构体，biz.ValuationBiz实现了两个功能：获取价格GetPrice()、获取驾驶信息GetDrivingInfo()，具体实现见后面；

```go
type PriceRuleWork struct {
    CityID      uint  `gorm:"" json:"city_id"`
    StartFee    int64 `gorm:"" json:"start_fee"`        // 起步费
    DistanceFee int64 `gorm:"" json:"distance_fee"`     // 里程费
    DurationFee int64 `gorm:"" json:"duration_fee"`     // 时长费
    StartAt     int   `gorm:"type:int" json:"start_at"` // 0 [0
    EndAt       int   `gorm:"type:int" json:"end_at"`   // 7 0)
}

type PriceRule struct {
    gorm.Model
    PriceRuleWork
}

// 定义操作priceRule的接口
type PriceRuleInterface interface {
    // 获取规则
    GetRule(cityid uint, curr int) (*PriceRule, error)
}

type ValuationBiz struct {
    pri PriceRuleInterface
}

func NewValuationBiz(pri PriceRuleInterface) *ValuationBiz {
    return &ValuationBiz{
       pri: pri,
    }
}
```

##### data层- PriceRuleData

在/data/valuation.go文件里面，PriceRuleData嵌入data，PriceRuleData真正实现了GetRule函数；

```go
type PriceRuleData struct {
    data *Data
}

func NewPriceRuleInterface(data *Data) biz.PriceRuleInterface {
    return &PriceRuleData{data: data}
}
```

##### data层- Data

在/data/data.go文件里面，Data嵌入gorm.DB成员，用于操作MySQL的客户端；

```go
type Data struct {
    // TODO wrapped database client
    // 操作MySQL的客户端
    Mdb *gorm.DB
}
```

使用NewData()函数进行初始化，初始化过程实现如下：

```go
// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
    data := &Data{}
    // 初始Mdb
    // 连接mysql，使用配置
    dsn := c.Database.Source
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
       data.Mdb = nil
       log.Fatal(err)
    }
    data.Mdb = db
    // 三，开发阶段，自动迁移表。发布阶段，表结构稳定，不需要migrate.
    migrateTable(db)
    cleanup := func() {
       log.NewHelper(logger).Info("closing the data resources")
    }
    return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
    if err := db.AutoMigrate(&biz.PriceRule{}); err != nil {
       log.Info("price_rule table migrate error, err:", err)
    }
    // 插入一些riceRule的测试数据
    rules := []biz.PriceRule{
       {
          Model: gorm.Model{ID: 1},
          PriceRuleWork: biz.PriceRuleWork{
             CityID:      1,
             StartFee:    300,
             DistanceFee: 35,
             DurationFee: 10, // 5m
             StartAt:     7,
             EndAt:       23,
          },
       },
       {
          Model: gorm.Model{ID: 2},
          PriceRuleWork: biz.PriceRuleWork{
             CityID:      1,
             StartFee:    350,
             DistanceFee: 35,
             DurationFee: 10, // 5m
             StartAt:     23,
             EndAt:       24,
          },
       },
       {
          Model: gorm.Model{ID: 3},
          PriceRuleWork: biz.PriceRuleWork{
             CityID:      1,
             StartFee:    400,
             DistanceFee: 35,
             DurationFee: 10, // 5m
             StartAt:     0,
             EndAt:       7,
          },
       },
    }
    // 如果记录已存在，将更新所有列（字段）
    db.Clauses(clause.OnConflict{UpdateAll: true}).Create(rules)
}
```

1. 初始化MySQL客户端。

   ```go
   data := &Data{}
   // 初始Mdb
   // 连接mysql，使用配置
   dsn := c.Database.Source
   db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
   if err != nil {
       data.Mdb = nil
       log.Fatal(err)
   }
   data.Mdb = db
   ```

2. 迁移数据表：migrateTable(db)自动迁移价格规则数据表（PriceRule），并插入几条基础数据，migrateTable函数具体实现如下：

   ```go
   func migrateTable(db *gorm.DB) {
       if err := db.AutoMigrate(&biz.PriceRule{}); err != nil {
          log.Info("price_rule table migrate error, err:", err)
       }
       // 插入一些riceRule的测试数据
       rules := []biz.PriceRule{
          {
             Model: gorm.Model{ID: 1},
             PriceRuleWork: biz.PriceRuleWork{
                CityID:      1,
                StartFee:    300,
                DistanceFee: 35,
                DurationFee: 10, // 5m
                StartAt:     7,
                EndAt:       23,
             },
          },
          {
             Model: gorm.Model{ID: 2},
             PriceRuleWork: biz.PriceRuleWork{
                CityID:      1,
                StartFee:    350,
                DistanceFee: 35,
                DurationFee: 10, // 5m
                StartAt:     23,
                EndAt:       24,
             },
          },
          {
             Model: gorm.Model{ID: 3},
             PriceRuleWork: biz.PriceRuleWork{
                CityID:      1,
                StartFee:    400,
                DistanceFee: 35,
                DurationFee: 10, // 5m
                StartAt:     0,
                EndAt:       7,
             },
          },
       }
       // 如果记录已存在，将更新所有列（字段）
       db.Clauses(clause.OnConflict{UpdateAll: true}).Create(rules)
   }
   ```

#### 获取预估费用

```go
func (s *ValuationService) GetEstimatePrice(ctx context.Context, req *pb.GetEstimatePriceReq) (*pb.GetEstimatePriceReply, error) {
    // 得到距离、时长
    distance, duration, err := s.vb.GetDrivingInfo(ctx, req.Origin, req.Destination)
    if err != nil {
       return nil, errors.New(200, "MAP ERROR", "get driving info error")
    }
    // 费用
    price, err := s.vb.GetPrice(ctx, distance, duration, 1, 9)
    if err != nil {
       return nil, errors.New(200, "PRICE ERROR", "cal price error")
    }
    return &pb.GetEstimatePriceReply{
       Origin:      req.Origin,
       Destination: req.Destination,
       Price:       price,
    }, nil
}
```

##### 获取驾驶信息

即获取行驶距离、时长。

```go
// 得到距离、时长
distance, duration, err := s.vb.GetDrivingInfo(ctx, req.Origin, req.Destination)
if err != nil {
    return nil, errors.New(200, "MAP ERROR", "get driving info error")
}
```

GetDrivingInfo实际上是调用map服务（服务发现），函数实现过程如下：

```go
// 获取距离和时长
func (*ValuationBiz) GetDrivingInfo(ctx context.Context, origin, destination string) (distance string, duration string, err error) {
    // 一，发出GRPC请求
    // 使用服务发现
    // 1.获取consul客户端
    consulConfig := api.DefaultConfig()
    consulConfig.Address = "192.168.43.144:8500"
    consulClient, err := api.NewClient(consulConfig)
    if err != nil {
       return
    }
    // 2.获取服务发现管理器
    dis := consul.New(consulClient)
    // 2.1,连接目标grpc服务器
    endpoint := "discovery:///Map"
    conn, err := grpc.DialInsecure(
       context.Background(),
       //ctx,
       grpc.WithEndpoint(endpoint), // 目标服务的名字
       grpc.WithDiscovery(dis),     // 使用服务发现
       // 中间件
       grpc.WithMiddleware(
          // tracing 的客户端中间件
          tracing.Client(),
       ),
    )

    if err != nil {
       return
    }
    //关闭
    defer func() {
       _ = conn.Close()
    }()

    // 2.2,发送获取驾驶距离和时长请求，RPC调用
    client := mapService.NewMapServiceClient(conn)
    reply, err := client.GetDrivingInfo(context.Background(), &mapService.GetDrivingInfoReq{
       Origin:      origin,
       Destination: destination,
    })

    if err != nil {
       return
    }
    distance, duration = reply.Distance, reply.Duration
    // 返回正确信息
    return
}
```

1. 新建consul客户端；

2. 新建consul服务注册中心；

3. 获取map服务的client连接：在配置中加入链路追踪客户端中间件tracing.Client()； 

   ```go
   conn, err := grpc.DialInsecure(
       context.Background(),
       grpc.WithEndpoint(endpoint), // 目标服务的名字
       grpc.WithDiscovery(dis),     // 使用服务发现
       // 中间件
       grpc.WithMiddleware(
          // tracing 的客户端中间件
          tracing.Client(),
       ),
   )
   ```

4. 创建map服务客户端；

5. 调用GetDrivingInfo方法：获取行驶距离、时长；

6. 返回距离、时长； 

##### 获取费用

```go
// 费用
price, err := s.vb.GetPrice(ctx, distance, duration, 1, 9)
if err != nil {
    return nil, errors.New(200, "PRICE ERROR", "cal price error")
}
```

这里调用了biz/valuation.go中的GetPrice方法，GetPrice方法属于ValuationBiz结构体。GetPrice函数具体实现如下：

> 注意：ValuationBiz嵌入了一个接口成员PriceRuleInterface，这个接口包含一个GetRule方法，参考前方的初始化相关结构体。

```go
// 获取价格
func (vb *ValuationBiz) GetPrice(ctx context.Context, distance, duration string, cityId uint, curr int) (int64, error) {
    // 一，获取规则
    rule, err := vb.pri.GetRule(cityId, curr)
    if err != nil {
       return 0, err
    }
    // 二，将距离和时长，转换为int64
    distancInt64, err := strconv.ParseInt(distance, 10, 64)
    if err != nil {
       return 0, err
    }

    durationInt64, err := strconv.ParseInt(duration, 10, 64)
    if err != nil {
       return 0, err
    }

    // 三，基于rule计算
    distancInt64 /= 1000        // 公里
    durationInt64 /= 60         // 分钟？
    var startDistance int64 = 5 // 起始距离 5公里？
    total := rule.StartFee +
       rule.DistanceFee*(distancInt64-startDistance) +
       rule.DurationFee*durationInt64
    return total, nil
}
```

1. **获取计价规则：**

   ```go
   rule, err := vb.pri.GetRule(cityId, curr)
   ```

   参数为城市id和当前时间：不同城市、不同时间有不同的计价规则；GetRule是PriceRuleInterface接口中的方法，PriceRuleInterface接口实际由data/valuation.go中的PriceRuleData实现，具体实现如下：

   > PriceRuleData结构体的初始化略，参考前面的初始化相关结构体。

   ```go
   // PriceRuleData 实现 PriceRuleInterface，curr 表示当前时间
   func (prd *PriceRuleData) GetRule(cityid uint, curr int) (*biz.PriceRule, error) {
       pr := &biz.PriceRule{}
       // "start_at <= ? AND end_at > ?" 表示当前时刻在某时间范围内，对应该范围内的规则
       result := prd.data.Mdb.Where("city_id=? AND start_at <= ? AND end_at > ?", cityid, curr, curr).First(pr)
       if result.Error != nil {
          return nil, result.Error
       }
       return pr, nil
   }
   ```

2. 将距离和时长转换为int64类型。

3. **计算总费用**：总费用 = 起步费 + 里程费 ×（路程 － 起始距离）+ 时长费 × 时长；

   ```go
   // 三，基于rule计算
   distancInt64 /= 1000        // 公里
   durationInt64 /= 60         // 分钟
   var startDistance int64 = 5 // 起始距离 5公里
   total := rule.StartFee +
       rule.DistanceFee*(distancInt64-startDistance) +
       rule.DurationFee*durationInt64
   return total, nil
   ```

##### 响应预估费用

包括起点、终点、预估费用。

```go
return &pb.GetEstimatePriceReply{
    Origin:      req.Origin,
    Destination: req.Destination,
    Price:       price,
}, nil
```

### 在GRPC中注册费用预估服务并添加中间件

费用预估服务只提供GRPC访问方式；

#### 注册费用预估服务到grpc中

```go
// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, valuationService *service.ValuationService, logger log.Logger) *grpc.Server {
    var opts = []grpc.ServerOption{
       grpc.Middleware(
          recovery.Recovery(),
          // 加入 tracing 中间件
          tracing.Server(),
       ),
    }
    if c.Grpc.Network != "" {
       opts = append(opts, grpc.Network(c.Grpc.Network))
    }
    if c.Grpc.Addr != "" {
       opts = append(opts, grpc.Address(c.Grpc.Addr))
    }
    if c.Grpc.Timeout != nil {
       opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
    }
    srv := grpc.NewServer(opts...)
    v1.RegisterGreeterServer(srv, greeter)
    // 注册
    valuation.RegisterValuationServer(srv, valuationService)
    return srv
}
```

#### 添加中间件

在grpc.go和http.go中添加链路追踪的服务端中间件tracing.Server()，http服务应该是用不到这个中间件，客户端需要使用时要加入tracing.Client()；

### 服务注册到consul

在main.go文件里进行操作。费用预估服务中多了jaeger链路追踪；

```go
// go build -ldflags "-X main.Version=x.y.z"
var (
    // Name is the name of the compiled software.
    Name string = "Valuation"
    // Version is the version of the compiled software.
    Version string = "1.0.0"
    // flagconf is the config flag.
    flagconf string
    // id, _ = os.Hostname()
    id = Name + "-" + uuid.NewString()
)

func init() {
    flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
    // 一，获取consul客户端
    consulConfig := api.DefaultConfig()
    consulConfig.Address = "192.168.43.144:8500"
    consulClient, err := api.NewClient(consulConfig)
    if err != nil {
       log.Fatal(err)
    }
    // 二，获取consul注册管理器
    reg := consul.New(consulClient)
    // 初始化 TP
    tracerURL := "http://192.168.43.144:14268/api/traces"
    if err := initTracer(tracerURL); err != nil {
       log.Error(err)
    }
    return kratos.New(
       kratos.ID(id),
       kratos.Name(Name),
       kratos.Version(Version),
       kratos.Metadata(map[string]string{}),
       kratos.Logger(logger),
       kratos.Server(
          gs,
          hs,
       ),
       // 三，创建服务时，指定服务器注册
       kratos.Registrar(reg),
    )
}

func main() {
    flag.Parse()
    logger := log.With(log.NewStdLogger(os.Stdout),
       "ts", log.DefaultTimestamp,
       "caller", log.DefaultCaller,
       "service.id", id,
       "service.name", Name,
       "service.version", Version,
       "trace.id", tracing.TraceID(),
       "span.id", tracing.SpanID(),
    )
    c := config.New(
       config.WithSource(
          file.NewSource(flagconf),
       ),
    )
    defer c.Close()

    if err := c.Load(); err != nil {
       panic(err)
    }

    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
       panic(err)
    }

    app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
    if err != nil {
       panic(err)
    }
    defer cleanup()

    // start and wait for stop signal
    if err := app.Run(); err != nil {
       panic(err)
    }
}

// 初始化Tracer
// @param url string 指定 Jaeger 的API接口
// :14268/api/traces
func initTracer(url string) error {
    //一，建立jaeger客户端，称之为：exporter，导出器
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
    if err != nil {
       return err
    }
    // 创建 TracerProvider，TracerProvider 是 API 的入口点 (Entry Point)
    tracerProvider := trace.NewTracerProvider(
       //采样器设置，AlwaysSample表示每个跨度都会被采样
       trace.WithSampler(trace.AlwaysSample()),
       // 使用 exporter 作为批处理程序
       trace.WithBatcher(exporter),
       // 将当前服务的信息，作为资源告知给TracerProvider
       trace.WithResource(resource.NewSchemaless(
          // 必要的配置，设置一个服务名称的键值对；
          semconv.ServiceNameKey.String(Name),
          // 任意的其他属性配置
          attribute.String("exporter", "jaeger"),
       )),
    )
    // 三，设置全局的TP
    otel.SetTracerProvider(tracerProvider)
    return nil
}
```

#### 配置费用预估服务基础信息

#### 新建consul客户端

#### 新建consul服务注册中心

#### 初始化链路追踪

使用jaeger作为链路追踪工具。

```go
// 初始化 TP
tracerURL := "http://192.168.43.144:14268/api/traces"
if err := initTracer(tracerURL); err != nil {
    log.Error(err)
}
```

initTracer()函数实现过程如下：

1. 新建jaeger客户端：jaeger客户端也称为导出器（exporter）。

   ```go
   //一，建立jaeger客户端，称之为：exporter，导出器
   exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
   if err != nil {
       return err
   }
   ```

2. 新建TracerProvider：这是Tracers的工厂，用于管理和创建跟踪器（Tracer）。

   ```go
   // 创建 TracerProvider，TracerProvider 是 API 的入口点 (Entry Point)
   tracerProvider := trace.NewTracerProvider(
       // 采样器设置，AlwaysSample表示每个跨度都会被采样
       trace.WithSampler(trace.AlwaysSample()),
       // 使用 exporter 作为批处理程序
       trace.WithBatcher(exporter),
       // 将当前服务的信息，作为资源告知给TracerProvider
       trace.WithResource(resource.NewSchemaless(
          // 必要的配置，设置一个服务名称的键值对；
          semconv.ServiceNameKey.String(Name),
          // 任意的其他属性配置，使用Jaeger作为导出器exporter
          attribute.String("exporter", "jaeger"),
       )),
   )
   ```

3. 设置全局的TracerProvider

   ```go
   otel.SetTracerProvider(tracerProvider)
   ```

   这个函数将使得整个应用程序使用TracerProvider 来创建 Tracer 实例。初始化链路追踪的函数完成；

4. 启动JaegerUI界面，查看详情！

#### 创建服务

```go
kratos.Registrar(reg)
```

### 依赖注入

共有3个位置需要进行依赖注入

1. service层：在service.go文件中

   ```go
   var ProviderSet = wire.NewSet(NewGreeterService, NewValuationService)
   ```

2. biz层：在biz.go文件中

   ```go
   var ProviderSet = wire.NewSet(NewGreeterUsecase, NewValuationBiz)
   ```

3. data层：在data.go文件中

   ```go
   var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewPriceRuleInterface)
   ```

4. 执行依赖注入命令：`go generate ./...`

### 启动服务

```shell
kratos run
```

## 司机服务（完成部分功能）

包括司机申请及审核、登录、退出登录、司机状态、司机位置管理。

### 创建司机服务模板

```shell
kratos new driver
```

### 修改配置文件

修改服务端口，并整合consul和jaeger的配置；

#### 优化config.yml配置

将前面零散的consul地址和和jaeger的url集中管理，config.yml配置如下：

```yaml
server:
  http:
    addr: 0.0.0.0:8400
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9400
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:123456@tcp(192.168.43.144:3306)/laomadj_driver?parseTime=True&loc=Local
  redis:
    addr: 192.168.43.144:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
service:
  consul:
    address: 192.168.43.144:8500
  jaeger:
    url: http://192.168.43.144:14268/api/traces
```

#### 修改conf.proto配置

配置信息和config.yml的配置相对应，在internal/conf/conf.proto中配置：

```protobuf
syntax = "proto3";
package kratos.api;

option go_package = "driver/internal/conf;conf";

import "google/protobuf/duration.proto";

message Bootstrap {
  Server server = 1;
  Data data = 2;
  Service service = 3;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  HTTP http = 1;
  GRPC grpc = 2;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }
  message Redis {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration read_timeout = 3;
    google.protobuf.Duration write_timeout = 4;
  }
  Database database = 1;
  Redis redis = 2;
}

// 下面为增加的内容
message Service {
  message Consul {
    string address = 1;
  }
  message Jaeger {
    string url = 1;
  }
  Consul consul = 1;
  Jaeger jaeger = 2;
}
```

#### 执行生成命令

```shell
kratos proto client internal/conf/conf.proto
```

生成最终的配置文件conf.pb.go；

### 生成司机服务proto文件

```shell
kratos proto add
```

### 编辑proto文件

包括verifyCode.proto文件（参考验证码服务）和driver.proto，后者包括5个api：校验身份证号码、获取验证码、提交电话号码、司机登录、退出登录。

```protobuf
syntax = "proto3";

package api.driver;
// 导入包
import "google/api/annotations.proto";

option go_package = "driver/api/driver;driver";

service Driver {

  // 校验身份证号码
  rpc IDNoCheck (IDNoCheckReq) returns (IDNoCheckResp) {
   option (google.api.http) = {
    post: "/driver/idno-check",
    body: "*",
   };
  }

  // 获取验证码
  rpc GetVerifyCode (GetVerifyCodeReq) returns (GetVerifyCodeResp) {
   option (google.api.http) = {
    get: "/driver/get-verify-code/{telephone}"
   };
  }

  // 提交电话号码
  rpc SubmitPhone (SubmitPhoneReq) returns (SubmitPhoneResp) {
   option (google.api.http) = {
    post: "/driver/submit-phone",
    body: "*",
   };
  }

  // 登录
  rpc Login (LoginReq) returns (LoginResp) {
   option (google.api.http) = {
    post: "/driver/login",
    body: "*",
   };
  }

  // 退出
  rpc Logout (LogoutReq) returns (LogoutResp) {
   option (google.api.http) = {
    delete: "/driver/logout",
   };
  }
}

// 校验身份证号码消息
message IDNoCheckReq {
  string name = 1;
  string idno = 2;
};
message IDNoCheckResp {
  int64 code = 1;
  string message = 2;
  string status = 3;
};

// 获取验证码的消息
message GetVerifyCodeReq {
  string telephone = 1;
};

message GetVerifyCodeResp {
  int64 code = 1;
  string message = 2;
  // 验证码
  string verify_code = 3;
  // 生成时间 unix timestamp
  int64 verify_code_time = 4;
  // 有效期，单位 second
  int32 verify_code_life = 5;
};

// 提交电话号码请求消息
message SubmitPhoneReq {
  string telephone = 1;
};
message SubmitPhoneResp {
  int64 code = 1;
  string message = 2;
  string status = 3;
};

// 登录的消息
message LoginReq {
  string telephone = 1;
  string verify_code = 2;
};

message LoginResp {
  int64 code = 1;
  string message = 2;
  // token,登录表示，特殊的字符串，JWT 编码格式
  string token = 3;
  // 生成时间 unix timestamp
  int64 token_create_at = 4;
  // 有效期，单位 second
  int32 token_life = 5;
};

message LogoutReq {
};

message LogoutResp {
  int64 code = 1;
  string message = 2;
};
```

注意看如何使用验证码服务？以电话号码为key，生成验证码为value，存储在redis里面，**验证码类型和长度实际上是写死的**。

### 生成client端代码

```shell
kratos proto client
```

### 生成server端代码

```shell
kratos proto server
```

命令生成driver.go文件，其中包含5个功能： 

- 校验身份证号码（暂未实现）
- 获取验证码（使用grpc）
- 提交电话号码（暂未实现）
- 司机登录
- 退出登录

#### 初始化司机服务相关结构体：

##### service层

初始化DriverService{}，嵌入成员biz.DriverBiz；

```go
type DriverService struct {
    pb.UnimplementedDriverServer
    Bz *biz.DriverBiz
}

func NewDriverService(bz *biz.DriverBiz) *DriverService {
    return &DriverService{
       Bz: bz,
    }
}
```

##### biz层

初始化DriverBiz{}，表示司机业务逻辑，嵌入DriverInterface接口（包含7个功能）；

```go
// 司机相关的资源操作接口
type DriverInterface interface {
    GetVerifyCode(context.Context, string) (string, error)
    FetchVerifyCode(context.Context, string) (string, error)
    FetchInfoByTel(context.Context, string) (*Driver, error)
    InitDriverInfo(context.Context, string) (*Driver, error)
    GetSavedVerifyCode(context.Context, string) (string, error)
    SaveToken(context.Context, string, string) error
    GetToken(context.Context, string) (string, error)
}

// 司机业务逻辑
type DriverBiz struct {
    DI DriverInterface
}

// DriverBiz 构造器
func NewDriverBiz(di DriverInterface) *DriverBiz {
    return &DriverBiz{
       DI: di,
    }
}
```

##### data层- DriverData

DriverData嵌入Data，并实现了DriverInterface接口，具体实现见后续：

```go
type DriverData struct {
    data *Data
}

func NewDriverInterface(data *Data) biz.DriverInterface {
    return &DriverData{data: data}
}
```

##### data层-Data

Data结构体嵌入了MySQL客户端、redis客户端、相关工具配置地址（consul和jaeger），使用NewData()函数对其进行初始化。特别注意NewData加入了参数conf.Service，并还要进行后续添加。

```go
// Data .
type Data struct {
    // TODO wrapped database client
    // 操作MySQL的客户端
    Mdb *gorm.DB
    // 操作Redis的客户端
    Rdb *redis.Client
    // 相关工具配置地址
    cs *conf.Service
}

// NewData .
func NewData(c *conf.Data, cs *conf.Service, logger log.Logger) (*Data, func(), error) {
    data := &Data{
       cs: cs,
    }
    // 连接redis，使用服务的配置，c就是解析之后的配置信息,此处的redis没有配置密码
    redisURL := fmt.Sprintf("redis://%s/2?dial_timeout=%d", c.Redis.Addr, 1)
    options, err := redis.ParseURL(redisURL)
    if err != nil {
       data.Rdb = nil
       log.Fatal(err)
    }
    // new client 不会立即连接，建立客户端，需要执行命令时才会连接
    data.Rdb = redis.NewClient(options)

    // 初始Mdb
    // 连接mysql，使用配置
    dsn := c.Database.Source
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
       data.Mdb = nil
       log.Fatal(err)
    }
    data.Mdb = db
    // 三，开发阶段，自动迁移表。发布阶段，表结构稳定，不需要migrate.
    migrateTable(db)

    cleanup := func() {
       log.NewHelper(logger).Info("closing the data resources")
    }
    return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
    // 自动迁移相关表
    if err := db.AutoMigrate(&biz.Driver{}); err != nil {
       log.Fatal(err)
    }
}
```

NewData具体过程如下：

1. 构造data基本结构体：本服务中多了一个配置cs；

   ```go
   data := &Data{
       cs: cs,
   }
   ```

2. 初始化redis：

   ```go
   // 连接redis，使用服务的配置，c就是解析之后的配置信息,此处的redis没有配置密码
   redisURL := fmt.Sprintf("redis://%s/2?dial_timeout=%d", c.Redis.Addr, 1)
   options, err := redis.ParseURL(redisURL)
   if err != nil {
       data.Rdb = nil
       log.Fatal(err)
   }
   // new client 不会立即连接，建立客户端，需要执行命令时才会连接
   data.Rdb = redis.NewClient(options)
   ```

3. 初始化mysql：

   ```go
   // 初始Mdb
   // 连接mysql，使用配置
   dsn := c.Database.Source
   db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
   if err != nil {
       data.Mdb = nil
       log.Fatal(err)
   }
   data.Mdb = db
   ```

4. 迁移数据表：开发阶段，自动迁移表，发布阶段，表结构稳定，不需要migrate.

   ```go
   migrateTable(db)
   ```

##### 相关函数需要添加conf.Service参数

cmd/driver中的3个文件（main.go、wire.go、wire_gen.go）都需要添加这个参数；

1. **wire_gen.go文件**：在wireApp()中增加 *conf.Service参数，在调用newApp()的时候使用参数。

   ```go
   // wireApp init kratos application.
   func wireApp(confService *conf.Service, confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
       dataData, cleanup, err := data.NewData(confData, confService, logger)
       if err != nil {
          return nil, nil, err
       }
       greeterRepo := data.NewGreeterRepo(dataData, logger)
       greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
       greeterService := service.NewGreeterService(greeterUsecase)
       driverInterface := data.NewDriverInterface(dataData)
       driverBiz := biz.NewDriverBiz(driverInterface)
       driverService := service.NewDriverService(driverBiz)
       grpcServer := server.NewGRPCServer(confServer, greeterService, driverService, logger)
       httpServer := server.NewHTTPServer(confServer, greeterService, driverService, logger)
       app := newApp(confService, logger, grpcServer, httpServer)
       return app, func() {
          cleanup()
       }, nil
   }
   ```

2. **wire.go文件**：在wireApp()中增加*conf.Service参数；

   ```go
   // wireApp init kratos application.
   func wireApp(*conf.Service, *conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
       panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
   }
   ```

3. **main.go文件**：2个地方加入；

   1. 在newApp()中增加*conf.Service参数；

      ```go
      func newApp(cs *conf.Service, logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
          //代码见前面
      }
      ```

   2. 在main()函数中，调用wireApp()时传入实参：

      ```go
      func main() {
      // 略
       app, cleanup, err := wireApp(bc.Service, bc.Server, bc.Data, logger)
      // 略
      }
      ```

#### 校验身份证号码：暂未实现；

#### 生成验证码并存储

1. 调用验证码grpc服务，以电话号码tel为key，生成验证码为value，存储在redis里面，**验证码类型和长度实际上是写死的**，具体过程如下：

```go
func (s *DriverService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
    // 获取验证码
    code, err := s.Bz.GetVerifyCode(ctx, req.Telephone)
    if err != nil {
       return &pb.GetVerifyCodeResp{
          Code:    1,
          Message: err.Error(),
       }, nil
    }
    // 响应
    return &pb.GetVerifyCodeResp{
       Code:           0,
       Message:        "SUCCESS",
       VerifyCode:     code,
       VerifyCodeTime: time.Now().Unix(),
       VerifyCodeLife: 60,
    }, nil
}
```

2. biz/driver.go的GetVerifyCode实现如下：

```go
// 实现获取验证码的业务逻辑
func (db *DriverBiz) GetVerifyCode(ctx context.Context, tel string) (string, error) {
    // 一，校验手机号
    pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
    // 返回一个编译后的正则表达式对象
    regexpPattern := regexp.MustCompile(pattern)
    if !regexpPattern.MatchString(tel) {
       return "", errors.New(200, "DRIVER", "driver telephone error")
    }
    // 二，调用data/获取验证码
    return db.DI.GetVerifyCode(ctx, tel)
}
```

3. DriverBiz嵌入了DriverInterface接口，由data/driver.go中的实现，因此此处的GetVerifyCode具体实现如下：

```go
func (dt *DriverData) GetVerifyCode(ctx context.Context, tel string) (string, error) {
    // grpc 请求
    consulConfig := api.DefaultConfig()
    consulConfig.Address = dt.data.cs.Consul.Address
    consulClient, err := api.NewClient(consulConfig)
    // 注册服务管理器
    dis := consul.New(consulClient)
    if err != nil {
       return "", err
    }
    endpoint := "discovery:///VerifyCode"
    conn, err := grpc.DialInsecure(
       //ctx
       context.Background(),
       grpc.WithEndpoint(endpoint), // 目标服务的名字
       grpc.WithDiscovery(dis),     // 使用服务发现
    )
    if err != nil {
       return "", err
    }
    //关闭
    defer conn.Close()
    // 2.2,发送获取验证码请求
    client := verifyCode.NewVerifyCodeClient(conn)
    // 作为客户端在进行远程调用
    reply, err := client.GetVerifyCode(ctx, &verifyCode.GetVerifyCodeRequest{
       Length: 6,
       Type:   1,
    })
    if err != nil {
       return "", err
    }
    // 三，redis的临时存储
    // 设置key, customer-verify-code
    status := dt.data.Rdb.Set(ctx, "DVC:"+tel, reply.Code, 60*time.Second)
    if _, err := status.Result(); err != nil {
       return "", err
    }
    return reply.Code, nil
}
```

#### 提交电话号码

1. 将司机信息入库，并设置状态为stop，函数实际调用了biz层的InitDriverInfo()方法，biz层又调用data层的InitDriverInfo()方法。

```go
// 提交电话号码
func (s *DriverService) SubmitPhone(ctx context.Context, req *pb.SubmitPhoneReq) (*pb.SubmitPhoneResp, error) {
    // 将司机信息入库，并设置状态为stop暂时停用(核心逻辑）
    driver, err := s.Bz.InitDriverInfo(ctx, req.Telephone)
    if err != nil {
       return &pb.SubmitPhoneResp{
          Code:    1,
          Message: "司机号码提交失败",
       }, nil
    }

    return &pb.SubmitPhoneResp{
       Code:    0,
       Message: "司机号码提交成功",
       Status:  driver.Status.String,
    }, nil
}
```

2. biz/driver.go中的InitDriverInfo方法实现：

```go
// 比对验证码是否一致
// 将司机信息入库
func (db *DriverBiz) InitDriverInfo(ctx context.Context, tel string) (*Driver, error) {
    // 校验验证码（略）
    // 司机是否已经注册的校验（略）？
    // 司机是否在黑名单中校验（略）？
    if tel == "" {
       return nil, errors.New(1, "telephone is empty", "")
    }

    return db.DI.InitDriverInfo(ctx, tel)
}
```

3. data/driver.go层的InitDriverInfo方法实现：

```go
// 初始化司机信息
func (dt *DriverData) InitDriverInfo(ctx context.Context, tel string) (*biz.Driver, error) {
    // 入库，设置状态为stop
    driver := biz.Driver{}
    driver.Telephone = tel
    driver.Status = sql.NullString{
       String: "stop",
       Valid:  true,
    }
    if err := dt.data.Mdb.Create(&driver).Error; err != nil {
       return nil, err
    }
    return &driver, nil
}
```

#### 司机登录

1. 函数实际调用了biz层的校验登录CheckLogin，生成token；

```go
// 登录 service
func (s *DriverService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
    // 由 biz 层完成业务逻辑处理
    token, err := s.Bz.CheckLogin(ctx, req.Telephone, req.VerifyCode)
    if err != nil {
       log.Println(err)
       return &pb.LoginResp{
          Code:    1,
          Message: "司机登录失败",
       }, nil
    }
    return &pb.LoginResp{
       Code:          0,
       Message:       "司机登录成功",
       Token:         token,
       TokenCreateAt: time.Now().Unix(),
       TokenLife:     biz.DriverTokenLife,
    }, nil
}
```

2. biz/driver.go的CheckLogin方法实现，共有3个主要步骤：

   ①校验验证码是否正确；

   ②登录时生成token。（token在登录时生成，这点很重要）

   ③存储token；通过调用相应位置的函数实现。

```go
// 验证登录信息方法
func (db *DriverBiz) CheckLogin(ctx context.Context, tel, verifyCode string) (string, error) {
    // 1.验证验证码是否正确
    code, err := db.DI.GetSavedVerifyCode(ctx, tel)
    if err != nil {
       return "", err
    }
    if verifyCode != code {
       return "", errors.New(1, "verify code error", "'")
    }
    // 2.生成token
    token, err := generateJWT(tel)
    if err != nil {
       return "", err
    }
    // 3.token存储到driver表中
    if err := db.DI.SaveToken(ctx, tel, token); err != nil {
       return "", err
    }
    // 返回token
    return token, nil
}
```

（1）data/driver.go中的GetSavedVerifyCode实现：获取存储在redis中的验证码（在前面生成验证码时存储在redis里面），以tel作为参数。

```go
// 获取已存储的验证码
func (dt *DriverData) GetSavedVerifyCode(ctx context.Context, tel string) (string, error) {
    return dt.data.Rdb.Get(ctx, "DVC:"+tel).Result()
}
```

（2）biz/driver.go中的generateJWT方法（和CheckLogin在同一层）：即生成token，以tel为参数。

```go
// 生成JWT token
func generateJWT(tel string) (string, error) {
    // 构建token类型
    claims := jwt.RegisteredClaims{
       Issuer:    "LaomaDJ",
       Subject:   "driver authentication",
       Audience:  []string{"driver"},
       ExpiresAt: jwt.NewNumericDate(time.Now().Add(DriverTokenLife * time.Second)),
       NotBefore: nil,
       IssuedAt:  jwt.NewNumericDate(time.Now()),
       ID:        tel,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    // 签名
    // 生成token字符串
    tokenString, err := token.SignedString([]byte(SecretKey))
    if err != nil {
       return "", err
    }
    return tokenString, nil
}
```

（3）data/driver.go中的SaveToken方法：将token存储到mysql数据库中（以tel、token为参数），

通过电话号码查找对应的driver，然后将token放入driver结构体中，然后更新保存Driver。

```go
// 存储token到数据库
func (dt *DriverData) SaveToken(ctx context.Context, tel, token string) error {
    //先获取司机信息
    driver := biz.Driver{}
    if err := dt.data.Mdb.Where("telephone=?", tel).First(&driver).Error; err != nil {
       return err
    }
    //再更新司机信息
    driver.Token = sql.NullString{
       String: token,
       Valid:  true,
    }
    if err := dt.data.Mdb.Save(&driver).Error; err != nil {
       return err
    }
    return nil
}
```

#### 退出登录：暂未实现

### 在grpc和http中注册司机服务并添加中间件

司机服务提供http和grpc两种访问方式。

#### 注册服务：在grpc和http中都添加司机服务

1. **注册服务到grpc**：在server/grpc.go中。

   ```go
   // NewGRPCServer new a gRPC server.
   func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, driverService *service.DriverService, logger log.Logger) *grpc.Server {
       var opts = []grpc.ServerOption{
          grpc.Middleware(
             recovery.Recovery(),
          ),
       }
       if c.Grpc.Network != "" {
          opts = append(opts, grpc.Network(c.Grpc.Network))
       }
       if c.Grpc.Addr != "" {
          opts = append(opts, grpc.Address(c.Grpc.Addr))
       }
       if c.Grpc.Timeout != nil {
          opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
       }
       srv := grpc.NewServer(opts...)
       v1.RegisterGreeterServer(srv, greeter)
       driver.RegisterDriverServer(srv, driverService)
       return srv
   }
   ```

2. **注册服务到http中并添加中间件**：在server/http.go中。

   ```go
   // NewHTTPServer new an HTTP server.
   func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, driverService *service.DriverService, logger log.Logger) *http.Server {
       var opts = []http.ServerOption{
          http.Middleware(
             recovery.Recovery(),
             // JWT 中间件
             selector.Server(
                jwt.Server(func(token *jwtv5.Token) (interface{}, error) { 
                   return []byte(biz.SecretKey), nil // 用于验证客户端发送的JWT令牌；
                }),
                DriverToken(driverService),
             ).Match(func(ctx context.Context, operation string) bool {
                // 记录不需要的校验的
                noJWT := map[string]struct{}{
                   "/api.driver.Driver/Login":         {},
                   "/api.driver.Driver/GetVerifyCode": {},
                   "/api.driver.Driver/SubmitPhone":   {},
                }
                if _, exists := noJWT[operation]; exists {
                   return false
                }
                return true
             }).Build(),
          ),
       }
       if c.Http.Network != "" {
          opts = append(opts, http.Network(c.Http.Network))
       }
       if c.Http.Addr != "" {
          opts = append(opts, http.Address(c.Http.Addr))
       }
       if c.Http.Timeout != nil {
          opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
       }
       srv := http.NewServer(opts...)
       v1.RegisterGreeterHTTPServer(srv, greeter)
       driver.RegisterDriverHTTPServer(srv, driverService)
       return srv
   }
   ```

#### 添加中间件

共有2个自定义的中间件，但是只需要在http服务中添加。

##### 登录状态校验-请求中间件

只有部分请求/响应使用，即token合法校验，token合法才能访问后面的资源。主要检验token的格式规范、有效期，签名是否有效等。使用kratos内置的jwt中间件。

```go
jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
    return []byte(biz.SecretKey), nil // 用于验证客户端发送的JWT令牌；
}),
```

##### 请求头中的token与顾客存储token的校验

```go
DriverToken(driverService)
```

只有部分请求/响应使用，具体实现在server/token.go中，步骤如下：

1. 从上下文中获取电话号码tel；
2. 根据tel获取到数据库中司机（driver结构体中）的token信息，注意这里只返回token信息；
3. 获取请求头（RequestHeader）中携带的token；
4. 将请求头中携带的token与数据库中司机的token进行比较，无错则继续往下执行；
5. 记录登录司机的信息到上下文中：根据tel查询司机信息driver（注意这里是返回整个driver结构体），然后将司机信息存储在上下文中，**这一步是司机服务独有的一个步骤。**

```go
// 返回中间件的函数
func DriverToken(service *service.DriverService) middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
       return func(ctx context.Context, req interface{}) (interface{}, error) {
          // 1.校验JWT，获取其中的司机标识tel
          claims, ok := jwt.FromContext(ctx)
          if !ok {
             return nil, errors.Unauthorized("Unauthorized", "claims not found")
          }
          claimsMap := claims.(jwtv5.MapClaims)
          tel := claimsMap["jti"]
          // 2.利用tel，获取存储在司机表（MySQL）中的token（这里是返回token）；
          token, err := service.Bz.DI.GetToken(ctx, tel.(string))
          if err != nil {
             return nil, errors.Unauthorized("Unauthorized", "driver token not found")
          }
          // 3.比对两个token(和请求头中的）
          header, _ := transport.FromServerContext(ctx)
          auths := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
          reqToken := auths[1]
          if token != reqToken {
             return nil, errors.Unauthorized("Unauthorized", "token was updated")
          }
          // 4.记录登录司机信息（这里是返回driver）
          driver, err := service.Bz.DI.FetchInfoByTel(ctx, tel.(string))
          if err != nil {
             return nil, errors.Unauthorized("Unauthorized", "driver was found")
          }
          // 基于当前的ctx，构建新的带有值的ctx
          ctxWithDriver := context.WithValue(ctx, "driver", driver)
          //ctxWithDriver.Value("driver")
          // 5.jwt校验通过，继续下一个handler
          return handler(ctxWithDriver, req)
       }
    }
}
```

### 注册司机服务到consul

在main.go文件里面操作；注意newApp()函数中增加了参数conf.Service；

```go
// go build -ldflags "-X main.Version=x.y.z"
var (
    // Name is the name of the compiled software.
    Name string = "Driver"
    // Version is the version of the compiled software.
    Version string = "1.0.0"
    // flagconf is the config flag.
    flagconf string
    // id, _ = os.Hostname()
    id = Name + "-" + uuid.NewString()
)

func init() {
    flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(cs *conf.Service, logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
    // 初始化consul服务注册中心
    reg, err := initServiceRegistry(cs.Consul.Address)
    if err != nil {
       panic(err)
    }
    // 链路追踪
    if err := initTracer(cs.Jaeger.Url); err != nil {
       panic(err)
    }
    return kratos.New(
       kratos.ID(id),
       kratos.Name(Name),
       kratos.Version(Version),
       kratos.Metadata(map[string]string{}),
       kratos.Logger(logger),
       kratos.Server(
          gs,
          hs,
       ),
       kratos.Registrar(reg),
    )
}

func main() {
    flag.Parse()
    logger := log.With(log.NewStdLogger(os.Stdout),
       "ts", log.DefaultTimestamp,
       "caller", log.DefaultCaller,
       "service.id", id,
       "service.name", Name,
       "service.version", Version,
       "trace.id", tracing.TraceID(),
       "span.id", tracing.SpanID(),
    )
    c := config.New(
       config.WithSource(
          file.NewSource(flagconf),
       ),
    )
    defer c.Close()

    if err := c.Load(); err != nil {
       panic(err)
    }

    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
       panic(err)
    }

    app, cleanup, err := wireApp(bc.Service, bc.Server, bc.Data, logger)
    if err != nil {
       panic(err)
    }
    defer cleanup()

    // start and wait for stop signal
    if err := app.Run(); err != nil {
       panic(err)
    }
}

// 初始化consul服务注册
func initServiceRegistry(address string) (*consul.Registry, error) {
    // 一，获取consul客户端
    consulConfig := api.DefaultConfig()
    consulConfig.Address = address
    consulClient, err := api.NewClient(consulConfig)
    if err != nil {
       return nil, err
    }
    // 二，获取consul注册管理器
    reg := consul.New(consulClient)
    return reg, nil
}

// 初始化Tracer
// @param url string 指定 Jaeger 的API接口
func initTracer(url string) error {
    //一，建立jaeger客户端，称之为：exporter，导出器
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
    if err != nil {
       return err
    }
    // 创建 TracerProvider
    tracerProvider := trace.NewTracerProvider(
       //采样设置
       trace.WithSampler(trace.AlwaysSample()),
       // 使用 exporter 作为批处理程序
       trace.WithBatcher(exporter),
       // 将当前服务的信息，作为资源告知给TracerProvider
       trace.WithResource(resource.NewSchemaless(
          // 必要的配置
          semconv.ServiceNameKey.String(Name),
          // 任意的其他属性配置
          attribute.String("exporter", "jaeger"),
       )),
    )
    // 三，设置全局的TP
    otel.SetTracerProvider(tracerProvider)
    return nil
}
```

1. **配置司机服务：**

   ```go
   var (
       // Name is the name of the compiled software.
       Name string = "Driver"
       // Version is the version of the compiled software.
       Version string = "1.0.0"
       // flagconf is the config flag.
       flagconf string
       // id, _ = os.Hostname()
       id = Name + "-" + uuid.NewString()
   )
   ```

2. **初始化consul服务注册中心**：合并了新建consul客户端、新建consul服务中心，优化了consul地址的获取（使用cs.Consul.Address作为参数）。

   ```go
   reg, err := initServiceRegistry(cs.Consul.Address)
   if err != nil {
       panic(err)
   }
   ```

3. **初始化链路追踪**：但是jaeger链路追踪在本服务中暂时没用到。

   ```go
   // 链路追踪
   if err := initTracer(cs.Jaeger.Url); err != nil {
       panic(err)
   }
   ```

4. **创建服务**：kratos.Registrar(reg)；

### 依赖注入

3个位置，servce层、biz层、data层，然后执行依赖注入命令go generate ./...；

```go
// service.go中
var ProviderSet = wire.NewSet(NewGreeterService, NewDriverService)
// biz.go中
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewDriverBiz)
// data.go中
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewDriverInterface)
```

### 启动服务

```shell
kratos run
```



 
