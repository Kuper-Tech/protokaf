package cmd

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
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

			msg := buildMessage(dynamic.NewMessage(messageDescriptor))

			b, err := msg.MarshalJSONPB(&jsonpb.Marshaler{
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

func buildMessage(message *dynamic.Message) *dynamic.Message {
	for _, field := range message.GetMessageDescriptor().GetFields() {
		if field.IsRepeated() {
			message.SetField(field, []interface{}{buildDefaultValue(field)})
		} else if field.IsMap() {
			message.SetField(
				field,
				map[interface{}]interface{}{
					buildDefaultValue(field.GetMapKeyType()): buildDefaultValue(field.GetMapValueType()),
				},
			)
		} else if field.GetOneOf() != nil {
			oneOfField := field.GetOneOf().GetChoices()[0]
			message.SetField(oneOfField, buildDefaultValue(oneOfField))
		} else {
			message.SetField(field, buildDefaultValue(field))
		}
	}

	return message
}

func buildDefaultValue(field *desc.FieldDescriptor) interface{} {
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return uint32(0)
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return int32(0)
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return uint64(0)
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return int64(0)
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return float32(0)
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return float64(0)
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return []byte(nil)
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return ""
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return field.GetEnumType().GetValues()[0].GetNumber()
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if field.GetMessageType().GetFullyQualifiedName() == "google.protobuf.Any" {
			val, _ := anypb.New(nil)

			return val
		}

		return buildMessage(dynamic.NewMessage(field.GetMessageType()))
	}

	return nil
}
