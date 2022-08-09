package grpc

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/gabihodoroaga/http-grpc-websocket/config"
)

// AuthInterceptor is a server interceptor for authentication and authorization
type authInterceptor struct {
	allowedUsers []string
}

// NewAuthInterceptor returns a new auth interceptor
func newAuthInterceptor(allowedUsers []string) *authInterceptor {
	return &authInterceptor{allowedUsers: allowedUsers}
}

func (interceptor *authInterceptor) unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		zap.L().Sugar().Debugf("grpc/server/auth-unary: authorizing request %s", info.FullMethod)
		err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (interceptor *authInterceptor) authorize(ctx context.Context, method string) error {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata required for authorization is not provided")
	}

	values := md["authorization"]
	zap.L().Sugar().Debugf("grpc/server/auth-authorize: received authorization header len=%i, values=%v", len(values), values)
	if len(values) == 0 {
		zap.L().Sugar().Warnf("authorization header is not provided or has invalid format. Expected: Bearer [token], found %s", values)
		return status.Errorf(codes.Unauthenticated, "authorization header is not provided or has invalid format. Expected: Bearer [token].")
	}
	authHeader := values[0]
	prefix := "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		zap.L().Sugar().Warnf("bearer prefix not found in authorization header. Expected: Bearer [token] found %s.", authHeader)
		return status.Errorf(codes.Unauthenticated, "bearer prefix not found in authorization header. Expected: Bearer [token]")
	}

	token := authHeader[strings.Index(authHeader, prefix)+len(prefix):]
	if token == "" {
		zap.L().Sugar().Warnf("grpc/server/auth-authorize: not a valid jwt token. Expected: Bearer [token] found %s.", authHeader)
		return status.Errorf(codes.Unauthenticated, "not a valid jwt token. Expected: Bearer [token]")
	}

	payload, err := idtoken.Validate(ctx, token, config.GetConfig().AuthAudience)
	if err != nil {
		zap.L().Sugar().Warnf("grpc/server/auth-authorize: token validation failed %v", err)
		return status.Errorf(codes.Unauthenticated, "invalid token is invalid: %v", err)
	}

	emailClaim, found := payload.Claims["email"]
	if !found {
		zap.L().Sugar().Warn("grpc/server/auth-authorize: token validation failed: email claim not found")
		return status.Error(codes.Unauthenticated, "invalid token: email claim not found")
	}

	email, ok := emailClaim.(string)
	if !ok {
		zap.L().Sugar().Warnf("grpc/server/auth-authorize: token validation failed: invalid email claim type, expected string found %T", emailClaim)
		return status.Errorf(codes.Unauthenticated, "invalid token: invalid email claim type, expected string found %T", emailClaim)
	}

	if !sliceContainsString(interceptor.allowedUsers, email) {
		return status.Error(codes.PermissionDenied, "no permission to access service")
	}

	return nil
}

func sliceContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
