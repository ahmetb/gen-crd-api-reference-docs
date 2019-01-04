{{ define "packages" }}
<style>
table, td, th {
    border: 1px solid #000;
    border-collapse: collapse;
}</style>

<h1>API Reference Documentation</h1>

{{ range .packages }}
    <h2>
        {{- packageDisplayName . -}}
    </h2>

    {{ with .DocComments }}
    <p>
        {{ safe (renderComments .) }}
    </p>
    {{ end }}

    Resource Types:
    <ul>
    {{- range (visibleTypes (sortedTypes .Types)) -}}
        {{ if isExportedType . -}}
        <li>
            <a href="#{{ typeIdentifier . }}">{{ typeIdentifier . }}</a>
        </li>
        {{- end }}
    {{- end -}}
    </ul>

    {{ range (visibleTypes (sortedTypes .Types))}}
        {{ template "type" .  }}
    {{ end }}
    <hr/>
{{ end }}

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>
    {{ with .gitCommit }} on git commit <code>{{ . }}</code>{{end}}.
</em></p>

{{ end }}
