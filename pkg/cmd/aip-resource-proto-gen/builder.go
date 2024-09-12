package main

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/stoewer/go-strcase"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type schemaBuilder struct {
	cfg *Config

	file     *builder.FileBuilder
	resource *builder.MessageBuilder
	service  *builder.ServiceBuilder
}

func (s *schemaBuilder) Build() (*desc.FileDescriptor, error) {
	return s.buildDescriptor()
}

func (s *schemaBuilder) buildDescriptor() (*desc.FileDescriptor, error) {
	b := builder.NewFile(s.cfg.Resource + ".proto")
	b.SetProto3(s.cfg.Syntax == "proto3")
	b.SetPackageName(s.cfg.Package)

	s.file = b

	s.buildResourceMessage()
	s.buildServiceDescriptor()

	return b.Build()
}

func (s *schemaBuilder) buildResourceMessage() {
	c := s.cfg

	b := builder.NewMessage(c.Resource)
	b.SetComments(comment(c.Resource+" resource.", ""))
	b.SetOptions(messageOptions(resource(
		&annotations.ResourceDescriptor{
			Type:     c.ResourceTypeName(),
			Singular: strcase.LowerCamelCase(c.Resource),
			Plural:   c.ResourceCollectionIdentifier(),
		},
	)))

	nameField := builder.NewField("name", builder.FieldTypeString())
	nameField.SetComments(comment("The resource's name.", ""))
	nameField.SetOptions(fieldOptions(identifier()))
	b.AddField(nameField)

	if c.WithDisplayName {
		displayNameField := builder.NewField("display_name", builder.FieldTypeString())
		displayNameField.SetComments(comment("The resource's display name.", ""))
		displayNameField.SetOptions(fieldOptions(optional()))
		b.AddField(displayNameField)
	}

	if c.WithTimestamps {
		tsDesc, err := desc.LoadMessageDescriptorForMessage((*timestamppb.Timestamp)(nil))
		if err != nil {
			panic(err)
		}

		ts := builder.FieldTypeImportedMessage(tsDesc)

		createTimeField := builder.NewField("create_time", ts)
		createTimeField.SetComments(comment("The time at which the resource was created.", ""))
		createTimeField.SetOptions(fieldOptions(outputOnly()))
		b.AddField(createTimeField)

		updateTimeField := builder.NewField("update_time", ts)
		updateTimeField.SetComments(comment("The time at which the resource was last updated.", ""))
		updateTimeField.SetOptions(fieldOptions(outputOnly()))
		b.AddField(updateTimeField)
	}

	if c.WithAnnotations {
		annotationsField := builder.NewMapField("annotations", builder.FieldTypeString(), builder.FieldTypeString())
		annotationsField.SetComments(comment("Custom annotations defined by the caller.", ""))
		annotationsField.SetOptions(fieldOptions(optional()))
		b.AddField(annotationsField)
	}

	s.file.AddMessage(b)
	s.resource = b
}

func (s *schemaBuilder) buildServiceDescriptor() {
	c := s.cfg

	b := builder.NewService(c.Resource)
	b.SetName(c.Resource + "Service")
	b.SetComments(comment("Service for managing the "+c.Resource+" resource.", ""))
	b.SetOptions(serviceOptions(defaultHost(c.Service)))

	s.file.AddService(b)
	s.service = b

	s.buildGetMethod()
	s.buildListMethod()
	s.buildCreateMethod()
	s.buildUpdateMethod()
	s.buildDeleteMethod()

}

func (s *schemaBuilder) buildGetMethod() {
	if !strings.Contains(s.cfg.Methods, "r") {
		return
	}
	c := s.cfg

	name := "Get" + c.Resource

	reqType := "Get" + c.Resource + "Request"
	req := builder.NewMessage(reqType)
	req.SetComments(comment("Request for "+name+" method.", ""))
	nameField := builder.NewField("name", builder.FieldTypeString())
	nameField.SetComments(comment("The name of the resource to retrieve.", ""))
	nameField.SetOptions(fieldOptions(required()))
	req.AddField(nameField)

	reqRpc := builder.RpcTypeMessage(req, false)

	resRpc := builder.RpcTypeMessage(s.resource, false)

	m := builder.NewMethod(name, reqRpc, resRpc)
	m.SetComments(comment("Get the "+c.Resource+" resource", ""))
	if c.WithHTTPOptions {
		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Get{
				Get: fmt.Sprintf("/v1/{name=%s}", c.ResourceNameUrlRef()),
			},
		}
		m.SetOptions(methodOptions(httpRule(rule), methodSignature("name")))
	}

	s.file.AddMessage(req)
	s.service.AddMethod(m)
}

func (s *schemaBuilder) buildListMethod() {
	if !strings.Contains(s.cfg.Methods, "l") {
		return
	}
	c := s.cfg

	name := "List" + c.Resource

	// Request
	reqType := "List" + c.Resource + "Request"
	req := builder.NewMessage(reqType)
	req.SetComments(comment("Request for "+name+" method.", ""))

	if c.HasParent() {
		parentField := builder.NewField("parent", builder.FieldTypeString())
		parentField.SetComments(comment("The resource's parent.", ""))
		parentField.SetOptions(fieldOptions(required()))
		req.AddField(parentField)
	}

	pageSizeField := builder.NewField("page_size", builder.FieldTypeInt32())
	pageSizeField.SetComments(comment("The maximum number of resources to return.", ""))
	pageSizeField.SetOptions(fieldOptions(optional()))
	req.AddField(pageSizeField)

	pageTokenField := builder.NewField("page_token", builder.FieldTypeString())
	pageTokenField.SetComments(comment("The page token to use for pagination. Provide this to retrieve subsequent page", ""))
	pageTokenField.SetOptions(fieldOptions(optional()))
	req.AddField(pageTokenField)

	if c.WithListFilter {
		filterField := builder.NewField("filter", builder.FieldTypeString())
		filterField.SetComments(comment("The filter to apply to list results.", ""))
		filterField.SetOptions(fieldOptions(optional()))
		req.AddField(filterField)
	}

	if c.WithListOrderBy {
		orderByField := builder.NewField("order_by", builder.FieldTypeString())
		orderByField.SetComments(comment("The order to list results by.", ""))
		orderByField.SetOptions(fieldOptions(optional()))
		req.AddField(orderByField)
	}

	reqRpc := builder.RpcTypeMessage(req, false)

	// Response
	resType := "List" + c.Resource + "Response"
	res := builder.NewMessage(resType)
	res.SetComments(comment("Response for "+name+" method.", ""))

	resourceField := builder.NewField(c.PluralResourceSnakeCase(), builder.FieldTypeMessage(s.resource))
	resourceField.SetRepeated()
	resourceField.SetComments(comment("The list of "+c.Resource+" resources.", ""))
	res.AddField(resourceField)

	nextTokenField := builder.NewField("next_page_token", builder.FieldTypeString())
	nextTokenField.SetComments(comment("The token to retrieve the next page of results, or empty if there are no more results.", ""))
	res.AddField(nextTokenField)

	resRpc := builder.RpcTypeMessage(res, false)

	m := builder.NewMethod(name, reqRpc, resRpc)
	m.SetComments(comment("List the "+c.Resource+" resources", ""))
	if c.WithHTTPOptions {
		parentVar := ""
		methodSig := ""
		if c.HasParent() {
			parentVar = fmt.Sprintf("{parent=%s}/", c.ParentNameUrlRef())
			methodSig = "parent"
		}

		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Get{
				Get: fmt.Sprintf("/v1/%s%s", parentVar, c.ResourceCollectionIdentifier()),
			},
		}
		m.SetOptions(methodOptions(httpRule(rule), methodSignature(methodSig)))
	}

	s.file.AddMessage(req)
	s.file.AddMessage(res)
	s.service.AddMethod(m)
}

func (s *schemaBuilder) buildCreateMethod() {
	if !strings.Contains(s.cfg.Methods, "c") {
		return
	}

	c := s.cfg

	name := "Create" + c.Resource

	// Request Message
	reqType := "Create" + c.Resource + "Request"
	req := builder.NewMessage(reqType)
	req.SetComments(comment("Request for "+name+" method.", ""))

	if c.HasParent() {
		parentField := builder.NewField("parent", builder.FieldTypeString())
		parentField.SetComments(comment("The resource's parent.", ""))
		parentField.SetOptions(fieldOptions(required()))
		req.AddField(parentField)
	}

	idField := builder.NewField(c.ResourceSnakeCase()+"_id", builder.FieldTypeString())
	idField.SetComments(comment("The ID to use for the resource. It will become the final component of the name.", ""))
	if c.IDRequired {
		idField.SetOptions(fieldOptions(required()))
	}
	req.AddField(idField)

	resourceField := builder.NewField(c.ResourceSnakeCase(), builder.FieldTypeMessage(s.resource))
	resourceField.SetComments(comment("The "+c.Resource+" resource to create.", ""))
	resourceField.SetOptions(fieldOptions(required()))
	req.AddField(resourceField)
	reqRpc := builder.RpcTypeMessage(req, false)

	// Response Message
	resRpc := builder.RpcTypeMessage(s.resource, false)

	m := builder.NewMethod(name, reqRpc, resRpc)
	m.SetComments(comment("Create a new "+c.Resource+" resource", ""))
	if c.WithHTTPOptions {
		parentVar := ""
		methodSig := c.ResourceSnakeCase()
		if c.IDRequired {
			methodSig += "," + c.ResourceSnakeCase() + "_id"
		}

		if c.HasParent() {
			parentVar = fmt.Sprintf("{parent=%s}/", c.ParentNameUrlRef())
			methodSig = "parent," + methodSig
		}

		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Post{
				Post: fmt.Sprintf("/v1/%s%s", parentVar, c.ResourceCollectionIdentifier()),
			},
			Body: c.ResourceSnakeCase(),
		}
		m.SetOptions(methodOptions(httpRule(rule), methodSignature(methodSig)))
	}

	s.file.AddMessage(req)
	s.service.AddMethod(m)
}

func (s *schemaBuilder) buildUpdateMethod() {
	if !strings.Contains(s.cfg.Methods, "u") {
		return
	}

	c := s.cfg

	name := "Update" + c.Resource

	// Request Message
	reqType := "Update" + c.Resource + "Request"
	req := builder.NewMessage(reqType)
	req.SetComments(comment("Request for "+name+" method.", ""))

	resourceField := builder.NewField(c.ResourceSnakeCase(), builder.FieldTypeMessage(s.resource))
	resourceField.SetComments(comment("The "+c.Resource+" resource to update. The resource must have", ""))
	resourceField.SetOptions(fieldOptions(required()))
	req.AddField(resourceField)

	if s.cfg.WithUpdateFieldMask {
		fieldMaskDesc, err := desc.LoadMessageDescriptorForMessage((*fieldmaskpb.FieldMask)(nil))
		if err != nil {
			panic(err)
		}
		fm := builder.FieldTypeImportedMessage(fieldMaskDesc)
		updateMaskField := builder.NewField("update_mask", fm)
		updateMaskField.SetComments(comment("The list of fields to update.", ""))
		updateMaskField.SetOptions(fieldOptions(optional()))
		req.AddField(updateMaskField)
	}

	if s.cfg.WithUpdateAllowMissing {
		allowMissingField := builder.NewField("allow_missing", builder.FieldTypeBool())
		allowMissingField.SetComments(comment("If set to true, and the resource is not found, a new resource will be created.", ""))
		allowMissingField.SetOptions(fieldOptions(optional()))
		req.AddField(allowMissingField)
	}

	reqRpc := builder.RpcTypeMessage(req, false)

	// Response message
	resRpc := builder.RpcTypeMessage(s.resource, false)

	m := builder.NewMethod(name, reqRpc, resRpc)
	m.SetComments(comment("Update the "+c.Resource+" resource", ""))
	if c.WithHTTPOptions {
		nameVar := fmt.Sprintf("{%s.name=%s}", c.ResourceSnakeCase(), c.ResourceNameUrlRef())
		methodSig := c.ResourceSnakeCase()
		if c.WithUpdateFieldMask {
			methodSig += ",update_mask"
		}

		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Patch{
				Patch: fmt.Sprintf("/v1/%s", nameVar),
			},
			Body: c.ResourceSnakeCase(),
		}
		m.SetOptions(methodOptions(httpRule(rule), methodSignature(methodSig)))
	}

	s.file.AddMessage(req)
	s.service.AddMethod(m)
}

func (s *schemaBuilder) buildDeleteMethod() {
	if !strings.Contains(s.cfg.Methods, "d") {
		return
	}

	c := s.cfg

	name := "Delete" + c.Resource

	reqType := "Delete" + c.Resource + "Request"
	req := builder.NewMessage(reqType)
	req.SetComments(comment("Request for "+name+" method.", ""))

	nameField := builder.NewField("name", builder.FieldTypeString())
	nameField.SetComments(comment("The name of the resource to delete.", ""))
	nameField.SetOptions(fieldOptions(required()))
	req.AddField(nameField)

	if s.cfg.WithUpdateAllowMissing {
		allowMissingField := builder.NewField("allow_missing", builder.FieldTypeBool())
		allowMissingField.SetComments(comment("If set to true, and the resource is not found, no errors will be returned.", ""))
		allowMissingField.SetOptions(fieldOptions(optional()))
		req.AddField(allowMissingField)
	}

	reqRpc := builder.RpcTypeMessage(req, false)

	emptyDesc, err := desc.LoadMessageDescriptorForMessage((*emptypb.Empty)(nil))
	if err != nil {
		panic(err)
	}

	resRpc := builder.RpcTypeImportedMessage(emptyDesc, false)

	m := builder.NewMethod(name, reqRpc, resRpc)
	m.SetComments(comment("Delete the "+c.Resource+" resource", ""))
	if c.WithHTTPOptions {
		nameVar := fmt.Sprintf("{name=%s}", c.ResourceNameUrlRef())
		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Delete{
				Delete: fmt.Sprintf("/v1/%s", nameVar),
			},
		}
		m.SetOptions(methodOptions(httpRule(rule), methodSignature("name")))
	}

	s.file.AddMessage(req)
	s.service.AddMethod(m)
}

func initPrinter(c *Config) *protoprint.Printer {
	p := &protoprint.Printer{Compact: c.Compact}
	return p
}

func comment(leading, trailing string) builder.Comments {
	return builder.Comments{
		LeadingComment:  leading,
		TrailingComment: trailing,
	}
}

func fieldOptions(opts ...fOpts) *descriptorpb.FieldOptions {
	if len(opts) == 0 {
		return nil
	}

	o := &descriptorpb.FieldOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

func fieldBehavior(behaviors ...annotations.FieldBehavior) fOpts {
	return fOptsFn(func(opts *descriptorpb.FieldOptions) {
		proto.SetExtension(opts, annotations.E_FieldBehavior, behaviors)
	})
}

func required() fOpts {
	return fieldBehavior(annotations.FieldBehavior_REQUIRED)
}

func identifier() fOpts {
	return fieldBehavior(annotations.FieldBehavior_IDENTIFIER)
}

func optional() fOpts {
	return fieldBehavior(annotations.FieldBehavior_OPTIONAL)
}

func outputOnly() fOpts {
	return fieldBehavior(annotations.FieldBehavior_OUTPUT_ONLY)
}

func methodOptions(opts ...mOpts) *descriptorpb.MethodOptions {
	if len(opts) == 0 {
		return nil
	}

	o := &descriptorpb.MethodOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

func httpRule(rule *annotations.HttpRule) mOpts {
	return mOptsFn(func(opts *descriptorpb.MethodOptions) {
		proto.SetExtension(opts, annotations.E_Http, rule)
	})
}

func methodSignature(signatures ...string) mOpts {
	return mOptsFn(func(opts *descriptorpb.MethodOptions) {
		proto.SetExtension(opts, annotations.E_MethodSignature, signatures)
	})
}

func resource(resource *annotations.ResourceDescriptor) msgOpts {
	return msgOptsFn(func(opts *descriptorpb.MessageOptions) {
		proto.SetExtension(opts, annotations.E_Resource, resource)
	})
}

func messageOptions(opts ...msgOpts) *descriptorpb.MessageOptions {
	if len(opts) == 0 {
		return nil
	}

	o := &descriptorpb.MessageOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

func defaultHost(host string) svcOpts {
	return svcOptsFn(func(opts *descriptorpb.ServiceOptions) {
		proto.SetExtension(opts, annotations.E_DefaultHost, host)
	})
}

func serviceOptions(opts ...svcOpts) *descriptorpb.ServiceOptions {
	if len(opts) == 0 {
		return nil
	}

	o := &descriptorpb.ServiceOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

type (
	fOpts interface {
		apply(*descriptorpb.FieldOptions)
	}

	fOptsFn func(*descriptorpb.FieldOptions)

	mOpts interface {
		apply(*descriptorpb.MethodOptions)
	}

	mOptsFn func(*descriptorpb.MethodOptions)

	msgOpts interface {
		apply(*descriptorpb.MessageOptions)
	}

	msgOptsFn func(*descriptorpb.MessageOptions)

	svcOpts interface {
		apply(*descriptorpb.ServiceOptions)
	}

	svcOptsFn func(*descriptorpb.ServiceOptions)
)

func (m mOptsFn) apply(opts *descriptorpb.MethodOptions) {
	m(opts)
}

func (f fOptsFn) apply(opts *descriptorpb.FieldOptions) {
	f(opts)
}

func (m msgOptsFn) apply(opts *descriptorpb.MessageOptions) {
	m(opts)
}

func (s svcOptsFn) apply(opts *descriptorpb.ServiceOptions) {
	s(opts)
}
