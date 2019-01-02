{{ define "packages" }}
<style>
table, td, th {
    border: 1px solid #000;
    border-collapse: collapse;
}</style>

<h1>API Reference Documentation</h1>

{{ range .packages }}
    <h2>
        {{ packageDisplayName . }}
    </h2>

    {{ with .DocComments }}
    <p>
        {{ showComments . }}
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
        <h3 id="{{ .Name.Name }}">
            {{- .Name.Name }}
            {{ if eq .Kind "Alias" }}(<code>{{.Underlying}}</code> alias)</p>{{ end -}}
        </h3>
        {{ with (typeReferences .) }}
            <p>
                (<em>Appears on:</em>
                {{- $prev := "" -}}
                {{- range . -}}
                    {{- if $prev -}}, {{ end -}}
                    {{ $prev = . }}
                    <a href="#{{ typeIdentifier . }}">{{ typeDisplayName . }}</a>
                {{- end -}}
                )
            </p>
        {{ end }}

        {{ template "type" .  }}
    {{ end }}
    <hr/>
{{ end }}

{{ end }}
