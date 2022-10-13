package handler

import (
	"context"
	"fmt"
)

// UnaryHandler is the handler that handles unary operations
type UnaryHandler = func(ctx context.Context, req interface{}) (interface{}, error)

// Interceptor is the interceptor that handles any incoming request
type Interceptor = func(ctx context.Context, req interface{},
	info *UnaryServerInfo, handler UnaryHandler) (interface{}, error)

// UnaryServerInfo is the server info
type UnaryServerInfo struct {
	// FullMethod is the method name
	FullMethod string
}

// Handler is the handler that handles unary operations
type Handler struct {
	// inter is the name of the interceptor
	inter Interceptor
	// methodHandler is the name of multiple handlers
	methodHandler map[string]UnaryHandler
}

// New creates a new handler
func New(methodHandler map[string]UnaryHandler, interceptors ...Interceptor) (*Handler, error) {
	return &Handler{
		inter:         MultiInterceptors(interceptors...),
		methodHandler: methodHandler,
	}, nil
}

// Handler is used to handler events
func (s *Handler) Handler(ctx context.Context, req interface{}, info *UnaryServerInfo) (interface{}, error) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		h, ok := s.methodHandler[info.FullMethod]
		if !ok {
			return nil, fmt.Errorf("no method")
		}
		return h(ctx, req)
	}
	return s.inter(ctx, req, info, handler)
}

// MultiInterceptors is used to handle multiple interceptors
func MultiInterceptors(interceptors ...Interceptor) Interceptor {
	n := len(interceptors)
	return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
		chainer := func(currentInter Interceptor, currentHandler UnaryHandler) UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInter(currentCtx, currentReq, info, currentHandler)
			}
		}
		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}
		return chainedHandler(ctx, req)
	}
}
