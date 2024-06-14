// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: pkg/pb/kolide-event-handler.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Health int32

const (
	Health_Unknown   Health = 0
	Health_Healthy   Health = 1
	Health_Unhealthy Health = 2
)

// Enum value maps for Health.
var (
	Health_name = map[int32]string{
		0: "Unknown",
		1: "Healthy",
		2: "Unhealthy",
	}
	Health_value = map[string]int32{
		"Unknown":   0,
		"Healthy":   1,
		"Unhealthy": 2,
	}
)

func (x Health) Enum() *Health {
	p := new(Health)
	*p = x
	return p
}

func (x Health) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Health) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_pb_kolide_event_handler_proto_enumTypes[0].Descriptor()
}

func (Health) Type() protoreflect.EnumType {
	return &file_pkg_pb_kolide_event_handler_proto_enumTypes[0]
}

func (x Health) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Health.Descriptor instead.
func (Health) EnumDescriptor() ([]byte, []int) {
	return file_pkg_pb_kolide_event_handler_proto_rawDescGZIP(), []int{0}
}

type EventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *EventsRequest) Reset() {
	*x = EventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_pb_kolide_event_handler_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventsRequest) ProtoMessage() {}

func (x *EventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_pb_kolide_event_handler_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventsRequest.ProtoReflect.Descriptor instead.
func (*EventsRequest) Descriptor() ([]byte, []int) {
	return file_pkg_pb_kolide_event_handler_proto_rawDescGZIP(), []int{0}
}

type DeviceEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp  *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Serial     string                 `protobuf:"bytes,2,opt,name=serial,proto3" json:"serial,omitempty"`
	Platform   string                 `protobuf:"bytes,3,opt,name=platform,proto3" json:"platform,omitempty"`
	State      Health                 `protobuf:"varint,4,opt,name=state,proto3,enum=kolide_event_handler.Health" json:"state,omitempty"`
	Message    string                 `protobuf:"bytes,5,opt,name=message,proto3" json:"message,omitempty"`
	ExternalID string                 `protobuf:"bytes,6,opt,name=externalID,proto3" json:"externalID,omitempty"`
}

func (x *DeviceEvent) Reset() {
	*x = DeviceEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_pb_kolide_event_handler_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceEvent) ProtoMessage() {}

func (x *DeviceEvent) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_pb_kolide_event_handler_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceEvent.ProtoReflect.Descriptor instead.
func (*DeviceEvent) Descriptor() ([]byte, []int) {
	return file_pkg_pb_kolide_event_handler_proto_rawDescGZIP(), []int{1}
}

func (x *DeviceEvent) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *DeviceEvent) GetSerial() string {
	if x != nil {
		return x.Serial
	}
	return ""
}

func (x *DeviceEvent) GetPlatform() string {
	if x != nil {
		return x.Platform
	}
	return ""
}

func (x *DeviceEvent) GetState() Health {
	if x != nil {
		return x.State
	}
	return Health_Unknown
}

func (x *DeviceEvent) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *DeviceEvent) GetExternalID() string {
	if x != nil {
		return x.ExternalID
	}
	return ""
}

var File_pkg_pb_kolide_event_handler_proto protoreflect.FileDescriptor

var file_pkg_pb_kolide_event_handler_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x6b, 0x6f, 0x6c, 0x69, 0x64, 0x65, 0x2d,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x2d, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x14, 0x6b, 0x6f, 0x6c, 0x69, 0x64, 0x65, 0x5f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x0f, 0x0a, 0x0d, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xe9, 0x01, 0x0a, 0x0b,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x32, 0x0a, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x6b, 0x6f, 0x6c, 0x69, 0x64,
	0x65, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2e,
	0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x49, 0x44, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x78, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x49, 0x44, 0x2a, 0x31, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00, 0x12, 0x0b,
	0x0a, 0x07, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x55,
	0x6e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x10, 0x02, 0x32, 0x6a, 0x0a, 0x12, 0x4b, 0x6f,
	0x6c, 0x69, 0x64, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72,
	0x12, 0x54, 0x0a, 0x06, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x23, 0x2e, 0x6b, 0x6f, 0x6c,
	0x69, 0x64, 0x65, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x72, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x21, 0x2e, 0x6b, 0x6f, 0x6c, 0x69, 0x64, 0x65, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x68,
	0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x22, 0x00, 0x30, 0x01, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x61, 0x69, 0x73, 0x2f, 0x6b, 0x6f, 0x6c, 0x69, 0x64, 0x65,
	0x2d, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2d, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_pb_kolide_event_handler_proto_rawDescOnce sync.Once
	file_pkg_pb_kolide_event_handler_proto_rawDescData = file_pkg_pb_kolide_event_handler_proto_rawDesc
)

func file_pkg_pb_kolide_event_handler_proto_rawDescGZIP() []byte {
	file_pkg_pb_kolide_event_handler_proto_rawDescOnce.Do(func() {
		file_pkg_pb_kolide_event_handler_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_pb_kolide_event_handler_proto_rawDescData)
	})
	return file_pkg_pb_kolide_event_handler_proto_rawDescData
}

var file_pkg_pb_kolide_event_handler_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_pb_kolide_event_handler_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_pb_kolide_event_handler_proto_goTypes = []interface{}{
	(Health)(0),                   // 0: kolide_event_handler.Health
	(*EventsRequest)(nil),         // 1: kolide_event_handler.EventsRequest
	(*DeviceEvent)(nil),           // 2: kolide_event_handler.DeviceEvent
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_pkg_pb_kolide_event_handler_proto_depIdxs = []int32{
	3, // 0: kolide_event_handler.DeviceEvent.timestamp:type_name -> google.protobuf.Timestamp
	0, // 1: kolide_event_handler.DeviceEvent.state:type_name -> kolide_event_handler.Health
	1, // 2: kolide_event_handler.KolideEventHandler.Events:input_type -> kolide_event_handler.EventsRequest
	2, // 3: kolide_event_handler.KolideEventHandler.Events:output_type -> kolide_event_handler.DeviceEvent
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_pb_kolide_event_handler_proto_init() }
func file_pkg_pb_kolide_event_handler_proto_init() {
	if File_pkg_pb_kolide_event_handler_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_pb_kolide_event_handler_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_pb_kolide_event_handler_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_pb_kolide_event_handler_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_pb_kolide_event_handler_proto_goTypes,
		DependencyIndexes: file_pkg_pb_kolide_event_handler_proto_depIdxs,
		EnumInfos:         file_pkg_pb_kolide_event_handler_proto_enumTypes,
		MessageInfos:      file_pkg_pb_kolide_event_handler_proto_msgTypes,
	}.Build()
	File_pkg_pb_kolide_event_handler_proto = out.File
	file_pkg_pb_kolide_event_handler_proto_rawDesc = nil
	file_pkg_pb_kolide_event_handler_proto_goTypes = nil
	file_pkg_pb_kolide_event_handler_proto_depIdxs = nil
}
