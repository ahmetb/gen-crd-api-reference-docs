{{ define "member" }}
 {{ if not (eq .Type.Kind "Builtin") }}

{{ if not (or (fieldEmbedded .Member) (hiddenMember .Member))}}
### {{ .Path }}
#### Description
{{ (comments .CommentLines) }}
####  Type
{{ (yamlType .Type) }}

{{ template "properties" .Type  }}

{{ template "members" (nodeParent .Type .Path) }}



{{ end }}
{{ end }}
{{ end }}

{{ define "members" }}
{{ $path := .Path }}
{{ if .Members }}

   {{ range .Members }}
       {{ if (or (eq .Name "ObjectMeta") (eq (fieldName .) "TypeMeta")) }}
        {{ else }}
         {{if (eq (yamlType .Type) "array") }}
             {{ $path := (printf "%s.%s[]" $path  (fieldName .)) }}
             ### {{ $path }}
             {{ template "type"  (nodeParent .Type.Elem $path ) }}
         {{ else }}
            {{ $path := (printf "%s.%s" $path  (fieldName .)) }}
            {{ template "member" (node . $path) }}
         {{ end }}
       {{ end }}
   {{ end }}

{{ else if .Elem }}
    {{ template "type" (nodeParent .Elem $path) }}
{{ else }}
{{end }}

{{end }}


{{ define "properties" }}
{{ if .Members }}
### Specification
|Property|Type|Description|
|---|---|---|
   {{ range .Members }}
       {{ if (or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .)) }}
       {{ else }}
           {{ if (isOptionalMember .) }}
           |{{ (fieldName .) }}|{{ (yamlType .Type)}}| {{ (comments .CommentLines "summary")}}|
           {{ else }}
           |{{ (fieldName .) }}|{{ (yamlType .Type)}}| (optional) {{ (comments .CommentLines "summary")}}|
           {{ end }}
       {{ end }}
   {{ end }}


{{ end }}
{{ end }}
