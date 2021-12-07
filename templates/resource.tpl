{{ define "resource" }}

{{ template "type" .type }}

{{ range (visibleTypes (sortedTypes (childTypes .type)))}}
    {{ template "type" .  }}
{{ end }}

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>
    {{ with .gitCommit }} on git commit <code>{{ . }}</code>{{end}}.
</em></p>

{{ end }}
