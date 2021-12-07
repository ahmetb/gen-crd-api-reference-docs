{{ define "shared" }}

<h2 id="{{- packageAnchorID .package -}}">
    {{- packageDisplayName .package -}}
</h2>

{{ with (index .package.GoPackages 0 )}}
    {{ with .DocComments }}
    <div>
        {{ safe (renderComments .) }}
    </div>
    {{ end }}
{{ end }}

Resource Types:
<ul>
{{- range (visibleTypes (sortedTypes .package.Types)) -}}
    {{ if isExportedType . -}}
    <li>
        <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
    </li>
    {{- end }}
{{- end -}}
</ul>

All Types:
<ul>
{{- range (visibleTypes (sortedTypes .package.Types)) -}}
    <li>
        <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
    </li>
{{- end -}}
</ul>

{{ range (visibleTypes (sortedTypes (sharedTypes .package)))}}
    {{ template "type" .  }}
{{ end }}
<hr/>

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>
    {{ with .gitCommit }} on git commit <code>{{ . }}</code>{{end}}.
</em></p>

{{ end }}
