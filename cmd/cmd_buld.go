package cmd

import (
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/descriptorpb"
)

func NewBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <MessageName>",
		Short: "Build json by proto message",
		Args:  messageNameRequired,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// parse protofiles & create proto object
			p, err := parseProtofiles()
			if err != nil {
				return
			}

			messageDescriptor, err := findMessage(p, args[0])
			if err != nil {
				return
			}

			m := createMessage(
				*dynamic.NewMessageFactoryWithDefaults(),
				messageDescriptor,
			)

			b, err := m.MarshalJSONPB(&jsonpb.Marshaler{
				EmitDefaults: true,
				Indent:       "  ",
			})
			if err != nil {
				return
			}

			cmd.Println(string(b))

			return
		},
	}

	return cmd
}

func createMessage(f dynamic.MessageFactory, md *desc.MessageDescriptor) *dynamic.Message {
	m := f.NewDynamicMessage(md)

	for _, v := range m.GetKnownFields() {
		if v.IsProto3Optional() {
			if v.IsRepeated() {
				if mt := v.GetMessageType(); mt != nil {
					m.AddRepeatedField(v, createMessage(f, mt))
					continue
				}
			}

			if v.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				m.SetField(v, createMessage(f, v.GetMessageType()))
				continue
			}

			m.SetField(v, v.GetDefaultValue())
		}

		if v.IsRepeated() {
			if mt := v.GetMessageType(); mt != nil {
				m.AddRepeatedField(v, createMessage(f, mt))
			} else {
				switch v.GetType() {
				case descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
					descriptorpb.FieldDescriptorProto_TYPE_UINT32:
					m.AddRepeatedField(v, uint32(0))
				case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
					descriptorpb.FieldDescriptorProto_TYPE_INT32,
					descriptorpb.FieldDescriptorProto_TYPE_SINT32:
					m.AddRepeatedField(v, int32(0))
				case descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
					descriptorpb.FieldDescriptorProto_TYPE_UINT64:
					m.AddRepeatedField(v, uint64(0))
				case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
					descriptorpb.FieldDescriptorProto_TYPE_INT64,
					descriptorpb.FieldDescriptorProto_TYPE_SINT64:
					m.AddRepeatedField(v, int64(0))
				case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
					m.AddRepeatedField(v, float32(0))
				case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
					m.AddRepeatedField(v, float64(0))
				case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
					m.AddRepeatedField(v, false)
				case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
					m.AddRepeatedField(v, []byte(nil))
				case descriptorpb.FieldDescriptorProto_TYPE_STRING:
					m.AddRepeatedField(v, "")
				case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
					if v.IsProto3Optional() {
						m.AddRepeatedField(v, 0)
					} else {
						enumVals := v.GetEnumType().GetValues()
						if len(enumVals) > 0 {
							m.AddRepeatedField(v, enumVals[0].GetNumber())
						} else {
							m.AddRepeatedField(v, 0)
						}
					}
				default:
					panic(fmt.Sprintf("Unknown field type: %v", v.GetType()))
				}
			}
			continue
		}

		if v.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			m.SetField(v, createMessage(f, v.GetMessageType()))
			continue
		}
	}

	return m
}
