// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: paymentpreference.proto

/*
Package paymentpreference is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package paymentpreference

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

var (
	filter_PaymentPreferenceService_GetPaymentPreference_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_PaymentPreferenceService_GetPaymentPreference_0(ctx context.Context, marshaler runtime.Marshaler, client PaymentPreferenceServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq PaymentPreferenceRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_PaymentPreferenceService_GetPaymentPreference_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPaymentPreference(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_PaymentPreferenceService_GetPaymentPreference_0(ctx context.Context, marshaler runtime.Marshaler, server PaymentPreferenceServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq PaymentPreferenceRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_PaymentPreferenceService_GetPaymentPreference_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetPaymentPreference(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterPaymentPreferenceServiceHandlerServer registers the http handlers for service PaymentPreferenceService to "mux".
// UnaryRPC     :call PaymentPreferenceServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterPaymentPreferenceServiceHandlerFromEndpoint instead.
func RegisterPaymentPreferenceServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server PaymentPreferenceServiceServer) error {

	mux.Handle("GET", pattern_PaymentPreferenceService_GetPaymentPreference_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/paymentpreference.PaymentPreferenceService/GetPaymentPreference", runtime.WithHTTPPathPattern("/paymentpreference/GetPaymentPreference"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_PaymentPreferenceService_GetPaymentPreference_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PaymentPreferenceService_GetPaymentPreference_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterPaymentPreferenceServiceHandlerFromEndpoint is same as RegisterPaymentPreferenceServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterPaymentPreferenceServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterPaymentPreferenceServiceHandler(ctx, mux, conn)
}

// RegisterPaymentPreferenceServiceHandler registers the http handlers for service PaymentPreferenceService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterPaymentPreferenceServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterPaymentPreferenceServiceHandlerClient(ctx, mux, NewPaymentPreferenceServiceClient(conn))
}

// RegisterPaymentPreferenceServiceHandlerClient registers the http handlers for service PaymentPreferenceService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "PaymentPreferenceServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "PaymentPreferenceServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "PaymentPreferenceServiceClient" to call the correct interceptors.
func RegisterPaymentPreferenceServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client PaymentPreferenceServiceClient) error {

	mux.Handle("GET", pattern_PaymentPreferenceService_GetPaymentPreference_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/paymentpreference.PaymentPreferenceService/GetPaymentPreference", runtime.WithHTTPPathPattern("/paymentpreference/GetPaymentPreference"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_PaymentPreferenceService_GetPaymentPreference_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PaymentPreferenceService_GetPaymentPreference_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_PaymentPreferenceService_GetPaymentPreference_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"paymentpreference", "GetPaymentPreference"}, ""))
)

var (
	forward_PaymentPreferenceService_GetPaymentPreference_0 = runtime.ForwardResponseMessage
)