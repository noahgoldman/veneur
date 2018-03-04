// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: forwardrpc/forward.proto

/*
	Package forwardrpc is a generated protocol buffer package.

	It is generated from these files:
		forwardrpc/forward.proto

	It has these top-level messages:
		MetricList
*/
package forwardrpc

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import metricpb "github.com/stripe/veneur/samplers/metricpb"
import google_protobuf1 "github.com/golang/protobuf/ptypes/empty"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type MetricList struct {
	Metrics []*metricpb.Metric `protobuf:"bytes,1,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *MetricList) Reset()                    { *m = MetricList{} }
func (m *MetricList) String() string            { return proto.CompactTextString(m) }
func (*MetricList) ProtoMessage()               {}
func (*MetricList) Descriptor() ([]byte, []int) { return fileDescriptorForward, []int{0} }

func (m *MetricList) GetMetrics() []*metricpb.Metric {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func init() {
	proto.RegisterType((*MetricList)(nil), "forwardrpc.MetricList")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Forward service

type ForwardClient interface {
	SendMetrics(ctx context.Context, in *MetricList, opts ...grpc.CallOption) (*google_protobuf1.Empty, error)
}

type forwardClient struct {
	cc *grpc.ClientConn
}

func NewForwardClient(cc *grpc.ClientConn) ForwardClient {
	return &forwardClient{cc}
}

func (c *forwardClient) SendMetrics(ctx context.Context, in *MetricList, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := grpc.Invoke(ctx, "/forwardrpc.Forward/SendMetrics", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Forward service

type ForwardServer interface {
	SendMetrics(context.Context, *MetricList) (*google_protobuf1.Empty, error)
}

func RegisterForwardServer(s *grpc.Server, srv ForwardServer) {
	s.RegisterService(&_Forward_serviceDesc, srv)
}

func _Forward_SendMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricList)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForwardServer).SendMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/forwardrpc.Forward/SendMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForwardServer).SendMetrics(ctx, req.(*MetricList))
	}
	return interceptor(ctx, in, info, handler)
}

var _Forward_serviceDesc = grpc.ServiceDesc{
	ServiceName: "forwardrpc.Forward",
	HandlerType: (*ForwardServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMetrics",
			Handler:    _Forward_SendMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "forwardrpc/forward.proto",
}

func (m *MetricList) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricList) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, msg := range m.Metrics {
			dAtA[i] = 0xa
			i++
			i = encodeVarintForward(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeVarintForward(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *MetricList) Size() (n int) {
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, e := range m.Metrics {
			l = e.Size()
			n += 1 + l + sovForward(uint64(l))
		}
	}
	return n
}

func sovForward(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozForward(x uint64) (n int) {
	return sovForward(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MetricList) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowForward
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MetricList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metrics", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowForward
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthForward
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, &metricpb.Metric{})
			if err := m.Metrics[len(m.Metrics)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipForward(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthForward
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipForward(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowForward
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowForward
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowForward
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthForward
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowForward
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipForward(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthForward = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowForward   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("forwardrpc/forward.proto", fileDescriptorForward) }

var fileDescriptorForward = []byte{
	// 234 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x48, 0xcb, 0x2f, 0x2a,
	0x4f, 0x2c, 0x4a, 0x29, 0x2a, 0x48, 0xd6, 0x87, 0x32, 0xf5, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85,
	0xb8, 0x10, 0x32, 0x52, 0xe6, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa,
	0xc5, 0x25, 0x45, 0x99, 0x05, 0xa9, 0xfa, 0x65, 0xa9, 0x79, 0xa9, 0xa5, 0x45, 0xfa, 0xc5, 0x89,
	0xb9, 0x05, 0x39, 0xa9, 0x45, 0xc5, 0xfa, 0xb9, 0xa9, 0x25, 0x45, 0x99, 0xc9, 0x05, 0x49, 0x50,
	0x06, 0xc4, 0x10, 0x29, 0x63, 0x24, 0x8d, 0xe9, 0xf9, 0x39, 0x89, 0x79, 0xe9, 0xfa, 0x60, 0x89,
	0xa4, 0xd2, 0x34, 0xfd, 0x82, 0x92, 0xca, 0x82, 0xd4, 0x62, 0xfd, 0xd4, 0xdc, 0x82, 0x92, 0x4a,
	0x08, 0x09, 0xd1, 0xa4, 0x64, 0xc1, 0xc5, 0xe5, 0x0b, 0x36, 0xc4, 0x27, 0xb3, 0xb8, 0x44, 0x48,
	0x8b, 0x8b, 0x1d, 0x62, 0x64, 0xb1, 0x04, 0xa3, 0x02, 0xb3, 0x06, 0xb7, 0x91, 0x80, 0x1e, 0xcc,
	0x2e, 0x3d, 0x88, 0xb2, 0x20, 0x98, 0x02, 0x23, 0x2f, 0x2e, 0x76, 0x37, 0x88, 0xab, 0x85, 0xec,
	0xb9, 0xb8, 0x83, 0x53, 0xf3, 0x52, 0x20, 0x2a, 0x8a, 0x85, 0xc4, 0xf4, 0x10, 0xde, 0xd1, 0x43,
	0x98, 0x2e, 0x25, 0xa6, 0x97, 0x9e, 0x9f, 0x9f, 0x9e, 0x93, 0xaa, 0x07, 0x73, 0x96, 0x9e, 0x2b,
	0xc8, 0x25, 0x4a, 0x0c, 0x4e, 0x02, 0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0,
	0x91, 0x1c, 0xe3, 0x84, 0xc7, 0x72, 0x0c, 0x49, 0x6c, 0x60, 0x35, 0xc6, 0x80, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x4b, 0x84, 0x21, 0xf9, 0x34, 0x01, 0x00, 0x00,
}
