package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"
)

type TokenAuth struct {
	T int64  // 时间戳
	S string // 签名
}

func (t *TokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"t": strconv.FormatInt(t.T, 10),
		"s": t.S,
	}, nil
}

// RequireTransportSecurity 是否强制使用TLS
func (t *TokenAuth) RequireTransportSecurity() bool {
	return false
}

// Verify 服务端校验
func (t *TokenAuth) Verify(token string) bool {
	tm := time.Unix(t.T, 0)
	if time.Now().After(tm.Add(300 * time.Second)) { // 时间超过了5分钟
		return false
	}
	return t.S == t.Sign(token)
}

func (t *TokenAuth) Sign(token string) string {
	sign := sha256.Sum256([]byte(fmt.Sprintf("%d%s", t.T, token)))
	return hex.EncodeToString(sign[:])
}

func getOne(vals []string) string {
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// ParseMap 从GRPC的meta信息中获取tokenAuth结构
func (t *TokenAuth) ParseMap(meta metadata.MD) *TokenAuth {
	t.T, _ = strconv.ParseInt(getOne(meta["t"]), 10, 64)
	t.S = getOne(meta["s"])
	return t
}
