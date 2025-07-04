package auth

import (
	"context"
	"encoding/base64"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type Checker interface {
	Check(user, pass string) bool
}

func NewUnaryBasicAuth(c Checker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, unauth()
		}
		auth := md.Get("authorization")
		if len(auth) == 0 || !strings.HasPrefix(auth[0], "Basic ") {
			return nil, unauth()
		}
		decoded, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth[0], "Basic "))
		cred := strings.SplitN(string(decoded), ":", 2)
		if len(cred) != 2 || !c.Check(cred[0], cred[1]) {
			return nil, unauth()
		}
		return handler(ctx, req)
	}
}

func NewUnaryBasicAuthWithFilter(
	c Checker,
	needAuth func(fullMethod string) bool,
) grpc.UnaryServerInterceptor {

	// базовый интерсептор, который всегда проверяет Basic-Auth
	base := NewUnaryBasicAuth(c)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		// если метод НЕ требует авторизации - пропускаем сразу
		if !needAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// иначе делаем обычную проверку Basic-Auth
		return base(ctx, req, info, handler)
	}
}

func unauth() error {
	return status.Error(codes.Unauthenticated, "unauthenticated")
}

type StaticCreds struct {
	User string
	Pass string
}

func (s StaticCreds) Check(u, p string) bool {
	return u == s.User && p == s.Pass
}
