{{ define "type" }}

### Description
{{ (comments .CommentLines) }}
###  Type
{{ (yamlType .Type) }}

{{ if .Members }}

  {{ template "properties" .Type  }}
  {{ template "members" (nodeParent .Type .Path) }}

{{ end }}


{{ end }}
