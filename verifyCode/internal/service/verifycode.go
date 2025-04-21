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
