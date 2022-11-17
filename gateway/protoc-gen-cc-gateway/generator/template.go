package generator

import (
	"text/template"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

type TemplateParams struct {
	*descriptor.File
	Imports []descriptor.GoPackage
	Opts    Opts
}

var (
	headerTemplate = template.Must(template.New("header").Funcs(funcMap).Parse(`
// Code generated by protoc-gen-cc-gateway. DO NOT EDIT.
// source: {{ .GetName }}

/*
Package {{ .GoPkg.Name }} contains
  *   chaincode methods names {service_name}Chaincode_{method_name}
  *   chaincode interface definition {service_name}Chaincode 
  *   chaincode gateway definition {service_name}}Gateway
  *   chaincode service to cckit router registration func
*/
package {{ .GoPkg.Name }}
import (
    {{ range $i := .Imports }}{{ if $i.Standard }}{{ $i | printf "%s\n" }}{{ end }}{{ end }}

    {{ range $i := .Imports }}{{ if not $i.Standard }}{{ $i | printf "%s\n" }}{{ end }}{{ end }}
)
`))
)

var ccTemplate = template.Must(template.New("chaincode").Funcs(funcMap).Option().Parse(`

{{ $chaincodeMethodServicePrefix := .Opts.ChaincodeMethodServicePrefix }}

{{ range $svc := .Services }}
 
// {{ $svc.GetName }}Chaincode method names
const (

{{ $methodPrefix := "" }}

{{ if $chaincodeMethodServicePrefix  }}
 {{ $methodPrefix = printf "%s." $svc.GetName  }}
{{ end }}

// {{ $svc.GetName }}ChaincodeMethodPrefix allows to use multiple services with same method names in one chaincode
{{ $svc.GetName }}ChaincodeMethodPrefix = "{{ $methodPrefix }}"

{{ range $m := $svc.Methods }}
 {{ $svc.GetName }}Chaincode_{{ $m.GetName }} = {{ $svc.GetName }}ChaincodeMethodPrefix + "{{ $m.GetName }}"
{{ end }}
)

// {{ $svc.GetName }}Chaincode chaincode methods interface
type {{ $svc.GetName }}Chaincode interface {
{{ range $m := $svc.Methods }}
   {{ $m.GetName }} (cckit_router.Context, *{{$m.RequestType.GoType $m.Service.File.GoPkg.Path | goTypeName }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}, error)
{{ end }}
}

// Register{{ $svc.GetName }}Chaincode registers service methods as chaincode router handlers
func Register{{ $svc.GetName }}Chaincode(r *cckit_router.Group, cc {{ $svc.GetName }}Chaincode) error {

    {{ range $m := $svc.Methods }}
    {{ $method := "Invoke"}}
    {{ if $m | hasGetBinding }}{{ $method = "Query"}}{{ end }}

 r.{{ $method }}({{ $svc.GetName }}Chaincode_{{ $m.GetName }}, 
		func(ctx cckit_router.Context) (interface{}, error) {
			return cc.{{ $m.GetName }}(ctx, ctx.Param().(*{{$m.RequestType.GoType $m.Service.File.GoPkg.Path | goTypeName }}))
		},
		cckit_defparam.Proto(&{{$m.RequestType.GoType $m.Service.File.GoPkg.Path | goTypeName }}{}))

   {{ end }}

   return nil 
}
 
{{ end }}
`))

var gatewayTemplate = template.Must(template.New("gateway").Funcs(funcMap).Option().Parse(`

{{ $source :=.Name }}
{{ $embedSwagger :=.Opts.EmbedSwagger }}

{{ range $svc := .Services }}

{{ if $embedSwagger }}
 //go:embed {{ $source | removeExtension }}.swagger.json
{{ end }} var {{ $svc.GetName }}Swagger []byte


// New{{ $svc.GetName }}Gateway creates gateway to access chaincode method via chaincode service
func New{{ $svc.GetName }}Gateway(sdk cckit_sdk.SDK , channel, chaincode string, opts ...cckit_gateway.Opt) *{{ $svc.GetName }}Gateway {
	return New{{ $svc.GetName }}GatewayFromInstance(
          cckit_gateway.NewChaincodeInstanceService ( 
                sdk, 
                &cckit_gateway.ChaincodeLocator { Channel : channel, Chaincode: chaincode },
                opts...,
    ))
}

func New{{ $svc.GetName }}GatewayFromInstance (chaincodeInstance cckit_gateway.ChaincodeInstance) *{{ $svc.GetName }}Gateway {
  return &{{ $svc.GetName }}Gateway{
       ChaincodeInstance: chaincodeInstance,
    }
}

// gateway implementation
// gateway can be used as kind of SDK, GRPC or REST server ( via grpc-gateway or clay )
type {{ $svc.GetName }}Gateway struct {
	ChaincodeInstance cckit_gateway.ChaincodeInstance
}


func (c *{{ $svc.GetName }}Gateway) Invoker() cckit_gateway.ChaincodeInstanceInvoker {
   return cckit_gateway.NewChaincodeInstanceServiceInvoker(c.ChaincodeInstance)
}


// ServiceDef returns service definition
func (c *{{ $svc.GetName }}Gateway) ServiceDef() cckit_gateway.ServiceDef {
	return cckit_gateway.NewServiceDef(
        _{{ $svc.GetName }}_serviceDesc.ServiceName,
        {{ $svc.GetName }}Swagger,
        &_{{ $svc.GetName }}_serviceDesc,
        c,
        Register{{ $svc.GetName }}HandlerFromEndpoint,
	)
}


 {{ range $m := $svc.Methods }}
 {{ $method := "Invoke"}}
 {{ if $m | hasGetBinding }}{{ $method = "Query"}}{{ end }}

 func (c *{{ $svc.GetName }}Gateway) {{ $m.GetName }}(ctx context.Context, in *{{$m.RequestType.GoType $m.Service.File.GoPkg.Path | goTypeName }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}, error) {
    var inMsg interface{} = in
    if v, ok := inMsg.(interface { Validate() error }); ok {
       if err := v.Validate(); err != nil {
		return nil, err
	   } 
     }

    if res, err := c.Invoker().{{ $method }}(ctx, {{ $svc.GetName }}Chaincode_{{ $m.GetName }} , []interface{}{in}, &{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}{}); err != nil {
		return nil, err
	} else {
		return res.(*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}), nil
	}
 }
 {{ end }}

{{ end }}
`))

var resolverTemplate = template.Must(template.New("resolver").Funcs(funcMap).Option().Parse(`

{{ range $svc := .Services }}

// {{ $svc.GetName }}ChaincodeResolver interface for service resolver
type (
 {{ $svc.GetName }}ChaincodeResolver interface {
		Resolve (ctx cckit_router.Context) ({{ $svc.GetName }}Chaincode, error)
 }

 {{ $svc.GetName }}ChaincodeLocalResolver struct {
   service {{ $svc.GetName }}Chaincode
 }

 {{ $svc.GetName }}ChaincodeLocatorResolver struct {
   locatorResolver cckit_gateway.ChaincodeLocatorResolver
   service {{ $svc.GetName }}Chaincode
 }
) 

func New{{ $svc.GetName }}ChaincodeLocalResolver (service {{ $svc.GetName }}Chaincode) *{{ $svc.GetName }}ChaincodeLocalResolver {
	return &{{ $svc.GetName }}ChaincodeLocalResolver {
		service: service,
	}
}

func (r *{{ $svc.GetName }}ChaincodeLocalResolver) Resolve(ctx cckit_router.Context) ({{ $svc.GetName }}Chaincode, error) {
	if r.service == nil {
		return nil, errors.New ("service not set for local chaincode resolver")
    }

    return r.service, nil
}

func New{{ $svc.GetName }}ChaincodeResolver (locatorResolver cckit_gateway.ChaincodeLocatorResolver) *{{ $svc.GetName }}ChaincodeLocatorResolver {
	return &{{ $svc.GetName }}ChaincodeLocatorResolver {
		locatorResolver: locatorResolver,
	}
}

func (r *{{ $svc.GetName }}ChaincodeLocatorResolver) Resolve(ctx cckit_router.Context) ({{ $svc.GetName }}Chaincode, error) {
	if r.service != nil {
		return r.service, nil
    }

    locator, err := r.locatorResolver(ctx, _{{ $svc.GetName }}_serviceDesc.ServiceName)
	if err != nil {
        return nil, err
	}

	r.service = New{{ $svc.GetName }}ChaincodeStubInvoker(locator)
	return r.service, nil
}


type {{ $svc.GetName }}ChaincodeStubInvoker struct {
  Invoker cckit_gateway.ChaincodeStubInvoker
}

func New{{ $svc.GetName }}ChaincodeStubInvoker(locator *cckit_gateway.ChaincodeLocator) *{{ $svc.GetName }}ChaincodeStubInvoker {
	return &{{ $svc.GetName }}ChaincodeStubInvoker {
		Invoker: &cckit_gateway.LocatorChaincodeStubInvoker{Locator: locator},
	}
}


 {{ range $m := $svc.Methods }}
 {{ $method := "Invoke"}}
 {{ if $m | hasGetBinding }}{{ $method = "Query"}}{{ end }}

 func (c *{{ $svc.GetName }}ChaincodeStubInvoker) {{ $m.GetName }}(ctx cckit_router.Context, in *{{$m.RequestType.GoType $m.Service.File.GoPkg.Path | goTypeName }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}, error) {

    {{ if $method | eq "Invoke" }}
       return nil, cckit_gateway.ErrInvokeMethodNotAllowed 

    {{ else }}
    var inMsg interface{} = in
    if v, ok := inMsg.(interface { Validate() error }); ok {
       if err := v.Validate(); err != nil {
		return nil, err
	   } 
     }

    if res, err := c.Invoker.{{ $method }}(ctx.Stub(), {{ $svc.GetName }}Chaincode_{{ $m.GetName }} , []interface{}{in}, &{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}{}); err != nil {
		return nil, err
	} else {
		return res.(*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path | goTypeName }}), nil
	}
    {{ end }}
 }
 {{ end }}


{{ end }}

`))