package templates

import (
	"github.com/chtison/libgo/tmpl"
	"text/template"
)

const CloudFormationTmplYaml = `AWSTemplateFormatVersion: 2010-09-09

Resources:
    {{- /* Add resources below */}}
`
const CloudFormationDataYaml = `
`
const LibawsYaml = `TemplateFiles: cloudformation.tmpl.yaml
DataFiles: cloudformation.data.yaml
`
const Makefile = `.PHONY: default template clean

TMPL_DEPS := cloudformation.tmpl.yaml cloudformation.data.yaml

OUT_FILE  := cloudformation.yaml
OUT_DIR   := out

default: template

$(OUT_DIR):
	@mkdir -p $@

template: $(OUT_DIR)/$(OUT_FILE)
$(OUT_DIR)/$(OUT_FILE): $(OUT_DIR) $(TMPL_DEPS)
	libaws template run > $@

clean:
	rm -rf -- $(OUT_DIR)
`

var Template = template.Must(template.New("libaws").Funcs(tmpl.Funcs()).Parse(`
{{- define "sns-topic" }}
{{- $TopicLID        := (or .TopicLID "Topic"              ) }}
{{- $SubscriptionLID := (or .SubscriptionLID "Subscription") }}

    {{- $TopicLID }}:
        Type: AWS::SNS::Topic

{{- range $i, $e := .Subscriptions }}

    {{ $SubscriptionLID }}{{ $i }}:
        Type: AWS::SNS::Subscription
        DependsOn: {{ $TopicLID }}
        Properties:
            TopicArn: !Ref {{ $TopicLID }}
            Protocol: '{{ index $e 0 }}'
            Endpoint: '{{ index $e 1 }}'
{{- end }}
{{- end }}
{{- define "lambda" }}
{{- $FunctionName   := (or .FunctionName   (printf "Function%s"   (or .DefaultName ""))) }}
{{- $RoleName       := (or .RoleName       (printf "Role%s"       (or .DefaultName ""))) }}
{{- $LogGroupName   := (or .LogGroupName   (printf "LogGroup%s"   (or .DefaultName ""))) }}
{{- $PolicyName     := (or .PolicyName     (printf "Policy%s"     (or .DefaultName ""))) }}
{{- $PermissionName := (or .PermissionName (printf "Permission%s" (or .DefaultName ""))) }}
{{- $EventName      := (or .EventName      (printf "Event%s"      (or .DefaultName ""))) }}

    {{- $FunctionName }}:
        Type: AWS::Lambda::Function
        DependsOn: {{ $RoleName }}
        Properties:
            Runtime: {{ with .Runtime }}{{ . }}{{ else }}python3.6{{ end }}
            Handler: {{ with .Handler }}{{ . }}{{ else }}lambda.handler{{ end }}
            Code: {{ with .Zip }}{{ . }}{{ else }}lambda.zip{{ end }}
            Role: !GetAtt {{ $RoleName }}.Arn
{{- with .Timeout }}
            Timeout: {{ . }}
{{- end }}
{{- with .Environment }}
            Environment:
                Variables:
{{- range $k, $v := . }}
                    {{ $k }}: {{ $v }}
{{- end }}{{ end }}

    {{ $RoleName }}:
        Type: AWS::IAM::Role
        Properties:
            AssumeRolePolicyDocument:
                Version: 2012-10-17
                Statement:
                    Effect: Allow
                    Action: sts:AssumeRole
                    Principal:
                        Service: lambda.amazonaws.com

    {{ $LogGroupName }}:
        Type: AWS::Logs::LogGroup
        DependsOn: {{ $FunctionName }}
        Properties:
            LogGroupName: !Sub '/aws/lambda/${ {{- $FunctionName -}} }'
            RetentionInDays: {{ with .LogGroupRetentionInDays }}{{ . }}{{ else }}7{{ end }}

    {{ $PolicyName }}:
        Type: AWS::IAM::Policy
        DependsOn: [{{ $RoleName }}, {{ $LogGroupName }}]
        Properties:
            PolicyName: Default
            Roles: [!Ref {{ $RoleName }}]
            PolicyDocument:
                Version: 2012-10-17
                Statement:
                    - Effect: Allow
                      Action:
                          - logs:CreateLogStream
                          - logs:PutLogEvents
                      Resource: !GetAtt {{ $LogGroupName }}.Arn
{{- range .Policies }}
                    - Effect: {{ with .Effect }}{{ . }}{{ else }}Allow{{ end }}
                      Action:
{{- range .Action }}
                          - {{ . }}
{{- end }}
                      Resource: {{ .Resource }}
{{- end }}

{{- range $i, $e := .Permissions }}{{ $PermissionName := (printf "%s%d" $PermissionName $i) }}

    {{ $PermissionName }}:
        Type: AWS::Lambda::Permission
        DependsOn: {{ $FunctionName }}
        Properties:
            Action: {{ with $e.Action }}{{ . }}{{ else }}lambda:InvokeFunction{{ end }}
            FunctionName: !GetAtt {{ $FunctionName }}.Arn
            Principal: {{ $e.Principal }}
{{- with $e.SourceAccount }}
            SourceAccount: {{ . }}
{{- end }}
{{- with $e.SourceArn }}
            SourceArn: {{ . }}
{{- end }}
{{- end }}

{{- range $i, $e := .Schedules }}{{ $EventName := (printf "%s%d" $EventName $i) }}

    {{ $EventName }}:
        Type: AWS::Events::Rule
        DependsOn: {{ $FunctionName }}
        Properties:
            Targets:
                - Arn: !GetAtt {{ $FunctionName }}.Arn
                  Id: !Ref {{ $FunctionName }}
            ScheduleExpression: '{{ $e }}'

    FunctionPermissionFor{{ $EventName }}:
        Type: AWS::Lambda::Permission
        DependsOn: [{{ $FunctionName }}, {{ $EventName }}]
        Properties:
            Action: lambda:InvokeFunction
            FunctionName: !GetAtt {{ $FunctionName }}.Arn
            Principal: events.amazonaws.com
            SourceArn: !GetAtt {{ $EventName }}.Arn
{{- end }}
{{- end }}
{{- define "cognito-userpool" }}
{{- "# Hello World !" }}
{{- end }}
{{- define "api" }}
{{- $ApiName        := (or .ApiName (printf "Api%s" (or .DefaultName ""))) }}
{{- $ApiTitle       := (or .ApiTitle $ApiName ) }}
{{- $StageName      := (or .StageName "api") }}
{{- $DeploymentName := (strings.NewReplacer "-" "" ":" "" "." "" " " "" "+" "").Replace time.Now }}

{{- if .PrintApiUrl -}}
!Sub 'https://${ {{- $ApiName -}} }.execute-api.${AWS::Region}.amazonaws.com/{{ $StageName }}'
{{- else }}
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

    {{ $DeploymentName }}:
        Type: AWS::ApiGateway::Deployment
        DependsOn: {{ $ApiName }}
        Properties:
            RestApiId: !Ref {{ $ApiName }}
            StageName: {{ $StageName }}
{{- end }}
{{- end }}
`))
