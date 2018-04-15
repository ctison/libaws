// Code generated by Makefile DO NOT EDIT

package libaws

// Templates ...
const Templates = `{{- define "LIBAWS::ApiGateway" }}

    {{- $m := map.Copy . }}

	{{- /* Logical IDs */}}
    {{- $m.SetDefault ""                                                           "LogicalIdSuffix" }}
    {{- $m.SetDefault (printf "RestApi%s" $m.LogicalIdSuffix)                      "RestApi" "LogicalId" }}
	{{- $m.SetDefault (printf "Deployment%s%d" $m.LogicalIdSuffix (time.Now.Unix)) "Deployment" "LogicalId" }}

	{{- /* Globals */}}
	{{- global.Set (printf "Deployment%sDependsOn" $m.LogicalIdSuffix) (slice.New) }}

	{{- /* RestApi */}}
    {{- template "AWS::ApiGateway::RestApi" $m.RestApi }}

	{{- /* Methods */}}
	{{- range $i, $v := $m.Methods }}
		{{- $v := map.New $v }}
		{{- $v.SetDefault (printf "Method%s%d" $m.LogicalIdSuffix $i)               "LogicalId"  }}
		{{- $v.SetDefault "GET"                                                     "HttpMethod" }}
		{{- $v.Set        (printf "!GetAtt %s.RootResourceId" $m.RestApi.LogicalId) "ResourceId" }}
		{{- $v.Set        (global.Get "RestApiId")                                  "RestApiId" }}
		{{- global.Append (printf "Deployment%sDependsOn" $m.LogicalIdSuffix) $v.LogicalId }}
		{{- if $v.LambdaFunction }}
			{{- map.SetDefault "AWS_PROXY" $v "Integration" "Type" }}
			{{- map.SetDefault (printf "!Sub 'arn:aws:apigateway:${AWS::Region}:lambda:path/2015-03-31/functions/${%s}/invocations'" $v.LambdaFunction ) $v.Integration "Uri" }}
		{{- end }}
		{{- template "AWS::ApiGateway::Method" $v }}
	{{- end }}

	{{- /* Resources */}}
	{{- global.Set "ParentId" (printf "!GetAtt %s.RootResourceId" $m.RestApi.LogicalId)  }}
	{{- template "LIBAWS::ApiGateway::Resources" $m.Resources }}

	{{- /* Deployment */}}
	{{- map.Set (global.Get "DeploymentDependsOn") $m.Deployment "DependsOn" }}
	{{- map.SetDefault (global.Get "RestApiId") $m.Deployment "RestApiId" }}
	{{- template "AWS::ApiGateway::Deployment" $m.Deployment }}

{{- end }}

{{- define "LIBAWS::ApiGateway::Resources" }}
	{{- range $v := .Resources }}
		{{- $Resource := map.Copy $v }}
		{{- map.SetDefault (printf "%s%s" (or $.LogicalIdSuffix "") (strings.Title $Resource.PathPart)) $Resource "LogicalIdSuffix" }}
		{{- map.SetDefault (printf "Resource%s" $Resource.LogicalIdSuffix) $Resource "LogicalId" }}
		{{- map.Set $.ParentId $Resource "ParentId" }}
		{{- map.Set (global.Get "RestApiId") $Resource "RestApiId" }}
		{{- template "AWS::ApiGateway::Resource" $Resource }}
		{{- if $Resource.Resources }}
			{{- $Resources := map.New }}
			{{- map.SetDefault $Resource.LogicalIdSuffix $Resources "LogicalIdSuffix" }}
			{{- map.Set (printf "!Ref %s" $Resource.LogicalId) $Resources "ParentId"  }}
			{{- map.Set (global.Get "RestApiId") $Resources "RestApiId" }}
			{{- map.Set $Resource.Resources $Resources "Resources" }}
			{{- template "LIBAWS::ApiGateway::Resources" $Resources }}
		{{- end }}
	{{- end }}
{{- end }}
{{- define "LIBAWS::Lambda::old" }}

	{{- $m := map.Copy . }}

	{{- /* Logica IDs */}}
	{{- $m.SetDefault ""                                       "LogicalIdSuffix"      }}
	{{- $m.SetDefault (printf "Function%s" $m.LogicalIdSuffix) "Function" "LogicalId" }}
	{{- $m.SetDefault (printf "Role%s"     $m.LogicalIdSuffix) "Role"     "LogicalId" }}
	{{- $m.SetDefault (printf "LogGroup%s" $m.LogicalIdSuffix) "LogGroup" "LogicalId" }}
	{{- $m.SetDefault (printf "Policy%s"   $m.LogicalIdSuffix) "Policy"   "LogicalId" }}

	{{- /* Function */}}
	{{- $m.SetDefault "go1.x"                                       "Function" "Runtime" }}
	{{- $m.SetDefault "main"                                        "Function" "Handler" }}
	{{- $m.SetDefault (printf "lambdas/%s.zip" $m.Function.Handler) "Function" "Code"    }}
	{{- $m.Set        (printf "!GetAtt %s.Arn" $m.Role.LogicalId)   "Function" "Role"    }}
	{{- template "AWS::Lambda::Function" $m.Function }}

	{{- /* Role */}}
	{{- $m.Set        (libaws.CreateAssumeRolePolicyDocument "lambda.amazonaws.com")                 "Role" "AssumeRolePolicyDocument" }}
	{{- $m.SetDefault (slice.New "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole") "Role" "ManagedPolicyArns"        }}
	{{- template "AWS::IAM::Role" $m.Role }}

	{{- /* LogGroup */}}
	{{- $m.SetDefault (printf "!Sub '/aws/lambda/${%s}'" $m.Function.LogicalId) "LogGroup" "LogGroupName" }}
	{{- template "AWS::Logs::LogGroup" $m.LogGroup }}

	{{- /* Permissions */}}
	{{- range $i, $e := $m.Permissions }}
		{{- $e := map.New $e }}
	    {{- $e.SetDefault (printf "Permission%s%d" $m.LogicalIdSuffix $i) "LogicalId"    }}
	    {{- $e.SetDefault "lambda:InvokeFunction"                         "Action"       }}
	    {{- $e.Set        (printf "!GetAtt %s.Arn" $m.Function.LogicalId) "FunctionName" }}
	    {{- template "AWS::Lambda::Permission" $e }}
	{{- end }}

{{- end }}
{{- define "api" }}
{{- $ApiName        := (or .ApiName (printf "Api%s" (or .DefaultName ""))) }}
{{- $ApiTitle       := (or .ApiTitle $ApiName ) }}
{{- $StageName      := (or .StageName "v1") }}

    {{ $ApiName }}:
        Type: AWS::ApiGateway::RestApi
        Properties:
            Body:
                swagger: '2.0'
                info:
                    title: {{ $ApiTitle }}
                    version: latest
{{- with .UserPools }}
                securityDefinitions:
{{- range . }}
                    {{ .Name }}:
                        type: apiKey
                        name: Authorization
                        in: header
                        x-amazon-apigateway-authtype: cognito_user_pools
                        x-amazon-apigateway-authorizer:
                            providerARNs:
                                - {{ .Arn }}
                            type: cognito_user_pools
{{- end }}
{{- end }}
                paths:
{{- range .Paths }}
                    {{ .Path }}:
{{- range .Methods }}
                        {{ .Method }}:
{{- with .UserPoolName }}
                            security: [{{ . }}: []]
{{- end }}
                            x-amazon-apigateway-integration:
                                type: aws_proxy
                                httpMethod: POST
                                uri: !Sub
                                    - 'arn:aws:apigateway:${AWS::Region}:lambda:path/2015-03-31/functions/${functionArn}/invocations'
                                    - functionArn: {{ .LambdaArn }}
{{- end }}
{{- end }}

    Deployment{{ (call time.Now).Unix }}:
        Type: AWS::ApiGateway::Deployment
        DependsOn: {{ $ApiName }}
        Properties:
            RestApiId: !Ref {{ $ApiName }}
            StageName: {{ $StageName }}
{{- end }}
{{- define "api.url" }}
{{- $ApiName        := (or .ApiName (printf "Api%s" (or .DefaultName ""))) }}
{{- $StageName      := (or .StageName "v1") -}}

!Sub 'https://${ {{- $ApiName -}} }.execute-api.${AWS::Region}.amazonaws.com/{{ $StageName }}'

{{- end }}
{{- define "apiv2" }}
{{- $ApiName := .ApiName | or (printf "Api%s" (.DefaultName | or "")) }}

    {{ $ApiName }}:
        Type: AWS::ApiGateway::RestApi
        Properties:
            Name: {{ $ApiName }}

{{- range $k, $v := .Resources }}
    Resource{{ $ApiName }}{{ $k }}:
        Type: AWS::ApiGateway::Resource
        Properties:
            ParentId: !GetAtt {{ $ApiName }}.RootResourceId
            PathPart: {{ $v.Path }}
            RestApiId: !Ref {{ $ApiName }}

{{- range $kk, $vv := $v.Methods }}
    Method{{ $ApiName }}{{ $k }}x{{ $kk }}:
        Type: AWS::ApiGateway::Method
        Properties:
            AuthorizationType: {{ with $vv.AuthorizationType }}{{ . }}{{ else }}NONE{{ end }}
            HttpMethod: {{ with $vv.HttpMethod }}{{ . }}{{ else }}GET{{ end }}
            ResourceId: !Ref Resource{{ $ApiName }}{{ $k }}
            RestApiId: !Ref {{ $ApiName }}
            Integration:
                Type: AWS_PROXY
                Uri: !Sub "arn:aws:apigateway:${AWS::Region}:lambda:path/2015-03-31/functions/${.LambdaFunction}/invocations"
{{- end }}
{{- end }}
{{- end }}
{{- define "LIBAWS::Lambda" }}
	{{- libaws.Lambda . }}
{{- end }}
{{- define "LIBAWS::Lambda::Function" }}
	{{- libaws.LambdaFunction . }}
{{- end }}
`