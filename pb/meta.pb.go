// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: meta.proto

package pb

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// All title types in the title.basics.tsv.gz dataset as of 2020-11-21.
type TitleType int32

const (
	TitleType_MOVIE          TitleType = 0
	TitleType_SHORT          TitleType = 1
	TitleType_TV_EPISODE     TitleType = 2
	TitleType_TV_MINI_SERIES TitleType = 3
	TitleType_TV_MOVIE       TitleType = 4
	TitleType_TV_SERIES      TitleType = 5
	TitleType_TV_SHORT       TitleType = 6
	TitleType_TV_SPECIAL     TitleType = 7
	TitleType_VIDEO          TitleType = 8
	TitleType_VIDEO_GAME     TitleType = 9
)

// Enum value maps for TitleType.
var (
	TitleType_name = map[int32]string{
		0: "MOVIE",
		1: "SHORT",
		2: "TV_EPISODE",
		3: "TV_MINI_SERIES",
		4: "TV_MOVIE",
		5: "TV_SERIES",
		6: "TV_SHORT",
		7: "TV_SPECIAL",
		8: "VIDEO",
		9: "VIDEO_GAME",
	}
	TitleType_value = map[string]int32{
		"MOVIE":          0,
		"SHORT":          1,
		"TV_EPISODE":     2,
		"TV_MINI_SERIES": 3,
		"TV_MOVIE":       4,
		"TV_SERIES":      5,
		"TV_SHORT":       6,
		"TV_SPECIAL":     7,
		"VIDEO":          8,
		"VIDEO_GAME":     9,
	}
)

func (x TitleType) Enum() *TitleType {
	p := new(TitleType)
	*p = x
	return p
}

func (x TitleType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TitleType) Descriptor() protoreflect.EnumDescriptor {
	return file_meta_proto_enumTypes[0].Descriptor()
}

func (TitleType) Type() protoreflect.EnumType {
	return &file_meta_proto_enumTypes[0]
}

func (x TitleType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TitleType.Descriptor instead.
func (TitleType) EnumDescriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{0}
}

type Meta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            string    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // IMDb ID, including "tt" prefix
	TitleType     TitleType `protobuf:"varint,2,opt,name=title_type,json=titleType,proto3,enum=imdb2meta.TitleType" json:"title_type,omitempty"`
	PrimaryTitle  string    `protobuf:"bytes,3,opt,name=primary_title,json=primaryTitle,proto3" json:"primary_title,omitempty"`
	OriginalTitle string    `protobuf:"bytes,4,opt,name=original_title,json=originalTitle,proto3" json:"original_title,omitempty"` // Only filled if different from the primary title
	IsAdult       bool      `protobuf:"varint,5,opt,name=is_adult,json=isAdult,proto3" json:"is_adult,omitempty"`
	StartYear     int32     `protobuf:"varint,6,opt,name=start_year,json=startYear,proto3" json:"start_year,omitempty"` // Start year for TV shows, release year for movies. Can be 0.
	EndYear       int32     `protobuf:"varint,7,opt,name=end_year,json=endYear,proto3" json:"end_year,omitempty"`       // Only relevant for TV shows
	Runtime       int32     `protobuf:"varint,8,opt,name=runtime,proto3" json:"runtime,omitempty"`                      // In minutes. Can be 0.
	Genres        []string  `protobuf:"bytes,9,rep,name=genres,proto3" json:"genres,omitempty"`                         // Up to three genres. Can be empty.
}

func (x *Meta) Reset() {
	*x = Meta{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Meta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Meta) ProtoMessage() {}

func (x *Meta) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Meta.ProtoReflect.Descriptor instead.
func (*Meta) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{0}
}

func (x *Meta) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Meta) GetTitleType() TitleType {
	if x != nil {
		return x.TitleType
	}
	return TitleType_MOVIE
}

func (x *Meta) GetPrimaryTitle() string {
	if x != nil {
		return x.PrimaryTitle
	}
	return ""
}

func (x *Meta) GetOriginalTitle() string {
	if x != nil {
		return x.OriginalTitle
	}
	return ""
}

func (x *Meta) GetIsAdult() bool {
	if x != nil {
		return x.IsAdult
	}
	return false
}

func (x *Meta) GetStartYear() int32 {
	if x != nil {
		return x.StartYear
	}
	return 0
}

func (x *Meta) GetEndYear() int32 {
	if x != nil {
		return x.EndYear
	}
	return 0
}

func (x *Meta) GetRuntime() int32 {
	if x != nil {
		return x.Runtime
	}
	return 0
}

func (x *Meta) GetGenres() []string {
	if x != nil {
		return x.Genres
	}
	return nil
}

var File_meta_proto protoreflect.FileDescriptor

var file_meta_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6d, 0x65, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x69, 0x6d,
	0x64, 0x62, 0x32, 0x6d, 0x65, 0x74, 0x61, 0x22, 0x9e, 0x02, 0x0a, 0x04, 0x4d, 0x65, 0x74, 0x61,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x33, 0x0a, 0x0a, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x69, 0x6d, 0x64, 0x62, 0x32, 0x6d, 0x65, 0x74, 0x61,
	0x2e, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x69, 0x6d, 0x61, 0x72, 0x79,
	0x5f, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x72,
	0x69, 0x6d, 0x61, 0x72, 0x79, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x6f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x54, 0x69, 0x74, 0x6c,
	0x65, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x73, 0x5f, 0x61, 0x64, 0x75, 0x6c, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x73, 0x41, 0x64, 0x75, 0x6c, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x79, 0x65, 0x61, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x59, 0x65, 0x61, 0x72, 0x12, 0x19, 0x0a, 0x08, 0x65,
	0x6e, 0x64, 0x5f, 0x79, 0x65, 0x61, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65,
	0x6e, 0x64, 0x59, 0x65, 0x61, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x67, 0x65, 0x6e, 0x72, 0x65, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x06, 0x67, 0x65, 0x6e, 0x72, 0x65, 0x73, 0x2a, 0x9b, 0x01, 0x0a, 0x09, 0x54, 0x69, 0x74,
	0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x4d, 0x4f, 0x56, 0x49, 0x45, 0x10,
	0x00, 0x12, 0x09, 0x0a, 0x05, 0x53, 0x48, 0x4f, 0x52, 0x54, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a,
	0x54, 0x56, 0x5f, 0x45, 0x50, 0x49, 0x53, 0x4f, 0x44, 0x45, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e,
	0x54, 0x56, 0x5f, 0x4d, 0x49, 0x4e, 0x49, 0x5f, 0x53, 0x45, 0x52, 0x49, 0x45, 0x53, 0x10, 0x03,
	0x12, 0x0c, 0x0a, 0x08, 0x54, 0x56, 0x5f, 0x4d, 0x4f, 0x56, 0x49, 0x45, 0x10, 0x04, 0x12, 0x0d,
	0x0a, 0x09, 0x54, 0x56, 0x5f, 0x53, 0x45, 0x52, 0x49, 0x45, 0x53, 0x10, 0x05, 0x12, 0x0c, 0x0a,
	0x08, 0x54, 0x56, 0x5f, 0x53, 0x48, 0x4f, 0x52, 0x54, 0x10, 0x06, 0x12, 0x0e, 0x0a, 0x0a, 0x54,
	0x56, 0x5f, 0x53, 0x50, 0x45, 0x43, 0x49, 0x41, 0x4c, 0x10, 0x07, 0x12, 0x09, 0x0a, 0x05, 0x56,
	0x49, 0x44, 0x45, 0x4f, 0x10, 0x08, 0x12, 0x0e, 0x0a, 0x0a, 0x56, 0x49, 0x44, 0x45, 0x4f, 0x5f,
	0x47, 0x41, 0x4d, 0x45, 0x10, 0x09, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x66, 0x6c, 0x69, 0x78, 0x2d, 0x74, 0x76, 0x2f, 0x69,
	0x6d, 0x64, 0x62, 0x32, 0x6d, 0x65, 0x74, 0x61, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_meta_proto_rawDescOnce sync.Once
	file_meta_proto_rawDescData = file_meta_proto_rawDesc
)

func file_meta_proto_rawDescGZIP() []byte {
	file_meta_proto_rawDescOnce.Do(func() {
		file_meta_proto_rawDescData = protoimpl.X.CompressGZIP(file_meta_proto_rawDescData)
	})
	return file_meta_proto_rawDescData
}

var file_meta_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_meta_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_meta_proto_goTypes = []interface{}{
	(TitleType)(0), // 0: imdb2meta.TitleType
	(*Meta)(nil),   // 1: imdb2meta.Meta
}
var file_meta_proto_depIdxs = []int32{
	0, // 0: imdb2meta.Meta.title_type:type_name -> imdb2meta.TitleType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_meta_proto_init() }
func file_meta_proto_init() {
	if File_meta_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_meta_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Meta); i {
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
			RawDescriptor: file_meta_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_meta_proto_goTypes,
		DependencyIndexes: file_meta_proto_depIdxs,
		EnumInfos:         file_meta_proto_enumTypes,
		MessageInfos:      file_meta_proto_msgTypes,
	}.Build()
	File_meta_proto = out.File
	file_meta_proto_rawDesc = nil
	file_meta_proto_goTypes = nil
	file_meta_proto_depIdxs = nil
}