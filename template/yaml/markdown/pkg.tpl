{{ define "packages" }}

{{ range .packages }}

        {{ range (sortedTypes (visibleTypes .Types ))}}
            {{ if isObjectRoot . }}

            ##  {{ .Name.Name }}
            ### Description
            {{ (comments .CommentLines) }}
            ###  Type
            {{ (yamlType .) }}
            ###  Required
            {{ if .Members }}
               {{ range .Members }}
              {{ if (or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .)) }}
              {{ else }}
               * {{ (fieldName .) }}
               {{ end }}
               {{ end }}

               {{ template "properties" .  }}

               {{ template "members" (nodeParent . "") }}
            {{ end }}

            {{ end }}
        {{ end }}
{{ end }}

{{ end }}

