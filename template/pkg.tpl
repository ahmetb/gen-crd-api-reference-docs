{{ define "packages" }}
<style>
table, td, th {
    border: 1px solid #000;
    border-collapse: collapse;
}</style>

<h1>API Reference Documentation</h1>

{{ range .packages }}
    <h2>{{trimPackagePrefix .Path}}</h2>
    Types:
    {{ range (visibleTypes (sortedTypes .Types)) }}
    <ul>
        {{ if isExportedType . }}
        <ol>
            <a href="#{{localTypeIdentifier .}}">{{localTypeIdentifier .}}</a>
        </ol>
        {{ end }}
    </ul>
    {{ end }}

    {{ range (visibleTypes (sortedTypes .Types))}}
        <h3 id="{{ .Name.Name }}">{{ .Name.Name }}</h3>
        {{ if eq .Kind "Alias" }}<p>(This type is an alias to
            <code>{{.Underlying}}</code>.)</p>{{ end }}

        {{ with (typeReferences .) }}
            <p>
                <em>Appears on:</em>
                {{ range . }}
                    <a href="#{{localTypeIdentifier .}}">{{localTypeDisplayName .}}</a>
                {{ end }}
            </p>
        {{ end }}

        {{ template "type" .  }}
    {{ end }}
    <hr/>
{{ end }}

{{ end }}
