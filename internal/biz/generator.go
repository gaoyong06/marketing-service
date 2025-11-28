package biz

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// RandomCodeGenerator 随机码生成器实现
type RandomCodeGenerator struct {
	mu  sync.Mutex
	log *log.Helper
}

// NewCodeGenerator 创建码生成器
func NewCodeGenerator(logger log.Logger) CodeGenerator {
	return &RandomCodeGenerator{
		log: log.NewHelper(logger),
	}
}

// Generate 生成唯一码
func (g *RandomCodeGenerator) Generate(rule CodeGenerateRule) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var result string

	switch rule.CodeType {
	case "RANDOM_NUM":
		result = g.generateRandomNum(rule.Length)
	case "ALPHANUM":
		result = g.generateAlphaNum(rule.Length, rule.ExcludeChars)
	case "UUID":
		result = g.generateUUID()
	default:
		result = g.generateAlphaNum(rule.Length, rule.ExcludeChars)
	}

	// 添加前缀
	if rule.Prefix != "" {
		result = rule.Prefix + result
	}

	return result, nil
}

// generateRandomNum 生成纯数字随机码
func (g *RandomCodeGenerator) generateRandomNum(length int) string {
	if length <= 0 {
		length = 8
	}

	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		sb.WriteByte(byte(num.Int64()) + '0')
	}

	return sb.String()
}

// generateAlphaNum 生成字母数字混合随机码
func (g *RandomCodeGenerator) generateAlphaNum(length int, excludeChars []string) string {
	if length <= 0 {
		length = 8
	}

	const (
		digits   = "0123456789"
		alphaUpp = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		alphaLow = "abcdefghijklmnopqrstuvwxyz"
	)

	// 构建字符池
	charPool := digits + alphaUpp

	// 排除指定字符
	for _, c := range excludeChars {
		charPool = strings.ReplaceAll(charPool, c, "")
	}

	var sb strings.Builder
	sb.Grow(length)

	poolLen := big.NewInt(int64(len(charPool)))
	for i := 0; i < length; i++ {
		index, _ := rand.Int(rand.Reader, poolLen)
		sb.WriteByte(charPool[index.Int64()])
	}

	return sb.String()
}

// generateUUID 生成UUID
func (g *RandomCodeGenerator) generateUUID() string {
	uuid := uuid.New().String()
	return strings.ReplaceAll(uuid, "-", "")[:12] // 取12位，去除连字符
}
