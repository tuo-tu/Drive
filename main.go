package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func main() {
	const (
		DEFAULT = iota
		DIGIT
		LETTER
		MIXED
	)
	fmt.Println(RandCode(7, DEFAULT))
}

func RandCode(l int, t int) string {
	switch t {
	case 0:
		fallthrough
	case 1:
		// idxBits表示使用4位二进制数就可以表示完chars的索引了
		return randCode3("0123456789", l, 4)
	case 2:
		return randCode3("abcdefghijklmnopqrstuvwxyz", l, 5)
	case 3:
		return randCode3("0123456789abcdefghijklmnopqrstuvwxyz", l, 6)
	default:
	}
	return ""
}

// 随机数的核心方法（优化实现）
// 一次随机多个随机位，分部分多次使用，
// idxBits表示使用4位二进制数就可以表示完chars的索引了

func randCode1(chars string, l, idxBits int) string {
	// 计算有效的二进制数位，基于 chars 的长度
	// 推荐写死，因为chars固定，对应的位数长度也固定
	// 形成掩码，mask
	// 例如，使用低idxBits位：00000000000111111
	idxMask := 1<<idxBits - 1 // 00001000000 - 1 = 00000111111
	// 63 位可以用多少次
	idxMax := 63 / idxBits
	// 利用string builder构建结果缓冲，高效拼接字符串
	sb := strings.Builder{}
	sb.Grow(l)

	// 生成随机字符,cache:随机位缓存 ;remain:当前还可以用几次
	for i, cache, remain := 0, rand.Int63(), idxMax; i < l; {
		// 如果使用的剩余次数为0，则重新获取随机
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		// 利用掩码获取cache的低位作为randIndex（索引）
		if idx := int(cache & int64(idxMask)); idx < len(chars) {
			sb.WriteByte(chars[idx])
			i--
		}
		// 使用下一组随机位。右移会丢掉先前的低位，高位补0
		cache >>= idxBits
		remain--
	}
	return sb.String()
}

func randCode2(chars string, idxBits, l int) string {
	// 形成掩码
	idxMask := 1<<idxBits - 1
	// 63 位可以使用的最大组次数
	idxMax := 63 / idxBits

	// 利用string builder构建结果缓冲
	sb := strings.Builder{}
	sb.Grow(l)

	// 循环生成随机数
	// i 索引
	// cache 随机数缓存
	// remain 随机数还可以用几次
	for i, cache, remain := l-1, rand.Int63(), idxMax; i >= 0; {
		// 随机缓存不足，重新生成
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		// 利用掩码生成随机索引，有效索引为小于字符集合长度
		if idx := int(cache & int64(idxMask)); idx < len(chars) {
			sb.WriteByte(chars[idx])
			i--
		}
		// 利用下一组随机数位
		cache >>= idxBits
		remain--
	}

	return sb.String()
}

func randCode3(chars string, l, idxBits int) string {
	// 计算有效的二进制数位，基于 chars 的长度
	// 推荐写死，因为chars固定，对应的位数长度也固定
	// 形成掩码，mask
	// 例如，使用低idxBits位：00000000000111111
	idxMask := 1<<idxBits - 1 // 00001000000 - 1 = 00000111111
	// 63 位可以用多少次
	idxMax := 63 / idxBits

	// 利用string builder构建结果缓冲
	sb := strings.Builder{}
	sb.Grow(l)
	//result := make([]byte, l)
	// 生成随机字符,cache:随机位缓存 ;remain:当前还可以用几次
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
